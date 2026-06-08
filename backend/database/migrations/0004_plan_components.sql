-- +goose Up
-- Each component is a line-item table: typed columns for what we query, plus a
-- jsonb column for attributes that vary per row or get added later (expandability).
-- Money is always bigint cents (matches int64 *_cents in financial.proto).

CREATE TABLE plan_revenues (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id      uuid NOT NULL REFERENCES financial_plans(id) ON DELETE CASCADE,
    name         text NOT NULL,
    model        text NOT NULL CHECK (model IN ('subscription','one_time','usage','other')),
    amount_cents bigint NOT NULL,
    recurrence   text NOT NULL DEFAULT 'monthly' CHECK (recurrence IN ('monthly','annual','one_time')),
    start_month  int NOT NULL DEFAULT 0,
    assumptions  jsonb NOT NULL DEFAULT '{}'::jsonb,  -- growth %, churn %, seasonality
    created_at   timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX plan_revenues_plan_idx ON plan_revenues (plan_id);

CREATE TABLE plan_staffing (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id      uuid NOT NULL REFERENCES financial_plans(id) ON DELETE CASCADE,
    role_title   text NOT NULL,
    headcount    int NOT NULL DEFAULT 1,
    salary_cents bigint NOT NULL,            -- annual gross per head
    start_month  int NOT NULL DEFAULT 0,
    end_month    int,
    attributes   jsonb NOT NULL DEFAULT '{}'::jsonb,  -- benefits %, equity, department
    created_at   timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX plan_staffing_plan_idx ON plan_staffing (plan_id);

CREATE TABLE plan_expenses (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id      uuid NOT NULL REFERENCES financial_plans(id) ON DELETE CASCADE,
    category     text NOT NULL,              -- rent, marketing, saas tools, ...
    amount_cents bigint NOT NULL,
    recurrence   text NOT NULL DEFAULT 'monthly' CHECK (recurrence IN ('monthly','annual','one_time')),
    start_month  int NOT NULL DEFAULT 0,
    attributes   jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at   timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX plan_expenses_plan_idx ON plan_expenses (plan_id);

CREATE TABLE plan_assets (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id        uuid NOT NULL REFERENCES financial_plans(id) ON DELETE CASCADE,
    name           text NOT NULL,
    kind           text NOT NULL CHECK (kind IN ('cash','equipment','inventory','ip','other')),
    value_cents    bigint NOT NULL,
    acquired_month int NOT NULL DEFAULT 0,
    depreciation   jsonb NOT NULL DEFAULT '{}'::jsonb,  -- method, useful_life_months
    created_at     timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX plan_assets_plan_idx ON plan_assets (plan_id);

CREATE TABLE plan_funding (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id      uuid NOT NULL REFERENCES financial_plans(id) ON DELETE CASCADE,
    source       text NOT NULL CHECK (source IN ('equity','loan','grant','revenue','other')),
    label        text,                       -- 'Seed', 'Series A'
    amount_cents bigint NOT NULL,
    close_month  int NOT NULL DEFAULT 0,
    terms        jsonb NOT NULL DEFAULT '{}'::jsonb,  -- valuation, equity %, interest rate
    created_at   timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX plan_funding_plan_idx ON plan_funding (plan_id);

-- Forecasts are OUTPUTS from the Go financial-engine: the computed result plus
-- the input snapshot that produced it, so every number stays reproducible.
CREATE TABLE plan_forecasts (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id        uuid NOT NULL REFERENCES financial_plans(id) ON DELETE CASCADE,
    kind           text NOT NULL CHECK (kind IN ('runway','cashflow','scenario')),
    result         jsonb NOT NULL,           -- computed numbers from financial-engine
    input_snapshot jsonb NOT NULL,           -- the exact inputs used (audit/repro)
    generated_at   timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX plan_forecasts_plan_kind_idx ON plan_forecasts (plan_id, kind);

-- Financial decisions = the agent's recommendation plus its audit trail.
CREATE TABLE plan_decisions (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id        uuid NOT NULL REFERENCES financial_plans(id) ON DELETE CASCADE,
    created_by     uuid REFERENCES users(id) ON DELETE SET NULL,
    question       text NOT NULL,            -- user's natural-language ask
    recommendation text NOT NULL,            -- agent's answer
    rationale      text,
    tool_calls     jsonb NOT NULL DEFAULT '[]'::jsonb,  -- which engine calls/numbers backed it
    created_at     timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX plan_decisions_plan_idx ON plan_decisions (plan_id);

-- +goose Down
DROP TABLE IF EXISTS plan_decisions;
DROP TABLE IF EXISTS plan_forecasts;
DROP TABLE IF EXISTS plan_funding;
DROP TABLE IF EXISTS plan_assets;
DROP TABLE IF EXISTS plan_expenses;
DROP TABLE IF EXISTS plan_staffing;
DROP TABLE IF EXISTS plan_revenues;

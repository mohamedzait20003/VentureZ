-- +goose Up
CREATE TABLE businesses (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id  uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name       text NOT NULL,
    industry   text,
    country    text,
    currency   char(3) NOT NULL DEFAULT 'USD',
    attributes jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX businesses_client_idx ON businesses (client_id);

CREATE TABLE financial_plans (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    business_id    uuid NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    name           text NOT NULL,
    status         text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','active','archived')),
    currency       char(3) NOT NULL DEFAULT 'USD',
    start_date     date NOT NULL DEFAULT current_date,
    horizon_months int NOT NULL DEFAULT 24,
    assumptions    jsonb NOT NULL DEFAULT '{}'::jsonb,  -- global params: tax %, inflation, discount rate
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX financial_plans_business_idx ON financial_plans (business_id);

-- +goose Down
DROP TABLE IF EXISTS financial_plans;
DROP TABLE IF EXISTS businesses;

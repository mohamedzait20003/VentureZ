-- +goose Up
-- Tier definitions drive per-client limits (change limits without code changes).
CREATE TABLE plan_tiers (
    id                     uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code                   text UNIQUE NOT NULL,   -- 'free','pro','enterprise'
    name                   text NOT NULL,
    price_cents            bigint NOT NULL DEFAULT 0,
    billing_interval       text NOT NULL DEFAULT 'month' CHECK (billing_interval IN ('month','year')),
    max_businesses         int NOT NULL DEFAULT 1,
    max_plans_per_business int NOT NULL DEFAULT 1,
    features               jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at             timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE subscriptions (
    id                       uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id                uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tier_id                  uuid NOT NULL REFERENCES plan_tiers(id),
    status                   text NOT NULL CHECK (status IN ('trialing','active','past_due','canceled')),
    provider                 text NOT NULL DEFAULT 'stripe',
    provider_customer_id     text,
    provider_subscription_id text,
    current_period_start     timestamptz,
    current_period_end       timestamptz,
    cancel_at_period_end     boolean NOT NULL DEFAULT false,
    created_at               timestamptz NOT NULL DEFAULT now(),
    updated_at               timestamptz NOT NULL DEFAULT now()
);

-- At most one live subscription per client.
CREATE UNIQUE INDEX subscriptions_one_active_idx ON subscriptions (client_id)
    WHERE status IN ('trialing','active','past_due');

-- +goose Down
DROP TABLE IF EXISTS subscriptions;
DROP TABLE IF EXISTS plan_tiers;

-- Dev seed data. Run AFTER migrations:
--   psql "$DATABASE_URL" -f database/seeds/dev_seed.sql
-- Idempotent: safe to run repeatedly.
--
-- NOTE: the admin password is "Admin@1234" (bcrypt, cost 10) — DEV ONLY.
-- In any real environment, create users through the identity service instead.

-- ---- Subscription tiers (drive per-client limits) ----
INSERT INTO plan_tiers (id, code, name, price_cents, billing_interval, max_businesses, max_plans_per_business, features) VALUES
    ('a0000000-0000-0000-0000-000000000001', 'free',       'Free',       0,      'month', 1,  1,  '{"forecast_horizon_months": 12}'),
    ('a0000000-0000-0000-0000-000000000002', 'pro',        'Pro',        4900,   'month', 3,  5,  '{"forecast_horizon_months": 60, "agent_decisions": true}'),
    ('a0000000-0000-0000-0000-000000000003', 'enterprise', 'Enterprise', 29900,  'month', 50, 50, '{"forecast_horizon_months": 120, "agent_decisions": true, "priority_support": true}')
ON CONFLICT (code) DO NOTHING;

-- ---- Admin user (password: Admin@1234) ----
INSERT INTO users (id, email, password_hash, role, full_name, email_verified_at) VALUES
    ('b0000000-0000-0000-0000-000000000001', 'admin@venturez.dev', '$2a$10$VFY2HjJTAA8oVVU0ZsSImePHqcc9ecShXOCUzhVwIcmUKTRVxA2Nu', 'admin', 'VentureZ Admin', now())
ON CONFLICT (email) DO NOTHING;
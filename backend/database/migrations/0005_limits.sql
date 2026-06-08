-- +goose Up
-- "1-N businesses per client" and "1-M plans per business" can't be a declarative
-- CHECK (they count across rows) and the cap is tier-dependent. These triggers are
-- a hard backstop; the service layer should also check the limit for friendly errors.

-- +goose StatementBegin
CREATE FUNCTION enforce_business_limit() RETURNS trigger AS $$
DECLARE
    max_allowed   int;
    current_count int;
BEGIN
    SELECT t.max_businesses INTO max_allowed
    FROM subscriptions s
    JOIN plan_tiers t ON t.id = s.tier_id
    WHERE s.client_id = NEW.client_id
      AND s.status IN ('trialing','active','past_due')
    LIMIT 1;

    max_allowed := COALESCE(max_allowed, 1);  -- no subscription => free tier

    SELECT count(*) INTO current_count FROM businesses WHERE client_id = NEW.client_id;

    IF current_count >= max_allowed THEN
        RAISE EXCEPTION 'business limit (%) reached for client %', max_allowed, NEW.client_id
            USING ERRCODE = 'check_violation';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION enforce_plan_limit() RETURNS trigger AS $$
DECLARE
    owner_client  uuid;
    max_allowed   int;
    current_count int;
BEGIN
    SELECT client_id INTO owner_client FROM businesses WHERE id = NEW.business_id;

    SELECT t.max_plans_per_business INTO max_allowed
    FROM subscriptions s
    JOIN plan_tiers t ON t.id = s.tier_id
    WHERE s.client_id = owner_client
      AND s.status IN ('trialing','active','past_due')
    LIMIT 1;

    max_allowed := COALESCE(max_allowed, 1);

    SELECT count(*) INTO current_count FROM financial_plans WHERE business_id = NEW.business_id;

    IF current_count >= max_allowed THEN
        RAISE EXCEPTION 'plan limit (%) reached for business %', max_allowed, NEW.business_id
            USING ERRCODE = 'check_violation';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER trg_business_limit
    BEFORE INSERT ON businesses
    FOR EACH ROW EXECUTE FUNCTION enforce_business_limit();

CREATE TRIGGER trg_plan_limit
    BEFORE INSERT ON financial_plans
    FOR EACH ROW EXECUTE FUNCTION enforce_plan_limit();

-- +goose Down
DROP TRIGGER IF EXISTS trg_plan_limit ON financial_plans;
DROP TRIGGER IF EXISTS trg_business_limit ON businesses;
DROP FUNCTION IF EXISTS enforce_plan_limit();
DROP FUNCTION IF EXISTS enforce_business_limit();

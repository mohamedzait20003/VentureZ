-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE users (
    id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email             citext UNIQUE NOT NULL,
    username          text UNIQUE NOT NULL,
    first_name        text,
    last_name         text,
    password_hash     text NOT NULL,
    role              text NOT NULL DEFAULT 'client' CHECK (role IN ('client','admin')),
    emailed_verified  boolean NOT NULL DEFAULT false,
    email_verified_at timestamptz,
    created_at        timestamptz NOT NULL DEFAULT now(),
    updated_at        timestamptz NOT NULL DEFAULT now()
);

-- One reset + one verify per user, enforced by the unique (user_id, type) pair.
CREATE TABLE user_tokens (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type        text NOT NULL CHECK (type IN ('password_reset','email_verify')),
    token_hash  text NOT NULL,              -- store a hash, never the raw token
    expires_at  timestamptz NOT NULL,
    consumed_at timestamptz,
    created_at  timestamptz NOT NULL DEFAULT now(),
    UNIQUE (user_id, type)
);

-- Refresh-token sessions: one row per device.
CREATE TABLE sessions (
    id                 uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id            uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token_hash text NOT NULL UNIQUE,  -- hashed; raw token lives only on the client
    user_agent         text,
    ip_address         inet,
    device_name        text,
    created_at         timestamptz NOT NULL DEFAULT now(),
    last_used_at       timestamptz NOT NULL DEFAULT now(),
    expires_at         timestamptz NOT NULL,
    revoked_at         timestamptz
);
CREATE INDEX sessions_active_idx ON sessions (user_id) WHERE revoked_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS user_tokens;
DROP TABLE IF EXISTS users;

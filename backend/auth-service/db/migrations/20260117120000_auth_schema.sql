-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS auth;
SET search_path TO auth, public;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE auth_accounts (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email TEXT NOT NULL,
  email_normalized TEXT NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('new','pending_verification','active','blocked','deleted')),
  role TEXT NOT NULL DEFAULT 'user' CHECK (role IN ('user','admin','system')),
  email_verified_at TIMESTAMPTZ,
  last_login_at TIMESTAMPTZ,
  blocked_at TIMESTAMPTZ,
  blocked_reason TEXT,
  deleted_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX ux_accounts_email_norm ON auth_accounts(email_normalized) WHERE deleted_at IS NULL;
CREATE INDEX ix_accounts_status ON auth_accounts(status);

CREATE TABLE auth_credentials (
  user_id UUID PRIMARY KEY REFERENCES auth_accounts(id) ON DELETE CASCADE,
  password_hash TEXT NOT NULL,
  algo_version SMALLINT NOT NULL DEFAULT 1,
  password_changed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  failed_login_count INT NOT NULL DEFAULT 0,
  last_failed_login_at TIMESTAMPTZ
);

CREATE TABLE auth_email_verifications (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES auth_accounts(id) ON DELETE CASCADE,
  email TEXT NOT NULL,
  code_salt BYTEA NOT NULL,
  code_hash TEXT NOT NULL,
  expires_at TIMESTAMPTZ NOT NULL,
  attempts_count INT NOT NULL DEFAULT 0,
  max_attempts INT NOT NULL DEFAULT 5,
  status TEXT NOT NULL CHECK (status IN ('pending','verified','expired','cancelled')),
  verified_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX ix_verif_user_status ON auth_email_verifications(user_id, status) WHERE status = 'pending';

CREATE TABLE auth_password_resets (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES auth_accounts(id) ON DELETE CASCADE,
  email TEXT NOT NULL,
  code_salt BYTEA NOT NULL,
  code_hash TEXT NOT NULL,
  token_type TEXT NOT NULL DEFAULT 'otp_code',
  requested_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  expires_at TIMESTAMPTZ NOT NULL,
  attempts_count INT NOT NULL DEFAULT 0,
  status TEXT NOT NULL CHECK (status IN ('pending','used','expired','cancelled')),
  used_at TIMESTAMPTZ
);

CREATE INDEX ix_reset_user_status ON auth_password_resets(user_id, status) WHERE status = 'pending';

CREATE TABLE auth_sessions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES auth_accounts(id) ON DELETE CASCADE,
  refresh_token_hash TEXT NOT NULL,
  session_family_id UUID NOT NULL,
  replaced_by_session_id UUID REFERENCES auth_sessions(id),
  device_id TEXT,
  device_name TEXT,
  platform TEXT CHECK (platform IS NULL OR platform IN ('ios','android','web','admin')),
  ip INET,
  user_agent TEXT,
  issued_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  expires_at TIMESTAMPTZ NOT NULL,
  last_seen_at TIMESTAMPTZ,
  revoked_at TIMESTAMPTZ,
  revoke_reason TEXT CHECK (revoke_reason IS NULL OR revoke_reason IN ('logout','logout_all','password_reset','password_changed','admin_force_logout','refresh_reuse_detected','expired','rotation')),
  compromised_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX ux_sessions_refresh ON auth_sessions(refresh_token_hash);
CREATE INDEX ix_sessions_user_active ON auth_sessions(user_id) WHERE revoked_at IS NULL;
CREATE INDEX ix_sessions_family ON auth_sessions(session_family_id);

CREATE TABLE auth_audit_log (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES auth_accounts(id) ON DELETE SET NULL,
  event_type TEXT NOT NULL,
  ip INET,
  user_agent TEXT,
  meta JSONB NOT NULL DEFAULT '{}',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX ix_audit_user_time ON auth_audit_log(user_id, created_at DESC);

CREATE TABLE auth_outbox (
  id BIGSERIAL PRIMARY KEY,
  aggregate_type TEXT NOT NULL,
  event_type TEXT NOT NULL,
  payload JSONB NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  published_at TIMESTAMPTZ
);

CREATE INDEX ix_outbox_unpublished ON auth_outbox(published_at) WHERE published_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS auth_outbox;
DROP TABLE IF EXISTS auth_audit_log;
DROP TABLE IF EXISTS auth_sessions;
DROP TABLE IF EXISTS auth_password_resets;
DROP TABLE IF EXISTS auth_email_verifications;
DROP TABLE IF EXISTS auth_credentials;
DROP TABLE IF EXISTS auth_accounts;
-- +goose StatementEnd

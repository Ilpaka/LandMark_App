-- +goose Up
-- +goose StatementBegin
-- Phone-based auth: nullable email, unique phone, OTP table.

DROP INDEX IF EXISTS ux_accounts_email_norm;

ALTER TABLE auth_accounts
  ALTER COLUMN email DROP NOT NULL,
  ALTER COLUMN email_normalized DROP NOT NULL;

UPDATE auth_accounts SET email = NULL, email_normalized = NULL WHERE email IS NOT NULL AND trim(email) = '';
UPDATE auth_accounts SET email = NULL, email_normalized = NULL WHERE email_normalized IS NOT NULL AND trim(email_normalized) = '';

CREATE UNIQUE INDEX ux_accounts_email_norm
  ON auth_accounts(email_normalized)
  WHERE deleted_at IS NULL AND email_normalized IS NOT NULL;

ALTER TABLE auth_accounts
  ADD COLUMN IF NOT EXISTS phone_e164 TEXT,
  ADD COLUMN IF NOT EXISTS phone_verified_at TIMESTAMPTZ;

CREATE UNIQUE INDEX ux_accounts_phone_e164
  ON auth_accounts(phone_e164)
  WHERE deleted_at IS NULL AND phone_e164 IS NOT NULL;

CREATE TABLE auth_phone_otp (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES auth_accounts(id) ON DELETE CASCADE,
  phone_e164 TEXT NOT NULL,
  code_salt BYTEA NOT NULL,
  code_hash TEXT NOT NULL,
  expires_at TIMESTAMPTZ NOT NULL,
  attempts_count INT NOT NULL DEFAULT 0,
  max_attempts INT NOT NULL DEFAULT 5,
  status TEXT NOT NULL CHECK (status IN ('pending','verified','expired','cancelled')),
  verified_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX ix_phone_otp_user_status ON auth_phone_otp(user_id, status) WHERE status = 'pending';
CREATE INDEX ix_phone_otp_phone_pending ON auth_phone_otp(phone_e164) WHERE status = 'pending';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS auth_phone_otp;

DROP INDEX IF EXISTS ix_accounts_phone;
DROP INDEX IF EXISTS ux_accounts_phone_e164;
ALTER TABLE auth_accounts DROP COLUMN IF EXISTS phone_verified_at;
ALTER TABLE auth_accounts DROP COLUMN IF EXISTS phone_e164;

DROP INDEX IF EXISTS ux_accounts_email_norm;
ALTER TABLE auth_accounts
  ALTER COLUMN email SET NOT NULL,
  ALTER COLUMN email_normalized SET NOT NULL;
CREATE UNIQUE INDEX ux_accounts_email_norm ON auth_accounts(email_normalized) WHERE deleted_at IS NULL;
-- +goose StatementEnd

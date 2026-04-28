-- +goose Up
-- +goose StatementBegin
DROP TABLE IF EXISTS auth_phone_otp;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE TABLE auth_phone_otp (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES auth_accounts(id) ON DELETE CASCADE,
  phone_e164 TEXT NOT NULL,
  purpose TEXT NOT NULL DEFAULT 'login'
    CHECK (purpose IN ('login', 'password_reset')),
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

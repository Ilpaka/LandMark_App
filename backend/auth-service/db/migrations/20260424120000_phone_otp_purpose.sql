-- +goose Up
-- +goose StatementBegin
ALTER TABLE auth_phone_otp
  ADD COLUMN IF NOT EXISTS purpose TEXT NOT NULL DEFAULT 'login'
    CHECK (purpose IN ('login', 'password_reset'));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE auth_phone_otp DROP COLUMN IF EXISTS purpose;
-- +goose StatementEnd

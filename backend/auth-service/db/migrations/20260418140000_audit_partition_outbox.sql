-- +goose Up
-- +goose StatementBegin
-- Partitioned audit log (monthly) + default catch-all; PK includes created_at (PostgreSQL requirement).

ALTER TABLE IF EXISTS auth_audit_log RENAME TO auth_audit_log_legacy;

CREATE TABLE auth_audit_log (
  id UUID NOT NULL,
  user_id UUID REFERENCES auth_accounts(id) ON DELETE SET NULL,
  event_type TEXT NOT NULL,
  ip INET,
  user_agent TEXT,
  meta JSONB NOT NULL DEFAULT '{}',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

DO $$
DECLARE
  d date := '2026-01-01';
  end_m date := '2030-01-01';
BEGIN
  WHILE d < end_m LOOP
    EXECUTE format(
      'CREATE TABLE %I PARTITION OF auth_audit_log FOR VALUES FROM (%L) TO (%L)',
      'auth_audit_log_' || to_char(d, 'YYYY_MM'),
      d,
      d + interval '1 month'
    );
    d := d + interval '1 month';
  END LOOP;
END $$;

CREATE TABLE auth_audit_log_default PARTITION OF auth_audit_log DEFAULT;

INSERT INTO auth_audit_log (id, user_id, event_type, ip, user_agent, meta, created_at)
SELECT id, user_id, event_type, ip, user_agent, meta, created_at FROM auth_audit_log_legacy;

DROP TABLE auth_audit_log_legacy;

CREATE INDEX ix_audit_user_time ON auth_audit_log(user_id, created_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS auth_audit_log CASCADE;

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
-- +goose StatementEnd

-- Migration: Add org_audit_logs table
-- Date: 2026-05-11
-- Description: Append-only audit trail for security-relevant events within organizations.

CREATE TABLE IF NOT EXISTS org_audit_logs (
    id          BIGSERIAL PRIMARY KEY,
    org_id      BIGINT      NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    actor_id    TEXT        NOT NULL,
    action      TEXT        NOT NULL,
    target_id   TEXT        NOT NULL DEFAULT '',
    target_type TEXT        NOT NULL DEFAULT '',
    detail      TEXT        NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_org_audit_logs_org_created
    ON org_audit_logs (org_id, created_at DESC);

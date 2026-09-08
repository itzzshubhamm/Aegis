-- Migration 000002: Add metadata column and performance indexes for alerts

ALTER TABLE alerts ADD COLUMN IF NOT EXISTS metadata JSONB NOT NULL DEFAULT '{}'::jsonb;

CREATE INDEX IF NOT EXISTS idx_alerts_tenant_status ON alerts(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_alerts_tenant_detection_type ON alerts(tenant_id, detection_type);
CREATE INDEX IF NOT EXISTS idx_alerts_tenant_timestamp ON alerts(tenant_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_honeytokens_tenant_id ON honeytokens(tenant_id);

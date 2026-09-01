-- Tenants
-- name: CreateTenant :one
INSERT INTO tenants (name)
VALUES ($1)
RETURNING *;

-- name: GetTenantByID :one
SELECT * FROM tenants
WHERE id = $1;

-- name: ListTenants :many
SELECT * FROM tenants
ORDER BY created_at DESC;

-- Users
-- name: CreateUser :one
INSERT INTO users (tenant_id, email, password_hash, role)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1;

-- Security Events
-- name: CreateSecurityEvent :one
INSERT INTO security_events (id, tenant_id, event_type, source, source_ip, user_id, resource, action, severity, status, metadata, timestamp)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING *;

-- name: GetSecurityEventByID :one
SELECT * FROM security_events
WHERE id = $1 AND tenant_id = $2;

-- name: ListRecentSecurityEvents :many
SELECT * FROM security_events
WHERE tenant_id = $1
ORDER BY timestamp DESC
LIMIT $2;

-- name: CountSecurityEvents :one
SELECT COUNT(*) FROM security_events
WHERE tenant_id = $1;

-- name: CountFailedLoginsInWindow :one
SELECT COUNT(*) FROM security_events
WHERE tenant_id = $1
  AND event_type = 'LOGIN_FAILED'
  AND (source_ip = $2 OR user_id = $3)
  AND timestamp >= $4;

-- Alerts
-- name: CreateAlert :one
INSERT INTO alerts (tenant_id, severity, detection_type, source_ip, affected_asset, description, status, timestamp)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: ListRecentAlerts :many
SELECT * FROM alerts
WHERE tenant_id = $1
ORDER BY created_at DESC
LIMIT $2;

-- name: CountAlerts :one
SELECT COUNT(*) FROM alerts
WHERE tenant_id = $1;

-- name: CountAlertsBySeverity :many
SELECT severity, COUNT(*) as count FROM alerts
WHERE tenant_id = $1
GROUP BY severity;

-- name: UpdateAlertStatus :one
UPDATE alerts
SET status = $2
WHERE id = $1 AND tenant_id = $3
RETURNING *;

-- Honeytokens
-- name: CreateHoneytoken :one
INSERT INTO honeytokens (tenant_id, name, type, token_value, status)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetHoneytokenByValue :one
SELECT * FROM honeytokens
WHERE token_value = $1;

-- name: ListHoneytokens :many
SELECT * FROM honeytokens
WHERE tenant_id = $1
ORDER BY created_at DESC;

-- name: UpdateHoneytokenStatus :one
UPDATE honeytokens
SET status = $2, last_triggered_at = $3
WHERE id = $1 AND tenant_id = $4
RETURNING *;

-- No tables needed for notification service (stateless - just sends emails)
-- Could add a notifications_log table for audit trail

CREATE TABLE IF NOT EXISTS notification_logs (
    id VARCHAR(36) PRIMARY KEY,
    recipient VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL, -- email, sms
    subject VARCHAR(255),
    body TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'sent',
    error TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

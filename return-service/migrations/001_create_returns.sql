CREATE TABLE IF NOT EXISTS return_requests (
    id VARCHAR(36) PRIMARY KEY,
    shipment_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    reason TEXT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'requested',
    carrier VARCHAR(50),
    return_label_id VARCHAR(36),
    refund_amount DECIMAL(10,2) DEFAULT 0,
    refund_status VARCHAR(50) DEFAULT 'pending',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

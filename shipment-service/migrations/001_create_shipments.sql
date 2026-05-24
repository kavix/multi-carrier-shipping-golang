CREATE TABLE IF NOT EXISTS shipments (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    sender_name VARCHAR(255) NOT NULL,
    sender_address TEXT NOT NULL,
    sender_email VARCHAR(255),
    receiver_name VARCHAR(255) NOT NULL,
    receiver_address TEXT NOT NULL,
    receiver_email VARCHAR(255),
    weight DECIMAL(10,2) NOT NULL,
    dimensions VARCHAR(50),
    carrier VARCHAR(50) NOT NULL,
    service_type VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    tracking_number VARCHAR(100),
    cost DECIMAL(10,2) DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

ALTER TABLE shipments ADD COLUMN IF NOT EXISTS sender_email VARCHAR(255);
ALTER TABLE shipments ADD COLUMN IF NOT EXISTS receiver_email VARCHAR(255);


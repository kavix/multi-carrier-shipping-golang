CREATE TABLE IF NOT EXISTS labels (
    id VARCHAR(36) PRIMARY KEY,
    shipment_id VARCHAR(36) NOT NULL UNIQUE,
    carrier VARCHAR(50) NOT NULL,
    tracking_number VARCHAR(100) NOT NULL,
    label_data TEXT NOT NULL,
    label_url VARCHAR(255) NOT NULL,
    format VARCHAR(10) NOT NULL DEFAULT 'PDF',
    status VARCHAR(50) NOT NULL DEFAULT 'generated',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

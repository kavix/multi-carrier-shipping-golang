CREATE TABLE IF NOT EXISTS rate_comparisons (
    id VARCHAR(36) PRIMARY KEY,
    shipment_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    from_address TEXT NOT NULL,
    to_address TEXT NOT NULL,
    weight DECIMAL(10,2) NOT NULL,
    best_carrier VARCHAR(50) NOT NULL,
    best_service VARCHAR(100) NOT NULL,
    best_cost DECIMAL(10,2) NOT NULL,
    best_days INT NOT NULL,
    all_rates_json TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

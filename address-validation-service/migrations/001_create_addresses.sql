CREATE TABLE IF NOT EXISTS validated_addresses (
    id VARCHAR(36) PRIMARY KEY,
    raw_address TEXT UNIQUE NOT NULL,
    street VARCHAR(255),
    city VARCHAR(100),
    state VARCHAR(100),
    postal_code VARCHAR(20),
    country VARCHAR(50),
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    is_valid BOOLEAN NOT NULL DEFAULT false,
    validated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS carriers (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    code VARCHAR(20) UNIQUE NOT NULL,
    api_key VARCHAR(255) NOT NULL,
    api_secret VARCHAR(255) NOT NULL,
    base_url VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Seed with demo carriers (replace with real API credentials)
INSERT INTO carriers (id, name, code, api_key, api_secret, base_url, is_active) VALUES
('carrier-dhl', 'DHL Express', 'dhl', 'dhl-demo-key', 'dhl-demo-secret', 'https://api.dhl.com', true),
('carrier-fedex', 'FedEx', 'fedex', 'l7d79cfde66a33429990b1e760138fb300', 'f43a6307ecc14f57b20b373186c6f3a9', 'https://apis-sandbox.fedex.com', true),
('carrier-ups', 'UPS', 'ups', 'ups-demo-key', 'ups-demo-secret', 'https://onlinetools.ups.com', true)
ON CONFLICT (code) DO UPDATE
SET api_key = EXCLUDED.api_key,
    api_secret = EXCLUDED.api_secret,
    base_url = EXCLUDED.base_url;

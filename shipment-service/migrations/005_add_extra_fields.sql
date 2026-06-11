-- Migration to add extra fields for FedEx and generic shipment description
ALTER TABLE shipments ADD COLUMN IF NOT EXISTS description TEXT;
ALTER TABLE shipments ADD COLUMN IF NOT EXISTS is_international BOOLEAN DEFAULT FALSE;
ALTER TABLE shipments ADD COLUMN IF NOT EXISTS customs_value DECIMAL(10,2) DEFAULT 0;
ALTER TABLE shipments ADD COLUMN IF NOT EXISTS customs_currency VARCHAR(10) DEFAULT 'USD';

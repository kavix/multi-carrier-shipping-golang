-- Migration to add pickup and drop location IDs to shipments
ALTER TABLE shipments ADD COLUMN IF NOT EXISTS pickup_location_id VARCHAR(255);
ALTER TABLE shipments ADD COLUMN IF NOT EXISTS drop_location_id VARCHAR(255);

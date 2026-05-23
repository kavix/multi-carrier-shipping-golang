-- Create database for auth service
SELECT 'CREATE DATABASE auth_db'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'auth_db')\gexec

-- Create database for shipment service
SELECT 'CREATE DATABASE shipment_db'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'shipment_db')\gexec

-- Create database for label service
SELECT 'CREATE DATABASE label_db'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'label_db')\gexec

-- Create database for notification service
SELECT 'CREATE DATABASE notification_db'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'notification_db')\gexec

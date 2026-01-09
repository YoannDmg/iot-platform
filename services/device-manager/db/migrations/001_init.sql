-- IoT Platform - Device Manager Schema
-- Initial migration: Create devices table with metadata support

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Device status enum
CREATE TYPE device_status AS ENUM ('UNKNOWN', 'ONLINE', 'OFFLINE', 'ERROR', 'MAINTENANCE');

-- Main devices table
CREATE TABLE devices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    type VARCHAR(100) NOT NULL,
    status device_status NOT NULL DEFAULT 'ONLINE',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'::jsonb,

    CONSTRAINT name_not_empty CHECK (name <> ''),
    CONSTRAINT type_not_empty CHECK (type <> '')
);

-- Indexes for performance
CREATE INDEX idx_devices_type ON devices(type);
CREATE INDEX idx_devices_status ON devices(status);
CREATE INDEX idx_devices_created_at ON devices(created_at DESC);
CREATE INDEX idx_devices_last_seen ON devices(last_seen DESC);

-- GIN index for JSONB metadata queries
CREATE INDEX idx_devices_metadata ON devices USING GIN (metadata);

-- Comments for documentation
COMMENT ON TABLE devices IS 'IoT devices managed by the platform';
COMMENT ON COLUMN devices.id IS 'Unique device identifier (UUID)';
COMMENT ON COLUMN devices.name IS 'Human-readable device name';
COMMENT ON COLUMN devices.type IS 'Device type (sensor, actuator, gateway, etc.)';
COMMENT ON COLUMN devices.status IS 'Current operational status';
COMMENT ON COLUMN devices.created_at IS 'Device registration timestamp';
COMMENT ON COLUMN devices.last_seen IS 'Last activity timestamp';
COMMENT ON COLUMN devices.metadata IS 'Flexible key-value storage for device-specific data';

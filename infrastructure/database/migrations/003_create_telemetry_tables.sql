-- Migration: Create telemetry tables with TimescaleDB
-- Description: Time-series storage for IoT device telemetry data

-- Enable TimescaleDB extension
CREATE EXTENSION IF NOT EXISTS timescaledb;

-- ============================================
-- TELEMETRY DATA TABLE (Hypertable)
-- ============================================

-- Main telemetry table for raw sensor data
CREATE TABLE device_telemetry (
    time        TIMESTAMPTZ NOT NULL,
    device_id   UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    metric_name VARCHAR(100) NOT NULL,
    value       DOUBLE PRECISION NOT NULL,
    unit        VARCHAR(50),
    metadata    JSONB DEFAULT '{}'::jsonb,

    -- Composite primary key for efficient time-series queries
    PRIMARY KEY (device_id, metric_name, time)
);

-- Convert to TimescaleDB hypertable (automatic time-based partitioning)
-- Chunks of 1 day for optimal query performance
SELECT create_hypertable(
    'device_telemetry',
    'time',
    chunk_time_interval => INTERVAL '1 day',
    if_not_exists => TRUE
);

-- ============================================
-- INDEXES
-- ============================================

-- Fast lookups by device and time range
CREATE INDEX idx_telemetry_device_time
    ON device_telemetry(device_id, time DESC);

-- Fast lookups by metric across all devices
CREATE INDEX idx_telemetry_metric_time
    ON device_telemetry(metric_name, time DESC);

-- Combined index for common query pattern
CREATE INDEX idx_telemetry_device_metric_time
    ON device_telemetry(device_id, metric_name, time DESC);

-- ============================================
-- DATA RETENTION POLICY
-- ============================================

-- Automatically drop data older than 90 days
SELECT add_retention_policy(
    'device_telemetry',
    INTERVAL '90 days',
    if_not_exists => TRUE
);

-- ============================================
-- CONTINUOUS AGGREGATES (Materialized Views)
-- ============================================

-- Hourly aggregations for dashboards and reports
CREATE MATERIALIZED VIEW telemetry_hourly
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 hour', time) AS bucket,
    device_id,
    metric_name,
    AVG(value) AS avg_value,
    MIN(value) AS min_value,
    MAX(value) AS max_value,
    COUNT(*) AS sample_count,
    FIRST(value, time) AS first_value,
    LAST(value, time) AS last_value
FROM device_telemetry
GROUP BY bucket, device_id, metric_name
WITH NO DATA;

-- Auto-refresh hourly aggregates
SELECT add_continuous_aggregate_policy('telemetry_hourly',
    start_offset => INTERVAL '3 hours',
    end_offset => INTERVAL '1 hour',
    schedule_interval => INTERVAL '1 hour',
    if_not_exists => TRUE
);

-- Daily aggregations for long-term trends
CREATE MATERIALIZED VIEW telemetry_daily
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 day', time) AS bucket,
    device_id,
    metric_name,
    AVG(value) AS avg_value,
    MIN(value) AS min_value,
    MAX(value) AS max_value,
    COUNT(*) AS sample_count
FROM device_telemetry
GROUP BY bucket, device_id, metric_name
WITH NO DATA;

-- Auto-refresh daily aggregates
SELECT add_continuous_aggregate_policy('telemetry_daily',
    start_offset => INTERVAL '3 days',
    end_offset => INTERVAL '1 day',
    schedule_interval => INTERVAL '1 day',
    if_not_exists => TRUE
);

-- ============================================
-- HELPER TABLE: Latest values cache
-- ============================================

-- Stores the latest value for each device/metric pair
-- Updated via trigger for fast "current state" queries
CREATE TABLE device_telemetry_latest (
    device_id   UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    metric_name VARCHAR(100) NOT NULL,
    time        TIMESTAMPTZ NOT NULL,
    value       DOUBLE PRECISION NOT NULL,
    unit        VARCHAR(50),

    PRIMARY KEY (device_id, metric_name)
);

-- Function to update latest values
CREATE OR REPLACE FUNCTION update_telemetry_latest()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO device_telemetry_latest (device_id, metric_name, time, value, unit)
    VALUES (NEW.device_id, NEW.metric_name, NEW.time, NEW.value, NEW.unit)
    ON CONFLICT (device_id, metric_name)
    DO UPDATE SET
        time = EXCLUDED.time,
        value = EXCLUDED.value,
        unit = EXCLUDED.unit
    WHERE EXCLUDED.time > device_telemetry_latest.time;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to auto-update latest values on insert
CREATE TRIGGER trg_update_telemetry_latest
    AFTER INSERT ON device_telemetry
    FOR EACH ROW
    EXECUTE FUNCTION update_telemetry_latest();

-- ============================================
-- COMMENTS
-- ============================================

COMMENT ON TABLE device_telemetry IS 'Raw telemetry data from IoT devices (TimescaleDB hypertable)';
COMMENT ON COLUMN device_telemetry.time IS 'Timestamp when the measurement was taken';
COMMENT ON COLUMN device_telemetry.device_id IS 'Reference to the device that sent the data';
COMMENT ON COLUMN device_telemetry.metric_name IS 'Name of the metric (e.g., temperature, humidity, pressure)';
COMMENT ON COLUMN device_telemetry.value IS 'Numeric value of the measurement';
COMMENT ON COLUMN device_telemetry.unit IS 'Unit of measurement (e.g., °C, %, hPa)';
COMMENT ON COLUMN device_telemetry.metadata IS 'Additional context as JSON (e.g., sensor_id, location)';

COMMENT ON MATERIALIZED VIEW telemetry_hourly IS 'Hourly aggregations of telemetry data (auto-refreshed)';
COMMENT ON MATERIALIZED VIEW telemetry_daily IS 'Daily aggregations of telemetry data (auto-refreshed)';
COMMENT ON TABLE device_telemetry_latest IS 'Cache of latest values per device/metric for fast lookups';

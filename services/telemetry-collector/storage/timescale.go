package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	pb "github.com/yourusername/iot-platform/shared/proto/telemetry"
)

// TimescaleStorage implements Storage using TimescaleDB.
type TimescaleStorage struct {
	pool *pgxpool.Pool
}

// NewTimescaleStorage creates a new TimescaleDB storage connection.
func NewTimescaleStorage(ctx context.Context, dsn string) (*TimescaleStorage, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DSN: %w", err)
	}

	// Connection pool settings
	config.MaxConns = 20
	config.MinConns = 5
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &TimescaleStorage{pool: pool}, nil
}

// InsertTelemetry inserts a single telemetry point.
func (s *TimescaleStorage) InsertTelemetry(ctx context.Context, deviceID, metricName string, value float64, unit string, timestamp int64, metadata map[string]string) error {
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		metadataJSON = []byte("{}")
	}

	ts := time.Unix(timestamp, 0)

	_, err = s.pool.Exec(ctx, `
		INSERT INTO device_telemetry (time, device_id, metric_name, value, unit, metadata)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, ts, deviceID, metricName, value, unit, metadataJSON)

	if err != nil {
		return fmt.Errorf("failed to insert telemetry: %w", err)
	}

	return nil
}

// InsertTelemetryBatch inserts multiple telemetry points in a single transaction.
func (s *TimescaleStorage) InsertTelemetryBatch(ctx context.Context, points []*TelemetryPoint) error {
	if len(points) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	for _, point := range points {
		metadataJSON, err := json.Marshal(point.Metadata)
		if err != nil {
			metadataJSON = []byte("{}")
		}

		ts := time.Unix(point.Timestamp, 0)
		batch.Queue(`
			INSERT INTO device_telemetry (time, device_id, metric_name, value, unit, metadata)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, ts, point.DeviceID, point.MetricName, point.Value, point.Unit, metadataJSON)
	}

	br := s.pool.SendBatch(ctx, batch)
	defer br.Close()

	for range points {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("failed to execute batch insert: %w", err)
		}
	}

	return nil
}

// GetTelemetry retrieves telemetry data for a device within a time range.
func (s *TimescaleStorage) GetTelemetry(ctx context.Context, deviceID, metricName string, fromTime, toTime int64, limit int) ([]*pb.TelemetryPoint, error) {
	if limit <= 0 || limit > 10000 {
		limit = 1000
	}

	fromTS := time.Unix(fromTime, 0)
	toTS := time.Unix(toTime, 0)

	rows, err := s.pool.Query(ctx, `
		SELECT time, value, unit
		FROM device_telemetry
		WHERE device_id = $1
		  AND metric_name = $2
		  AND time >= $3
		  AND time <= $4
		ORDER BY time DESC
		LIMIT $5
	`, deviceID, metricName, fromTS, toTS, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query telemetry: %w", err)
	}
	defer rows.Close()

	var points []*pb.TelemetryPoint
	for rows.Next() {
		var ts time.Time
		var value float64
		var unit *string

		if err := rows.Scan(&ts, &value, &unit); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		point := &pb.TelemetryPoint{
			Time:  ts.Unix(),
			Value: value,
		}
		if unit != nil {
			point.Unit = *unit
		}

		points = append(points, point)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return points, nil
}

// GetTelemetryAggregated retrieves aggregated telemetry data.
func (s *TimescaleStorage) GetTelemetryAggregated(ctx context.Context, deviceID, metricName string, fromTime, toTime int64, interval string) ([]*pb.TelemetryAggregation, error) {
	fromTS := time.Unix(fromTime, 0)
	toTS := time.Unix(toTime, 0)

	// Validate interval to prevent SQL injection
	validIntervals := map[string]bool{
		"1 minute":  true,
		"5 minutes": true,
		"15 minutes": true,
		"30 minutes": true,
		"1 hour":    true,
		"6 hours":   true,
		"12 hours":  true,
		"1 day":     true,
		"1 week":    true,
	}

	if !validIntervals[interval] {
		interval = "1 hour" // Default to 1 hour if invalid
	}

	query := fmt.Sprintf(`
		SELECT
			time_bucket('%s', time) AS bucket,
			AVG(value) AS avg_value,
			MIN(value) AS min_value,
			MAX(value) AS max_value,
			COUNT(*) AS sample_count
		FROM device_telemetry
		WHERE device_id = $1
		  AND metric_name = $2
		  AND time >= $3
		  AND time <= $4
		GROUP BY bucket
		ORDER BY bucket DESC
	`, interval)

	rows, err := s.pool.Query(ctx, query, deviceID, metricName, fromTS, toTS)
	if err != nil {
		return nil, fmt.Errorf("failed to query aggregated telemetry: %w", err)
	}
	defer rows.Close()

	var aggregations []*pb.TelemetryAggregation
	for rows.Next() {
		var bucket time.Time
		var avg, min, max float64
		var count int64

		if err := rows.Scan(&bucket, &avg, &min, &max, &count); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		aggregations = append(aggregations, &pb.TelemetryAggregation{
			Bucket: bucket.Format(time.RFC3339),
			Avg:    avg,
			Min:    min,
			Max:    max,
			Count:  count,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return aggregations, nil
}

// GetLatestMetric retrieves the latest value for a specific metric.
func (s *TimescaleStorage) GetLatestMetric(ctx context.Context, deviceID, metricName string) (*pb.TelemetryPoint, error) {
	// Try to get from the latest cache table first
	var ts time.Time
	var value float64
	var unit *string

	err := s.pool.QueryRow(ctx, `
		SELECT time, value, unit
		FROM device_telemetry_latest
		WHERE device_id = $1 AND metric_name = $2
	`, deviceID, metricName).Scan(&ts, &value, &unit)

	if err == pgx.ErrNoRows {
		// Fallback to main table if not in cache
		err = s.pool.QueryRow(ctx, `
			SELECT time, value, unit
			FROM device_telemetry
			WHERE device_id = $1 AND metric_name = $2
			ORDER BY time DESC
			LIMIT 1
		`, deviceID, metricName).Scan(&ts, &value, &unit)
	}

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("no telemetry found for device %s metric %s", deviceID, metricName)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query latest metric: %w", err)
	}

	point := &pb.TelemetryPoint{
		Time:  ts.Unix(),
		Value: value,
	}
	if unit != nil {
		point.Unit = *unit
	}

	return point, nil
}

// GetDeviceMetrics retrieves all available metrics for a device.
func (s *TimescaleStorage) GetDeviceMetrics(ctx context.Context, deviceID string) ([]string, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT DISTINCT metric_name
		FROM device_telemetry
		WHERE device_id = $1
		ORDER BY metric_name
	`, deviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to query device metrics: %w", err)
	}
	defer rows.Close()

	var metrics []string
	for rows.Next() {
		var metricName string
		if err := rows.Scan(&metricName); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		metrics = append(metrics, metricName)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return metrics, nil
}

// Close closes the storage connection.
func (s *TimescaleStorage) Close() error {
	s.pool.Close()
	return nil
}

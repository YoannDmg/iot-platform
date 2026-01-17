// Package storage provides the storage interface and implementations for telemetry data.
package storage

import (
	"context"

	pb "github.com/yourusername/iot-platform/shared/proto/telemetry"
)

// Storage defines the interface for telemetry data persistence.
type Storage interface {
	// InsertTelemetry inserts a single telemetry point.
	InsertTelemetry(ctx context.Context, deviceID, metricName string, value float64, unit string, timestamp int64, metadata map[string]string) error

	// InsertTelemetryBatch inserts multiple telemetry points in a single transaction.
	InsertTelemetryBatch(ctx context.Context, points []*TelemetryPoint) error

	// GetTelemetry retrieves telemetry data for a device within a time range.
	GetTelemetry(ctx context.Context, deviceID, metricName string, fromTime, toTime int64, limit int) ([]*pb.TelemetryPoint, error)

	// GetTelemetryAggregated retrieves aggregated telemetry data.
	GetTelemetryAggregated(ctx context.Context, deviceID, metricName string, fromTime, toTime int64, interval string) ([]*pb.TelemetryAggregation, error)

	// GetLatestMetric retrieves the latest value for a specific metric.
	GetLatestMetric(ctx context.Context, deviceID, metricName string) (*pb.TelemetryPoint, error)

	// GetDeviceMetrics retrieves all available metrics for a device.
	GetDeviceMetrics(ctx context.Context, deviceID string) ([]string, error)

	// Close closes the storage connection.
	Close() error
}

// TelemetryPoint represents a single telemetry data point for batch operations.
type TelemetryPoint struct {
	DeviceID   string
	MetricName string
	Value      float64
	Unit       string
	Timestamp  int64
	Metadata   map[string]string
}

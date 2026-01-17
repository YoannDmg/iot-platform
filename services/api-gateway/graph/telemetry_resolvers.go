package graph

import (
	"context"
	"log"

	"github.com/yourusername/iot-platform/services/api-gateway/graph/model"
	telemetrypb "github.com/yourusername/iot-platform/shared/proto/telemetry"
)

// DeviceTelemetryImpl retrieves raw telemetry data for a device.
func (r *queryResolver) DeviceTelemetryImpl(ctx context.Context, deviceID string, metricName string, from int, to int, limit *int) (*model.TelemetrySeries, error) {
	log.Printf("📊 Query deviceTelemetry: device=%s, metric=%s", deviceID, metricName)

	limitValue := int32(1000)
	if limit != nil {
		limitValue = int32(*limit)
	}

	resp, err := r.TelemetryClient.GetTelemetry(ctx, &telemetrypb.GetTelemetryRequest{
		DeviceId:   deviceID,
		MetricName: metricName,
		FromTime:   int64(from),
		ToTime:     int64(to),
		Limit:      limitValue,
	})
	if err != nil {
		log.Printf("❌ Failed to get telemetry: %v", err)
		return nil, err
	}

	points := make([]*model.TelemetryPoint, len(resp.Points))
	for i, p := range resp.Points {
		points[i] = &model.TelemetryPoint{
			Time:  int(p.Time),
			Value: p.Value,
			Unit:  &p.Unit,
		}
	}

	return &model.TelemetrySeries{
		MetricName: metricName,
		Points:     points,
	}, nil
}

// DeviceTelemetryAggregatedImpl retrieves aggregated telemetry data.
func (r *queryResolver) DeviceTelemetryAggregatedImpl(ctx context.Context, deviceID string, metricName string, from int, to int, interval string) ([]*model.TelemetryAggregation, error) {
	log.Printf("📊 Query deviceTelemetryAggregated: device=%s, metric=%s, interval=%s", deviceID, metricName, interval)

	resp, err := r.TelemetryClient.GetTelemetryAggregated(ctx, &telemetrypb.GetTelemetryAggregatedRequest{
		DeviceId:   deviceID,
		MetricName: metricName,
		FromTime:   int64(from),
		ToTime:     int64(to),
		Interval:   interval,
	})
	if err != nil {
		log.Printf("❌ Failed to get aggregated telemetry: %v", err)
		return nil, err
	}

	aggregations := make([]*model.TelemetryAggregation, len(resp.Aggregations))
	for i, a := range resp.Aggregations {
		aggregations[i] = &model.TelemetryAggregation{
			Bucket: a.Bucket,
			Avg:    a.Avg,
			Min:    a.Min,
			Max:    a.Max,
			Count:  int(a.Count),
		}
	}

	return aggregations, nil
}

// DeviceLatestMetricImpl retrieves the latest value for a specific metric.
func (r *queryResolver) DeviceLatestMetricImpl(ctx context.Context, deviceID string, metricName string) (*model.TelemetryPoint, error) {
	log.Printf("📊 Query deviceLatestMetric: device=%s, metric=%s", deviceID, metricName)

	resp, err := r.TelemetryClient.GetLatestMetric(ctx, &telemetrypb.GetLatestMetricRequest{
		DeviceId:   deviceID,
		MetricName: metricName,
	})
	if err != nil {
		log.Printf("❌ Failed to get latest metric: %v", err)
		return nil, nil // Return nil instead of error for optional field
	}

	if resp.Point == nil {
		return nil, nil
	}

	return &model.TelemetryPoint{
		Time:  int(resp.Point.Time),
		Value: resp.Point.Value,
		Unit:  &resp.Point.Unit,
	}, nil
}

// DeviceMetricsImpl retrieves all available metrics for a device.
func (r *queryResolver) DeviceMetricsImpl(ctx context.Context, deviceID string) ([]string, error) {
	log.Printf("📊 Query deviceMetrics: device=%s", deviceID)

	resp, err := r.TelemetryClient.GetDeviceMetrics(ctx, &telemetrypb.GetDeviceMetricsRequest{
		DeviceId: deviceID,
	})
	if err != nil {
		log.Printf("❌ Failed to get device metrics: %v", err)
		return []string{}, nil // Return empty array instead of error
	}

	return resp.Metrics, nil
}

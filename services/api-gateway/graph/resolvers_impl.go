package graph

import (
	"context"
	"fmt"

	devicepb "github.com/yourusername/iot-platform/shared/proto/device"
	"github.com/yourusername/iot-platform/services/api-gateway/graph/model"
)

// Helper functions to convert between Protobuf and GraphQL types

func protoToGraphQLDevice(d *devicepb.Device) *model.Device {
	if d == nil {
		return nil
	}

	// Convert map to slice of MetadataEntry
	metadata := make([]*model.MetadataEntry, 0, len(d.Metadata))
	for k, v := range d.Metadata {
		metadata = append(metadata, &model.MetadataEntry{
			Key:   k,
			Value: v,
		})
	}

	return &model.Device{
		ID:        d.Id,
		Name:      d.Name,
		Type:      d.Type,
		Status:    protoToGraphQLStatus(d.Status),
		CreatedAt: int(d.CreatedAt),
		LastSeen:  int(d.LastSeen),
		Metadata:  metadata,
	}
}

func protoToGraphQLStatus(s devicepb.DeviceStatus) model.DeviceStatus {
	switch s {
	case devicepb.DeviceStatus_ONLINE:
		return model.DeviceStatusOnline
	case devicepb.DeviceStatus_OFFLINE:
		return model.DeviceStatusOffline
	case devicepb.DeviceStatus_ERROR:
		return model.DeviceStatusError
	case devicepb.DeviceStatus_MAINTENANCE:
		return model.DeviceStatusMaintenance
	default:
		return model.DeviceStatusUnknown
	}
}

func graphQLToProtoStatus(s *model.DeviceStatus) devicepb.DeviceStatus {
	if s == nil {
		return devicepb.DeviceStatus_UNKNOWN
	}

	switch *s {
	case model.DeviceStatusOnline:
		return devicepb.DeviceStatus_ONLINE
	case model.DeviceStatusOffline:
		return devicepb.DeviceStatus_OFFLINE
	case model.DeviceStatusError:
		return devicepb.DeviceStatus_ERROR
	case model.DeviceStatusMaintenance:
		return devicepb.DeviceStatus_MAINTENANCE
	default:
		return devicepb.DeviceStatus_UNKNOWN
	}
}

// Mutation resolvers

func (r *mutationResolver) CreateDeviceImpl(ctx context.Context, input model.CreateDeviceInput) (*model.Device, error) {
	// Convert GraphQL input to Protobuf request
	// Convert slice to map
	metadata := make(map[string]string)
	if input.Metadata != nil {
		for _, kv := range input.Metadata {
			metadata[kv.Key] = kv.Value
		}
	}

	req := &devicepb.CreateDeviceRequest{
		Name:     input.Name,
		Type:     input.Type,
		Metadata: metadata,
	}

	// Call Device Manager via gRPC
	resp, err := r.DeviceClient.CreateDevice(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create device: %w", err)
	}

	// Convert Protobuf response to GraphQL model
	return protoToGraphQLDevice(resp.Device), nil
}

func (r *mutationResolver) UpdateDeviceImpl(ctx context.Context, input model.UpdateDeviceInput) (*model.Device, error) {
	// Convert metadata if provided
	var metadata map[string]string
	if input.Metadata != nil {
		metadata = make(map[string]string)
		for _, kv := range input.Metadata {
			metadata[kv.Key] = kv.Value
		}
	}

	req := &devicepb.UpdateDeviceRequest{
		Id:       input.ID,
		Name:     stringPtrToValue(input.Name),
		Status:   graphQLToProtoStatus(input.Status),
		Metadata: metadata,
	}

	resp, err := r.DeviceClient.UpdateDevice(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update device: %w", err)
	}

	return protoToGraphQLDevice(resp.Device), nil
}

func (r *mutationResolver) DeleteDeviceImpl(ctx context.Context, id string) (*model.DeleteResult, error) {
	req := &devicepb.DeleteDeviceRequest{
		Id: id,
	}

	resp, err := r.DeviceClient.DeleteDevice(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to delete device: %w", err)
	}

	return &model.DeleteResult{
		Success: resp.Success,
		Message: resp.Message,
	}, nil
}

// Query resolvers

func (r *queryResolver) DeviceImpl(ctx context.Context, id string) (*model.Device, error) {
	req := &devicepb.GetDeviceRequest{
		Id: id,
	}

	resp, err := r.DeviceClient.GetDevice(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get device: %w", err)
	}

	return protoToGraphQLDevice(resp.Device), nil
}

func (r *queryResolver) DevicesImpl(ctx context.Context, page *int, pageSize *int, typeArg *string, status *model.DeviceStatus) (*model.DeviceConnection, error) {
	// Default values
	p := int32(1)
	ps := int32(10)

	if page != nil {
		p = int32(*page)
	}
	if pageSize != nil {
		ps = int32(*pageSize)
	}

	req := &devicepb.ListDevicesRequest{
		Page:     p,
		PageSize: ps,
	}

	resp, err := r.DeviceClient.ListDevices(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list devices: %w", err)
	}

	// Convert devices
	devices := make([]*model.Device, len(resp.Devices))
	for i, d := range resp.Devices {
		devices[i] = protoToGraphQLDevice(d)
	}

	// Client-side filtering (TODO: move to server-side)
	if typeArg != nil || status != nil {
		filtered := make([]*model.Device, 0)
		for _, d := range devices {
			match := true
			if typeArg != nil && d.Type != *typeArg {
				match = false
			}
			if status != nil && d.Status != *status {
				match = false
			}
			if match {
				filtered = append(filtered, d)
			}
		}
		devices = filtered
	}

	return &model.DeviceConnection{
		Devices:  devices,
		Total:    int(resp.Total),
		Page:     int(resp.Page),
		PageSize: int(resp.PageSize),
	}, nil
}

func (r *queryResolver) StatsImpl(ctx context.Context) (*model.Stats, error) {
	// Get all devices to compute stats
	req := &devicepb.ListDevicesRequest{
		Page:     1,
		PageSize: 1000, // TODO: Implement server-side stats endpoint
	}

	resp, err := r.DeviceClient.ListDevices(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	// Compute stats
	total := len(resp.Devices)
	online := 0
	offline := 0
	errorDevices := 0

	for _, d := range resp.Devices {
		switch d.Status {
		case devicepb.DeviceStatus_ONLINE:
			online++
		case devicepb.DeviceStatus_OFFLINE:
			offline++
		case devicepb.DeviceStatus_ERROR:
			errorDevices++
		}
	}

	return &model.Stats{
		TotalDevices:   total,
		OnlineDevices:  online,
		OfflineDevices: offline,
		ErrorDevices:   errorDevices,
	}, nil
}

// Helper functions

func stringPtrToValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

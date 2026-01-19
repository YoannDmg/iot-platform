package pubsub

import (
	"sync"

	"github.com/yourusername/iot-platform/services/api-gateway/graph/model"
)

// Broker manages subscriptions for real-time telemetry data
type Broker struct {
	subscribers map[string]map[chan *model.TelemetryPoint]struct{} // deviceID -> set of channels
	mu          sync.RWMutex
}

// NewBroker creates a new subscription broker
func NewBroker() *Broker {
	return &Broker{
		subscribers: make(map[string]map[chan *model.TelemetryPoint]struct{}),
	}
}

// Subscribe creates a new subscription channel for a device
func (b *Broker) Subscribe(deviceID string) chan *model.TelemetryPoint {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan *model.TelemetryPoint, 10) // buffered channel

	if b.subscribers[deviceID] == nil {
		b.subscribers[deviceID] = make(map[chan *model.TelemetryPoint]struct{})
	}
	b.subscribers[deviceID][ch] = struct{}{}

	return ch
}

// Unsubscribe removes a subscription channel
func (b *Broker) Unsubscribe(deviceID string, ch chan *model.TelemetryPoint) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if subs, ok := b.subscribers[deviceID]; ok {
		delete(subs, ch)
		close(ch)

		// Clean up empty device entries
		if len(subs) == 0 {
			delete(b.subscribers, deviceID)
		}
	}
}

// Publish sends a telemetry point to all subscribers of a device
func (b *Broker) Publish(deviceID string, point *model.TelemetryPoint) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if subs, ok := b.subscribers[deviceID]; ok {
		for ch := range subs {
			// Non-blocking send to avoid slow subscribers blocking others
			select {
			case ch <- point:
			default:
				// Channel full, skip this message for this subscriber
			}
		}
	}
}

// SubscriberCount returns the number of active subscribers for a device
func (b *Broker) SubscriberCount(deviceID string) int {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if subs, ok := b.subscribers[deviceID]; ok {
		return len(subs)
	}
	return 0
}

// TotalSubscribers returns the total number of active subscriptions
func (b *Broker) TotalSubscribers() int {
	b.mu.RLock()
	defer b.mu.RUnlock()

	count := 0
	for _, subs := range b.subscribers {
		count += len(subs)
	}
	return count
}

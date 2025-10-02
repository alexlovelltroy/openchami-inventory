package events

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestNewInMemoryEventBus(t *testing.T) {
	bus := NewInMemoryEventBus(100, 5)
	defer bus.Close()

	if bus == nil {
		t.Fatal("Expected non-nil event bus")
	}

	if bus.bufferSize != 100 {
		t.Errorf("Expected buffer size 100, got %d", bus.bufferSize)
	}

	if bus.workerCount != 5 {
		t.Errorf("Expected worker count 5, got %d", bus.workerCount)
	}
}

func TestNewInMemoryEventBus_Defaults(t *testing.T) {
	bus := NewInMemoryEventBus(0, 0)
	defer bus.Close()

	if bus.bufferSize != 1000 {
		t.Errorf("Expected default buffer size 1000, got %d", bus.bufferSize)
	}

	if bus.workerCount != 10 {
		t.Errorf("Expected default worker count 10, got %d", bus.workerCount)
	}
}

func TestInMemoryEventBus_PublishAndSubscribe(t *testing.T) {
	bus := NewInMemoryEventBus(100, 2)
	bus.Start()
	defer bus.Close()

	// Create a channel to receive events
	received := make(chan Event, 1)

	// Subscribe to events
	_, err := bus.Subscribe("io.openchami.inventory.bmc.connected", func(ctx context.Context, event Event) error {
		received <- event
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// Create and publish an event
	event, err := NewResourceEvent(
		"io.openchami.inventory.bmc.connected",
		"BMC",
		"bmc-123",
		map[string]string{"address": "10.0.0.1"},
	)
	if err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}

	if err := bus.Publish(context.Background(), *event); err != nil {
		t.Fatalf("Failed to publish event: %v", err)
	}

	// Wait for event to be received
	select {
	case rcvd := <-received:
		if rcvd.Type() != event.Type() {
			t.Errorf("Expected event type %s, got %s", event.Type(), rcvd.Type())
		}
		if rcvd.ID() != event.ID() {
			t.Errorf("Expected event ID %s, got %s", event.ID(), rcvd.ID())
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for event")
	}
}

func TestInMemoryEventBus_WildcardSubscription(t *testing.T) {
	bus := NewInMemoryEventBus(100, 2)
	bus.Start()
	defer bus.Close()

	// Track received events
	var receivedMu sync.Mutex
	receivedTypes := []string{}

	// Subscribe to all BMC events with wildcard
	_, err := bus.Subscribe("io.openchami.inventory.bmc.*", func(ctx context.Context, event Event) error {
		receivedMu.Lock()
		receivedTypes = append(receivedTypes, event.Type())
		receivedMu.Unlock()
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// Publish multiple BMC events
	events := []string{
		"io.openchami.inventory.bmc.connected",
		"io.openchami.inventory.bmc.disconnected",
		"io.openchami.inventory.bmc.updated",
	}

	for _, eventType := range events {
		event, _ := NewResourceEvent(eventType, "BMC", "bmc-123", nil)
		bus.Publish(context.Background(), *event)
	}

	// Wait for events
	time.Sleep(100 * time.Millisecond)

	receivedMu.Lock()
	defer receivedMu.Unlock()

	if len(receivedTypes) != 3 {
		t.Errorf("Expected 3 events, got %d", len(receivedTypes))
	}

	// Check that all event types were received
	for _, expectedType := range events {
		found := false
		for _, rcvdType := range receivedTypes {
			if rcvdType == expectedType {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Event type %s not received", expectedType)
		}
	}
}

func TestInMemoryEventBus_MultiWildcardSubscription(t *testing.T) {
	bus := NewInMemoryEventBus(100, 2)
	bus.Start()
	defer bus.Close()

	var receivedMu sync.Mutex
	receivedTypes := []string{}

	// Subscribe to all inventory events with multi-wildcard
	_, err := bus.Subscribe("io.openchami.inventory.**", func(ctx context.Context, event Event) error {
		receivedMu.Lock()
		receivedTypes = append(receivedTypes, event.Type())
		receivedMu.Unlock()
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// Publish events from different resource types
	events := []string{
		"io.openchami.inventory.bmc.connected",
		"io.openchami.inventory.fru.discovered",
		"io.openchami.inventory.node.ready",
	}

	for _, eventType := range events {
		event, _ := NewResourceEvent(eventType, "BMC", "bmc-123", nil)
		bus.Publish(context.Background(), *event)
	}

	// Wait for events
	time.Sleep(100 * time.Millisecond)

	receivedMu.Lock()
	defer receivedMu.Unlock()

	if len(receivedTypes) != 3 {
		t.Errorf("Expected 3 events, got %d", len(receivedTypes))
	}
}

func TestInMemoryEventBus_Unsubscribe(t *testing.T) {
	bus := NewInMemoryEventBus(100, 2)
	bus.Start()
	defer bus.Close()

	received := make(chan Event, 10)

	// Subscribe
	id, err := bus.Subscribe("io.openchami.inventory.bmc.*", func(ctx context.Context, event Event) error {
		received <- event
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// Publish event (should be received)
	event1, _ := NewResourceEvent("io.openchami.inventory.bmc.connected", "BMC", "bmc-123", nil)
	bus.Publish(context.Background(), *event1)

	// Wait for event
	select {
	case <-received:
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout waiting for first event")
	}

	// Unsubscribe
	if err := bus.Unsubscribe(id); err != nil {
		t.Fatalf("Failed to unsubscribe: %v", err)
	}

	// Publish another event (should NOT be received)
	event2, _ := NewResourceEvent("io.openchami.inventory.bmc.disconnected", "BMC", "bmc-123", nil)
	bus.Publish(context.Background(), *event2)

	// Wait a bit to see if event arrives
	select {
	case <-received:
		t.Fatal("Received event after unsubscribing")
	case <-time.After(100 * time.Millisecond):
		// Expected - no event received
	}
}

func TestMatchesPattern(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
		pattern   string
		expected  bool
	}{
		{
			name:      "exact match",
			eventType: "io.openchami.inventory.bmc.connected",
			pattern:   "io.openchami.inventory.bmc.connected",
			expected:  true,
		},
		{
			name:      "single wildcard match",
			eventType: "io.openchami.inventory.bmc.connected",
			pattern:   "io.openchami.inventory.bmc.*",
			expected:  true,
		},
		{
			name:      "multi wildcard match",
			eventType: "io.openchami.inventory.bmc.connected",
			pattern:   "io.openchami.inventory.**",
			expected:  true,
		},
		{
			name:      "multi wildcard from root",
			eventType: "io.openchami.inventory.bmc.connected",
			pattern:   "io.**",
			expected:  true,
		},
		{
			name:      "no match - different resource",
			eventType: "io.openchami.inventory.bmc.connected",
			pattern:   "io.openchami.inventory.fru.*",
			expected:  false,
		},
		{
			name:      "no match - pattern longer",
			eventType: "io.openchami.inventory.bmc",
			pattern:   "io.openchami.inventory.bmc.*",
			expected:  false,
		},
		{
			name:      "no match - event longer",
			eventType: "io.openchami.inventory.bmc.connected",
			pattern:   "io.openchami.inventory.bmc",
			expected:  false,
		},
		{
			name:      "middle wildcard",
			eventType: "io.openchami.inventory.bmc.connected",
			pattern:   "io.openchami.*.bmc.connected",
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesPattern(tt.eventType, tt.pattern)
			if result != tt.expected {
				t.Errorf("matchesPattern(%q, %q) = %v, expected %v",
					tt.eventType, tt.pattern, result, tt.expected)
			}
		})
	}
}

func TestInMemoryEventBus_MultipleSubscribers(t *testing.T) {
	bus := NewInMemoryEventBus(100, 2)
	bus.Start()
	defer bus.Close()

	// Create multiple subscribers
	received1 := make(chan Event, 1)
	received2 := make(chan Event, 1)

	bus.Subscribe("io.openchami.inventory.bmc.connected", func(ctx context.Context, event Event) error {
		received1 <- event
		return nil
	})

	bus.Subscribe("io.openchami.inventory.bmc.*", func(ctx context.Context, event Event) error {
		received2 <- event
		return nil
	})

	// Publish event
	event, _ := NewResourceEvent("io.openchami.inventory.bmc.connected", "BMC", "bmc-123", nil)
	bus.Publish(context.Background(), *event)

	// Both subscribers should receive the event
	timeout := time.After(1 * time.Second)

	select {
	case <-received1:
		// Expected
	case <-timeout:
		t.Fatal("Timeout waiting for event on subscriber 1")
	}

	select {
	case <-received2:
		// Expected
	case <-timeout:
		t.Fatal("Timeout waiting for event on subscriber 2")
	}
}

func TestInMemoryEventBus_Close(t *testing.T) {
	bus := NewInMemoryEventBus(100, 2)
	bus.Start()

	// Close the bus
	if err := bus.Close(); err != nil {
		t.Fatalf("Failed to close bus: %v", err)
	}

	// Publishing after close should fail
	event, _ := NewResourceEvent("io.openchami.inventory.bmc.connected", "BMC", "bmc-123", nil)
	err := bus.Publish(context.Background(), *event)
	if err == nil {
		t.Fatal("Expected error when publishing to closed bus")
	}
}

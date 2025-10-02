// Package events provides a CloudEvents-based event system for inventory resources.
//
// This package implements event publishing and subscription using the CloudEvents
// standard (https://cloudevents.io/), enabling interoperability with external systems
// and cloud-native event tooling.
package events

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

// Event wraps CloudEvents specification
type Event struct {
	cloudevents.Event
}

// NewEvent creates a CloudEvents-compliant event
func NewEvent(eventType, source string, data interface{}) (*Event, error) {
	event := cloudevents.NewEvent()
	event.SetID(generateEventID())
	event.SetType(eventType)
	event.SetSource(source)
	event.SetTime(time.Now())
	event.SetDataContentType("application/json")

	if err := event.SetData(cloudevents.ApplicationJSON, data); err != nil {
		return nil, fmt.Errorf("failed to set event data: %w", err)
	}

	return &Event{Event: event}, nil
}

// NewResourceEvent creates an event for a resource change
func NewResourceEvent(eventType, resourceKind, resourceUID string, data interface{}) (*Event, error) {
	source := fmt.Sprintf("/inventory/%s/%s", resourceKind, resourceUID)
	event, err := NewEvent(eventType, source, data)
	if err != nil {
		return nil, err
	}

	// Add inventory-specific extension attributes
	event.SetExtension("inventoryresourcekind", resourceKind)
	event.SetExtension("inventoryresourceuid", resourceUID)

	return event, nil
}

// ResourceKind returns the resource kind extension attribute
func (e *Event) ResourceKind() string {
	if val, ok := e.Extensions()["inventoryresourcekind"]; ok {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}

// ResourceUID returns the resource UID extension attribute
func (e *Event) ResourceUID() string {
	if val, ok := e.Extensions()["inventoryresourceuid"]; ok {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}

// EventHandler processes CloudEvents
type EventHandler func(ctx context.Context, event Event) error

// SubscriptionID uniquely identifies a subscription
type SubscriptionID string

// EventBus manages event publishing and subscription
type EventBus interface {
	// Publish a CloudEvent
	Publish(ctx context.Context, event Event) error

	// Subscribe to events by type pattern (supports wildcards)
	Subscribe(eventType string, handler EventHandler) (SubscriptionID, error)

	// Unsubscribe from events
	Unsubscribe(id SubscriptionID) error

	// Close the event bus
	Close() error
}

// generateEventID generates a unique event ID
func generateEventID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return "evt-" + hex.EncodeToString(b)[:12]
}

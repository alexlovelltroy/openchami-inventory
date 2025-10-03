// Package events provides inventory-specific event utilities.
//
// This package re-exports the fabrica event system and provides
// inventory-specific event creation helpers.
package events

import (
	"fmt"

	"github.com/alexlovelltroy/fabrica/pkg/events"
)

// Re-export fabrica event types for backwards compatibility
type (
	Event          = events.Event
	EventHandler   = events.EventHandler
	SubscriptionID = events.SubscriptionID
	EventBus       = events.EventBus
)

// Re-export fabrica event functions
var (
	NewEvent          = events.NewEvent
	NewInMemoryEventBus = events.NewInMemoryEventBus
)

// NewResourceEvent creates an inventory-specific resource event
//
// This wraps the fabrica NewResourceEvent with inventory-specific conventions:
//   - Source format: /inventory/{kind}/{uid}
//   - Extension attributes: inventoryresourcekind, inventoryresourceuid
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

// ResourceKind returns the inventory resource kind extension attribute
func ResourceKind(e Event) string {
	if val, ok := e.Extensions()["inventoryresourcekind"]; ok {
		if s, ok := val.(string); ok {
			return s
		}
	}
	// Fallback to generic resourcekind
	if val, ok := e.Extensions()["resourcekind"]; ok {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}

// ResourceUID returns the inventory resource UID extension attribute
func ResourceUID(e Event) string {
	if val, ok := e.Extensions()["inventoryresourceuid"]; ok {
		if s, ok := val.(string); ok {
			return s
		}
	}
	// Fallback to generic resourceuid
	if val, ok := e.Extensions()["resourceuid"]; ok {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}

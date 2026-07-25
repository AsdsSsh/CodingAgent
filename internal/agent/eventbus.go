package agent

import "sync"

// EventBus is a lightweight in-process pub/sub for agent events.
// Thread-safe: subscribers stored with sync.RWMutex, emit iterates a snapshot.
type EventBus struct {
	mu          sync.RWMutex
	subscribers map[string][]chan any
}

// NewEventBus creates an empty event bus.
func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string][]chan any),
	}
}

// Subscribe registers a channel to receive events of a given type name.
// Returns an unsubscribe function.
func (eb *EventBus) Subscribe(eventType string, ch chan any) func() {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.subscribers[eventType] = append(eb.subscribers[eventType], ch)
	return func() {
		eb.mu.Lock()
		defer eb.mu.Unlock()
		chs := eb.subscribers[eventType]
		for i, c := range chs {
			if c == ch {
				eb.subscribers[eventType] = append(chs[:i], chs[i+1:]...)
				return
			}
		}
	}
}

// Emit publishes an event to all subscribers of its type name.
// Non-blocking: if a subscriber's buffer is full, the event is dropped.
func (eb *EventBus) Emit(eventType string, event any) {
	eb.mu.RLock()
	chs := eb.subscribers[eventType]
	eb.mu.RUnlock()

	for _, ch := range chs {
		select {
		case ch <- event:
		default:
			// Drop if buffer full (non-blocking)
		}
	}
}

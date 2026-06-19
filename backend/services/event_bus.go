package services

import (
	"context"
	"log/slog"
	"sync"
)

// Handler is a subscriber callback. ctx is the publishing context.
type Handler func(ctx context.Context, e Event)

// EventBus is a synchronous, in-process pub-sub bus with panic recovery.
// Handlers run in the caller's goroutine; a panic in one handler is logged
// and suppressed so subsequent handlers still execute.
type EventBus struct {
	mu       sync.RWMutex
	handlers map[string][]Handler
}

func NewEventBus() *EventBus {
	return &EventBus{handlers: make(map[string][]Handler)}
}

// Subscribe registers h to receive events of the given type.
// Pass "*" to receive every event regardless of type.
func (b *EventBus) Subscribe(eventType string, h Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventType] = append(b.handlers[eventType], h)
}

// Publish delivers e to all matching subscribers synchronously.
func (b *EventBus) Publish(ctx context.Context, e Event) {
	b.mu.RLock()
	specific := append([]Handler(nil), b.handlers[e.EventType()]...)
	wildcard := append([]Handler(nil), b.handlers["*"]...)
	b.mu.RUnlock()

	for _, h := range specific {
		safeDispatch(ctx, e, h)
	}
	for _, h := range wildcard {
		safeDispatch(ctx, e, h)
	}
}

func safeDispatch(ctx context.Context, e Event, h Handler) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("event bus: subscriber panicked", "event_type", e.EventType(), "panic", r)
		}
	}()
	h(ctx, e)
}

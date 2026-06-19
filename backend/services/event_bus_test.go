package services

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testEvent struct{ val string }

func (e testEvent) EventType() string { return "test.event" }

type otherEvent struct{}

func (e otherEvent) EventType() string { return "other.event" }

func TestEventBus_SpecificSubscriberReceivesMatchingEvent(t *testing.T) {
	bus := NewEventBus()
	var got []string
	bus.Subscribe("test.event", func(_ context.Context, e Event) {
		got = append(got, e.(testEvent).val)
	})

	bus.Publish(context.Background(), testEvent{val: "hello"})
	bus.Publish(context.Background(), otherEvent{})

	require.Equal(t, []string{"hello"}, got)
}

func TestEventBus_WildcardSubscriberReceivesAllEvents(t *testing.T) {
	bus := NewEventBus()
	var count int32
	bus.Subscribe("*", func(_ context.Context, e Event) {
		atomic.AddInt32(&count, 1)
	})

	bus.Publish(context.Background(), testEvent{val: "a"})
	bus.Publish(context.Background(), otherEvent{})

	assert.Equal(t, int32(2), atomic.LoadInt32(&count))
}

func TestEventBus_MultipleSubscribersAllReceiveEvent(t *testing.T) {
	bus := NewEventBus()
	var a, b int32
	bus.Subscribe("test.event", func(_ context.Context, e Event) { atomic.AddInt32(&a, 1) })
	bus.Subscribe("test.event", func(_ context.Context, e Event) { atomic.AddInt32(&b, 1) })

	bus.Publish(context.Background(), testEvent{val: "x"})

	assert.Equal(t, int32(1), atomic.LoadInt32(&a))
	assert.Equal(t, int32(1), atomic.LoadInt32(&b))
}

func TestEventBus_PanicInSubscriberDoesNotPropagateToPublisher(t *testing.T) {
	bus := NewEventBus()
	var afterPanic int32
	bus.Subscribe("test.event", func(_ context.Context, e Event) { panic("boom") })
	bus.Subscribe("test.event", func(_ context.Context, e Event) { atomic.AddInt32(&afterPanic, 1) })

	require.NotPanics(t, func() {
		bus.Publish(context.Background(), testEvent{val: "y"})
	})
	assert.Equal(t, int32(1), atomic.LoadInt32(&afterPanic))
}

func TestEventBus_AuditServiceSubscription(t *testing.T) {
	bus := NewEventBus()
	var logged []string
	// Simulate audit subscription by hooking the wildcard slot directly.
	bus.Subscribe("*", func(_ context.Context, e Event) {
		if ae, ok := e.(AuditableEvent); ok {
			entry := ae.ToAuditEntry()
			logged = append(logged, entry.Action)
		}
	})

	bus.Publish(context.Background(), RequestCommentAddedEvent{
		RequestType: "subnet", RequestID: 1, AuthorID: 2, CommentID: 3,
	})
	bus.Publish(context.Background(), testEvent{val: "ignored"}) // not AuditableEvent

	require.Equal(t, []string{"request_comment_added"}, logged)
}

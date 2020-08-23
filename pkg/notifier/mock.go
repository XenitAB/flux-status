package notifier

import (
	"context"
)

// Mock implements a dummy notifier that doesn nothing.
type Mock struct {
	Events chan Event
}

// NewMock creates and returns a Mock instance.
func NewMock() *Mock {
	return &Mock{
		Events: make(chan Event, 100),
	}
}

// Send adds the event to the Events channel buffer.
func (n *Mock) Send(ctx context.Context, e Event) error {
	n.Events <- e
	return nil
}

// Get always returns nil.
func (n *Mock) Get(commitID string, action string) (*Status, error) {
	return nil, nil
}

// String returns the name of the struct.
func (n *Mock) String() string {
	return "Mock"
}

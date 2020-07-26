package notifier

import (
	"context"
)

type Mock struct {
	Events chan Event
}

func NewMock() *Mock {
	return &Mock{
		Events: make(chan Event, 100),
	}
}

func (n *Mock) Send(ctx context.Context, e Event) error {
	n.Events <- e
	return nil
}

func (n *Mock) Get(commitId string, action string) (*Status, error) {
	return nil, nil
}

func (n *Mock) String() string {
	return "Mock"
}

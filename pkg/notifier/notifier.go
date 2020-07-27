package notifier

import (
	"context"
	"errors"
)

const StatusId string = "flux-status"

type EventType string

const (
	EventTypeSync     EventType = "sync"
	EventTypeWorkload EventType = "workload"
)

type EventState string

const (
	EventStateFailed    EventState = "failed"
	EventStatePending   EventState = "pending"
	EventStateSucceeded EventState = "succeeded"
	EventStateCanceled  EventState = "canceled"
)

type Event struct {
	Type     EventType
	Message  string
	CommitId string
	State    EventState
}

type Status struct {
	Name  string     `json:"name"`
	State EventState `json:"state"`
}

type Notifier interface {
	Send(context.Context, Event) error
	Get(string, string) (*Status, error)
	String() string
}

func GetNotifier(inst string, url string, azdoPat string, gitlabToken string) (Notifier, error) {
	gitlab, err := NewGitlab(inst, url, gitlabToken)
	if err == nil {
		return gitlab, nil
	}

	azdo, err := NewAzureDevops(inst, url, azdoPat)
	if err == nil {
		return azdo, nil
	}

	return nil, errors.New("Could not find a compatible Notifier")
}

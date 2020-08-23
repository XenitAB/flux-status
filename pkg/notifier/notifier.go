package notifier

import (
	"context"
	"errors"
)

// StatusID is a project specific identifier to avoid conflicts in the commit status.
const StatusID string = "flux-status"

// EventType represents the different types of actions an event can occur for.
type EventType string

// These constants represents all valid EventType values.
const (
	// EventTypeSync when the remote state has successfully reconciled.
	EventTypeSync EventType = "sync"
	// EventTypeWorkload occurs when all workloads have started.
	EventTypeWorkload EventType = "workload"
)

// EventState represents the different states an event can be in.
type EventState string

// These constants represents all valid EventState values.
const (
	// EventStateFailed occurs when an action has failed.
	EventStateFailed EventState = "failed"
	// EventStatePending occurs when an event is not yet completed.
	EventStatePending EventState = "pending"
	// EventStateSucceeded occurs when an event has completed successfully.
	EventStateSucceeded EventState = "succeeded"
	// EventStateCanceled occurs when an event has been preemptively canceled.
	EventStateCanceled EventState = "canceled"
)

// Event wraps information about a specific event occurence.
type Event struct {
	Type     EventType
	Message  string
	CommitID string
	State    EventState
}

// Status represents the current status of a commit id.
type Status struct {
	Name  string     `json:"name"`
	State EventState `json:"state"`
}

// Notifier is the interface that wraps the required methods to send events to a git provider.
type Notifier interface {
	Send(context.Context, Event) error
	Get(string, string) (*Status, error)
	String() string
}

// GetNotifier returns the best matching notifier given the configuration data.
// It works by attempting to create each available notifier one by one, and returns
// the first one that succeededs.
func GetNotifier(inst string, url string, azdoPat string, glToken string, ghToken string) (Notifier, error) {
	github, err := NewGitHub(inst, url, ghToken)
	if err == nil {
		return github, nil
	}

	gitlab, err := NewGitlab(inst, url, glToken)
	if err == nil {
		return gitlab, nil
	}

	azdo, err := NewAzureDevops(inst, url, azdoPat)
	if err == nil {
		return azdo, nil
	}

	return nil, errors.New("Could not find a compatible Notifier")
}

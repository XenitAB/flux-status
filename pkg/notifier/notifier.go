package notifier

import "errors"

const StatusId string = "flux-status"

type EventState string

const (
	EventStateFailed    = "failed"
	EventStatePending   = "pending"
	EventStateSucceeded = "succeeded"
	EventStateCanceled  = "canceled"
)

type Event struct {
	Event    string
	Message  string
	CommitId string
	State    EventState
}

type Status struct {
	Name  string     `json:"name"`
	State EventState `json:"state"`
}

type Notifier interface {
	Send(Event) error
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

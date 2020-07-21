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
	Instance string
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

func GetNotifier(url string, azdoPat string, gitlabToken string) (Notifier, error) {
	gitlab, err := NewGitlab(gitlabToken, url)
	if err == nil {
		return gitlab, nil
	}

	azdo, err := NewAzureDevops(azdoPat, url)
	if err == nil {
		return azdo, nil
	}

	return nil, errors.New("Could not find a compatible Notifier")
}

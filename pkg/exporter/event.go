package exporter

type EventState int

const (
	EventStateFailed    = iota
	EventStatePending   = iota
	EventStateSucceeded = iota
)

type Event struct {
	Sender   string
	Message  string
	CommitId string
	State    EventState
}

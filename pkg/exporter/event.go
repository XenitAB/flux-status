package exporter

type EventState int

const (
	EventStateFailed    = iota
	EventStatePending   = iota
	EventStateSucceeded = iota
	EventStateCanceled  = iota
)

type Event struct {
	Id       string
	Event    string
	Instance string
	Message  string
	CommitId string
	State    EventState
}

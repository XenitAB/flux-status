package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/xenitab/flux-status/pkg/exporter"
	"github.com/xenitab/flux-status/pkg/flux"
)

func (s *Server) HandleEvents(w http.ResponseWriter, r *http.Request) {
	// Read Flux event
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	fluxEvent := flux.Event{}
	if err := json.Unmarshal(body, &fluxEvent); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Send exporter event
	event := convertToEvent(fluxEvent)
	if err := s.Exporter.Send(event); err != nil {
		s.Log.Error(err, "Could not send event through exporter")
		http.Error(w, err.Error(), 500)
		return
	}

	// Poll workload status
	if event.State != exporter.EventStateFailed {
		go func() {
			if err := s.Poller.Poll(event.CommitId, s.Exporter); err != nil {
				s.Log.Error(err, "Polling service status failed")
			}
		}()
	}

	s.Log.Info("Sent sync event", "commit-id", event.CommitId)
	w.WriteHeader(200)
}

func convertToEvent(e flux.Event) exporter.Event {
	commitId := e.Metadata.Commits[0].Revision

	var message string
	var state exporter.EventState
	if len(e.Metadata.Errors) == 0 {
		state = exporter.EventStateSucceeded
		message = "Succeeded"
	} else {
		state = exporter.EventStateFailed
		message = "Errors:"
		for _, err := range e.Metadata.Errors {
			message = message + err.Id + ","
		}
	}

	return exporter.Event{
		Sender:   "flux",
		Message:  message,
		CommitId: commitId,
		State:    state,
	}
}

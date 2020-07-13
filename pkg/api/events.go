package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/fluxcd/flux/pkg/event"

	"github.com/xenitab/flux-status/pkg/exporter"
)

func (s *Server) HandleEvents(w http.ResponseWriter, r *http.Request) {
	// Read Flux event
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	fluxEvent := event.Event{}
	if err := json.Unmarshal(body, &fluxEvent); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Send exporter event
	event, err := convertToEvent(fluxEvent)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if err := s.Exporter.Send(event); err != nil {
		s.Log.Error(err, "Could not send event through exporter")
		http.Error(w, err.Error(), 500)
		return
	}

	// Poll workload status
	if event.State != exporter.EventStateFailed {
		s.Poller.Stop()
		go func() {
			if err := s.Poller.Poll(event.CommitId, s.Exporter); err != nil {
				s.Log.Error(err, "Polling service status failed")
			}
		}()
	}

	s.Log.Info("Sent sync event", "commit-id", event.CommitId)
	w.WriteHeader(200)
}

func convertToEvent(e event.Event) (exporter.Event, error) {
	if e.Metadata.Type() != event.EventSync {
		return exporter.Event{}, fmt.Errorf("Could not parse event metatada type: %v", e.Metadata.Type())
	}

	syncMetadata := e.Metadata.(*event.SyncEventMetadata)
	commitId := syncMetadata.Commits[0].Revision

	var message string
	var state exporter.EventState
	if len(syncMetadata.Errors) == 0 {
		state = exporter.EventStateSucceeded
		message = "Succeeded"
	} else {
		state = exporter.EventStateFailed
		message = "Errors:"
		for _, err := range syncMetadata.Errors {
			message = message + err.ID.String() + ","
		}
	}

	event := exporter.Event{
		Id:       "flux-status",
		Instance: "dev",
		Event:    "sync",
		Message:  message,
		CommitId: commitId,
		State:    state,
	}

	return event, nil
}

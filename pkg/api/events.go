package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/fluxcd/flux/pkg/event"
	"github.com/go-logr/logr"

	"github.com/xenitab/flux-status/pkg/notifier"
)

func eventHandler(log logr.Logger, noti notifier.Notifier, events chan<- string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read Flux event
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Error(err, "Could not read request body")
			http.Error(w, err.Error(), 400)
			return
		}
		fluxEvent := event.Event{}
		if err := json.Unmarshal(body, &fluxEvent); err != nil {
			log.Error(err, "Could not parse Flux Event")
			http.Error(w, err.Error(), 400)
			return
		}

		// Send notifier event
		event, err := convertToEvent(fluxEvent)
		if err != nil {
			log.Error(err, "Could not convert event to notifier event")
			http.Error(w, err.Error(), 400)
			return
		}
		if err := noti.Send(r.Context(), event); err != nil {
			log.Error(err, "Could not send event through notifier")
			http.Error(w, err.Error(), 500)
			return
		}
		log.Info("Sent sync event", "commit-id", event.CommitId)

		// Only send commit Event if it not failed
		if event.State != notifier.EventStateFailed && events != nil {
			events <- event.CommitId
		}

		w.WriteHeader(200)
	})
}

func convertToEvent(e event.Event) (notifier.Event, error) {
	if e.Metadata.Type() != event.EventSync {
		return notifier.Event{}, fmt.Errorf("Could not parse event metatada type: %v", e.Metadata.Type())
	}

	syncMetadata := e.Metadata.(*event.SyncEventMetadata)
	commitId := syncMetadata.Commits[0].Revision

	var message string
	var state notifier.EventState
	if len(syncMetadata.Errors) == 0 {
		state = notifier.EventStateSucceeded
		message = "Succeeded"
	} else {
		state = notifier.EventStateFailed
		message = "Errors:"
		for _, err := range syncMetadata.Errors {
			message = message + err.ID.String() + ","
		}
	}

	return notifier.Event{
		Type:     notifier.EventTypeSync,
		Message:  message,
		CommitId: commitId,
		State:    state,
	}, nil
}

package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fluxcd/flux/pkg/event"
	"github.com/fluxcd/flux/pkg/resource"
	logr "github.com/go-logr/logr/testing"
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/gstruct"

	"github.com/xenitab/flux-status/pkg/notifier"
)

func TestEnabledPoller(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	log := logr.TestLogger{T: t}
	noti := notifier.NewMock()
	events := make(chan string, 1)

	commitID := "foobar"
	fluxEvent := event.Event{
		ID:         1,
		Type:       "sync",
		ServiceIDs: []resource.ID{},
		LogLevel:   "info",
		Message:    "",
		StartedAt:  time.Now(),
		EndedAt:    time.Now(),
		Metadata: &event.SyncEventMetadata{
			Commits: []event.Commit{
				{
					Revision: commitID,
					Message:  "",
				},
			},
		},
	}
	body, err := json.Marshal(fluxEvent)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	req, err := http.NewRequest("GET", "/v6/events", bytes.NewReader(body))
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	rr := httptest.NewRecorder()
	eventHandler(log, noti, events).ServeHTTP(rr, req)
	g.Expect(rr.Code).Should(gomega.Equal(http.StatusOK))

	g.Expect(events).Should(gomega.Receive(gomega.Equal(commitID)))
	g.Expect(noti.Events).Should(gomega.Receive(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
		"Type":     gomega.Equal(notifier.EventTypeSync),
		"CommitID": gomega.Equal(commitID),
		"State":    gomega.Equal(notifier.EventStateSucceeded),
	})))
	g.Consistently(events).ShouldNot(gomega.Receive())
	g.Consistently(noti.Events).ShouldNot(gomega.Receive())
}

func TestDisabledPoller(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	log := logr.TestLogger{T: t}
	noti := notifier.NewMock()

	fluxEvent := event.Event{
		ID:         1,
		Type:       "sync",
		ServiceIDs: []resource.ID{},
		LogLevel:   "info",
		Message:    "",
		StartedAt:  time.Now(),
		EndedAt:    time.Now(),
		Metadata: &event.SyncEventMetadata{
			Commits: []event.Commit{
				{
					Revision: "",
					Message:  "",
				},
			},
		},
	}
	body, err := json.Marshal(fluxEvent)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	for i := 0; i < 5; i++ {
		req, err := http.NewRequest("GET", "/v6/events", bytes.NewReader(body))
		g.Expect(err).ShouldNot(gomega.HaveOccurred())
		rr := httptest.NewRecorder()
		eventHandler(log, noti, nil).ServeHTTP(rr, req)
		g.Expect(rr.Code).Should(gomega.Equal(http.StatusOK))
	}
}

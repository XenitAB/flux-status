package poller

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/go-logr/logr"

	"github.com/xenitab/flux-status/pkg/exporter"
	"github.com/xenitab/flux-status/pkg/flux"
)

type Poller struct {
	Log      logr.Logger
	FluxPort int
	Interval int
	Timeout  int
}

func NewPoller(l logr.Logger, fp int, pi int, pt int) *Poller {
	return &Poller{
		Log:      l,
		FluxPort: fp,
		Interval: pi,
		Timeout:  pt,
	}
}

func (p *Poller) Poll(commitId string, e exporter.Exporter) error {
	p.Log.Info("Started polling", "commit-id", commitId)
	event := exporter.Event{
		Sender:   "services",
		Message:  "Polling service status",
		CommitId: commitId,
		State:    exporter.EventStatePending,
	}
	if err := e.Send(event); err != nil {
		return err
	}

	tick := time.Tick(time.Duration(p.Interval) * time.Second)
	timeout := time.After(time.Duration(p.Timeout) * time.Second)
	serviceUrl := "http://localhost:" + strconv.Itoa(p.FluxPort) + "/api/flux/v6/services"

	for {
		select {
		case <-timeout:
			event.Message = "Service health check timed out"
			event.State = exporter.EventStateFailed
			e.Send(event)
			return errors.New("Timed out poll")
		case <-tick:
			resp, err := http.Get(serviceUrl)
			if err != nil {
				p.Log.Error(err, "Failed to request failed flux services.")
				break
			}

			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				p.Log.Error(err, "Could not read response data.")
				break
			}

			services := &[]flux.Service{}
			if err := json.Unmarshal(body, services); err != nil {
				p.Log.Error(err, "Could not parse service response.")
				break
			}

			pendingServices := verifyServices(*services)
			if len(pendingServices) == 0 {
				p.Log.Info("All services are running")
				event.Message = "All Services are running"
				event.State = exporter.EventStateSucceeded
				err := e.Send(event)
				return err
			}

			p.Log.Info("Waiting for services to be ready", "services", pendingServices)
		}
	}
}

func (p *Poller) Stop() {

}

func verifyServices(ss []flux.Service) []string {
	result := []string{}
	for _, s := range ss {
		if s.ReadOnly == "NotInRepo" {
			continue
		}

		if s.Status != "ready" {
			result = append(result, s.Id)
		}
	}

	return result
}

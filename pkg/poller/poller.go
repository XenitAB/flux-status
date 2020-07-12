package poller

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/go-logr/logr"

	"github.com/xenitab/flux-status/pkg/exporter"
	"github.com/xenitab/flux-status/pkg/flux"
)

type Poller struct {
	Log        logr.Logger
	Interval   int
	Timeout    int
	ServiceUrl string

	stop chan bool
}

func NewPoller(l logr.Logger, fp int, pi int, pt int) *Poller {
	return &Poller{
		Log:        l,
		Interval:   pi,
		Timeout:    pt,
		ServiceUrl: "http://localhost:" + strconv.Itoa(fp) + "/api/flux/v6/services",
		stop:       make(chan bool),
	}
}

func (p *Poller) Poll(commitId string, e exporter.Exporter) error {
	p.Log.Info("Started polling", "commit-id", commitId)
	event := exporter.Event{
		Id:       "flux-status",
		Event:    "workload",
		Instance: "dev",
		Message:  "Polling workload status",
		CommitId: commitId,
		State:    exporter.EventStatePending,
	}
	if err := e.Send(event); err != nil {
		return err
	}

	var timeout <-chan time.Time
	if p.Timeout != 0 {
		timeout = time.After(time.Duration(p.Timeout) * time.Second)
	}
	tick := time.Tick(time.Duration(p.Interval) * time.Second)

	for {
		select {
		case <-p.stop:
			p.Log.Info("Poller stopped")
			if err := handleStop(e, event); err != nil {
				return err
			}
			return nil
		case <-timeout:
			p.Log.Info("Poller timed out")
			if err := handleTimeout(e, event); err != nil {
				return err
			}
			return errors.New("Timed Out")
		case <-tick:
			p.Log.Info("Poller tick")
			if err := handleTick(e, event, p.ServiceUrl); err != nil {
				p.Log.Error(err, "All workloads are not healthy")
				continue
			}
			p.Log.Info("All workloads are healthy")
			return nil
		}
	}
}

func (p *Poller) Stop() {
	p.Log.Info("Stopping poller")
	close(p.stop)
	p.stop = make(chan bool)
}

func handleStop(e exporter.Exporter, ev exporter.Event) error {
	ev.Message = "Workload check stopped"
	ev.State = exporter.EventStateCanceled
	return e.Send(ev)
}

func handleTimeout(e exporter.Exporter, ev exporter.Event) error {
	ev.Message = "Service health check timed out"
	ev.State = exporter.EventStateFailed
	return e.Send(ev)
}

func handleTick(e exporter.Exporter, ev exporter.Event, url string) error {
	ev.Message = "All Services are running"
	ev.State = exporter.EventStateSucceeded

	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	services := &[]flux.Service{}
	if err := json.Unmarshal(body, services); err != nil {
		return err
	}

	pendingServices := verifyServices(*services)
	if len(pendingServices) > 0 {
		return fmt.Errorf("Waiting for workloads: %v", pendingServices)
	}

	return e.Send(ev)
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

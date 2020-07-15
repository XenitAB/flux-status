package poller

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/fluxcd/flux/pkg/api/v6"
	transport "github.com/fluxcd/flux/pkg/http"
	"github.com/fluxcd/flux/pkg/http/client"
	"github.com/fluxcd/flux/pkg/resource"
	"github.com/go-logr/logr"

	"github.com/xenitab/flux-status/pkg/exporter"
)

type Poller struct {
	Log      logr.Logger
	Interval int
	Timeout  int
	Instance string

	client *client.Client
	stop   chan bool
}

func NewPoller(l logr.Logger, fAddr string, pi int, pt int, i string) *Poller {
	fluxUrl := "http://" + fAddr + "/api/flux"
	return &Poller{
		Log:      l,
		Interval: pi,
		Timeout:  pt,
		Instance: i,
		client:   client.New(http.DefaultClient, transport.NewAPIRouter(), fluxUrl, ""),
		stop:     make(chan bool),
	}
}

func (p *Poller) Poll(commitId string, exp exporter.Exporter) error {
	p.Log.Info("Started polling", "commit-id", commitId)

	baseEvent := exporter.Event{
		Id:       "flux-status",
		Event:    "workload",
		Instance: p.Instance,
		CommitId: commitId,
	}

	// Snap shot intitial workloads
	ctx := context.Background()
	workloads, err := p.client.ListServices(ctx, "")
	if err != nil {
		return err
	}
	snap := snapshotWorkloads(workloads)

	// Send pending event
	pendingEvent := baseEvent
	pendingEvent.State = exporter.EventStatePending
	pendingEvent.Message = "Waiting for workloads to be ready"
	if err := exp.Send(pendingEvent); err != nil {
		return err
	}

	// Poll workloads until timeout or ready
	var timeout <-chan time.Time
	if p.Timeout != 0 {
		timeout = time.After(time.Duration(p.Timeout) * time.Second)
	}
	tick := time.Tick(time.Duration(p.Interval) * time.Second)
	for {
		select {
		case <-p.stop:
			p.Log.Info("Poller stopped")
			if err := handleStop(exp, baseEvent); err != nil {
				return err
			}
			return nil
		case <-timeout:
			p.Log.Info("Poller timed out")
			if err := handleTimeout(exp, baseEvent); err != nil {
				return err
			}
			return errors.New("Timed Out")
		case <-tick:
			p.Log.Info("Poller tick")
			if err := handleTick(exp, p.client, baseEvent, snap); err != nil {
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

// snapshotWorkloads returns a list of resource ids created by flux
func snapshotWorkloads(ww []v6.ControllerStatus) resource.IDSet {
	result := resource.IDSet{}
	for _, w := range ww {
		if w.ReadOnly != "NotInRepo" {
			result.Add([]resource.ID{w.ID})
		}
	}

	return result
}

func handleTick(e exporter.Exporter, c *client.Client, ev exporter.Event, snap resource.IDSet) error {
	ctx := context.Background()
	workloads, err := c.ListServices(ctx, "")
	if err != nil {
		return err
	}

	// Make sure initial snapshot matches currently generated snapshot
	newSnap := snapshotWorkloads(workloads)
	if len(newSnap.Intersection(snap)) != len(snap) {
		return errors.New("Current workloads do not match workloads at sync")
	}

	pending := pendingWorkloads(workloads)
	if len(pending) > 0 {
		return fmt.Errorf("Pending: %v", pending)
	}

	ev.Message = "All workloads have started successfully"
	ev.State = exporter.EventStateSucceeded
	return e.Send(ev)
}

func handleStop(e exporter.Exporter, ev exporter.Event) error {
	ev.Message = "Workload polling stopped"
	ev.State = exporter.EventStateCanceled
	return e.Send(ev)
}

func handleTimeout(e exporter.Exporter, ev exporter.Event) error {
	ev.Message = "Workload polling timed out"
	ev.State = exporter.EventStateFailed
	return e.Send(ev)
}

// verifyServices returns any workload created by flux that is not ready
func pendingWorkloads(ww []v6.ControllerStatus) resource.IDSet {
	result := resource.IDSet{}
	for _, w := range ww {
		if w.ReadOnly == v6.ReadOnlyMissing {
			continue
		}

		if w.Status == "deployed" || w.Status == "ready" {
			continue
		}

		result.Add([]resource.ID{w.ID})
	}

	return result
}

package poller

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/fluxcd/flux/pkg/api/v6"
	transport "github.com/fluxcd/flux/pkg/http"
	"github.com/fluxcd/flux/pkg/http/client"
	"github.com/fluxcd/flux/pkg/resource"
	"github.com/go-logr/logr"

	"github.com/xenitab/flux-status/pkg/flux"
	"github.com/xenitab/flux-status/pkg/notifier"
)

type Poller struct {
	Log      logr.Logger
	Notifier notifier.Notifier
	Events   <-chan string
	Interval int
	Timeout  int
	Client   flux.Client

	wg   sync.WaitGroup
	quit chan struct{}
}

func NewPoller(l logr.Logger, n notifier.Notifier, e <-chan string, fAddr string, pi int, pt int) (*Poller, error) {
	fluxUrl, err := url.Parse(fmt.Sprintf("http://%v/api/flux", fAddr))
	if err != nil {
		return nil, err
	}

	return &Poller{
		Log:      l,
		Events:   e,
		Notifier: n,
		Interval: pi,
		Timeout:  pt,
		Client:   client.New(http.DefaultClient, transport.NewAPIRouter(), fluxUrl.String(), ""),

		wg:   sync.WaitGroup{},
		quit: make(chan struct{}),
	}, nil
}

// Starts the poller and waits for new events
func (p *Poller) Start() error {
	wg := sync.WaitGroup{}
	var pollCtx context.Context
	var pollCancel context.CancelFunc = func() {}
	for {
		select {
		case <-p.quit:
			pollCancel()
			return nil
		case commitId := <-p.Events:
			pollCancel()
			pollCtx, pollCancel = context.WithCancel(context.Background())
			wg.Add(1)
			go p.poll(pollCtx, &wg, commitId)
		}
	}
}

func (p *Poller) Stop(ctx context.Context) error {
	c := make(chan struct{})
	go func() {
		defer close(c)
		p.wg.Wait()
	}()

	close(p.quit)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-c:
			return nil
		}
	}
}

func (p *Poller) poll(ctx context.Context, wg *sync.WaitGroup, commitId string) error {
	defer wg.Done()
	log := p.Log.WithValues("commitId", commitId)
	log.Info("Received event")

	// Sed pending event
	p.Notifier.Send(ctx, notifier.Event{
		Type:     notifier.EventTypeWorkload,
		CommitId: commitId,
		State:    notifier.EventStatePending,
		Message:  "Waiting for workloads to be ready",
	})

	// Snap shot intitial workloads
	workloads, err := p.Client.ListServices(ctx, "")
	if err != nil {
		return err
	}
	snap := snapshotWorkloads(workloads)

	// Start polling workloads
	tickCh := time.NewTicker(time.Duration(p.Interval) * time.Second)
	timeoutCh := timeoutChannel(p.Timeout)
	for {
		select {
		case <-ctx.Done():
			log.Info("Poller stopped")
			tickCh.Stop()
			timeoutCh.Stop()
			return p.Notifier.Send(ctx, notifier.Event{
				Type:     notifier.EventTypeWorkload,
				CommitId: commitId,
				State:    notifier.EventStateCanceled,
				Message:  "Workload polling stopped",
			})
		case <-timeoutCh.C:
			log.Info("Poller timed out")
			tickCh.Stop()
			timeoutCh.Stop()
			return p.Notifier.Send(ctx, notifier.Event{
				Type:     notifier.EventTypeWorkload,
				CommitId: commitId,
				State:    notifier.EventStateFailed,
				Message:  "Workload polling timed out",
			})
		case <-tickCh.C:
			log.Info("Poller tick")

			// Make a new snapshot of the workload state
			newWorkloads, err := p.Client.ListServices(ctx, "")
			if err != nil {
				return err
			}
			newSnap := snapshotWorkloads(newWorkloads)

			// Make sure initial snapshot matches currently generated snapshot
			if len(newSnap.Intersection(snap)) != len(snap) {
				log.Info("Current workloads do not match workloads at sync")
				continue
			}

			// Check if there are any pending workloads
			pending := pendingWorkloads(workloads)
			if len(pending) > 0 {
				log.Info("Waiting for workloads to be healthy", "pending", pending)
				continue
			}

			// End poller as it has successfully completed
			log.Info("All workloads are healthy")
			p.Notifier.Send(ctx, notifier.Event{
				Type:     notifier.EventTypeWorkload,
				CommitId: commitId,
				State:    notifier.EventStateSucceeded,
				Message:  "All workloads have started successfully",
			})
			return nil
		}
	}
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

func timeoutChannel(timeout int) *time.Timer {
	timerCh := time.NewTimer(time.Duration(timeout) * time.Second)
	if timeout == 0 {
		timerCh.Stop()
	}

	return timerCh
}

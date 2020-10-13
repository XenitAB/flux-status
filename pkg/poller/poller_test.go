package poller

import (
	"context"
	"crypto/sha1"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	v6 "github.com/fluxcd/flux/pkg/api/v6"
	"github.com/fluxcd/flux/pkg/resource"
	logr "github.com/go-logr/logr/testing"
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/gstruct"
	"github.com/xenitab/flux-status/pkg/notifier"
	"go.uber.org/goleak"

	"github.com/xenitab/flux-status/pkg/flux"
)

func TestVerifyReadyDeployment(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	ww := []v6.ControllerStatus{
		{
			ID:       resource.MustParseID("namespace:deployment/resource-name"),
			Status:   "ready",
			ReadOnly: "ReadOnlyMode",
		},
	}
	res := pendingWorkloads(ww)
	g.Expect(res.String()).Should(gomega.Equal("{}"))
}

func TestVerifyNotReadyDeployment(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	ww := []v6.ControllerStatus{
		{
			ID:       resource.MustParseID("namespace:deployment/resource-name"),
			Status:   "updating",
			ReadOnly: "ReadOnlyMode",
		},
	}
	res := pendingWorkloads(ww)
	g.Expect(res.String()).Should(gomega.Equal("{namespace:deployment/resource-name}"))
}

func TestVerifyReadyHelmRelease(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	ww := []v6.ControllerStatus{
		{
			ID:       resource.MustParseID("namespace:helmrelease/resource-name"),
			Status:   "deployed",
			ReadOnly: "ReadOnlyMode",
		},
	}
	res := pendingWorkloads(ww)
	g.Expect(res.String()).Should(gomega.Equal("{}"))
}

func TestVerifyNotReadyHelmRelease(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	ww := []v6.ControllerStatus{
		{
			ID:       resource.MustParseID("namespace:helmrelease/resource-name"),
			Status:   "failed",
			ReadOnly: "ReadOnlyMode",
		},
	}
	res := pendingWorkloads(ww)
	g.Expect(res.String()).Should(gomega.Equal("{namespace:helmrelease/resource-name}"))
}

func TestVerifyHelmReleaseWithDeployment(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	ww := []v6.ControllerStatus{
		{
			ID:       resource.MustParseID("namespace:helmrelease/resource-name"),
			Status:   "deployed",
			ReadOnly: "ReadOnlyMode",
		},
		{
			ID:       resource.MustParseID("namespace:deployment/resource-name"),
			Status:   "ready",
			ReadOnly: "ReadOnlyMode",
		},
	}
	res := pendingWorkloads(ww)
	g.Expect(res.String()).Should(gomega.Equal("{}"))
}

func randHash() string {
	timestamp := time.Now().Unix()
	h := sha1.New()
	h.Write([]byte(strconv.FormatInt(timestamp, 10)))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func TestPollSuccessful(t *testing.T) {
	defer goleak.VerifyNone(t)
	g := gomega.NewGomegaWithT(t)

	log := logr.TestLogger{T: t}
	events := make(chan string)
	noti := notifier.NewMock()
	client := &flux.Mock{}

	poller := Poller{
		Log:      log,
		Notifier: noti,
		Events:   events,
		Interval: 3,
		Timeout:  10,
		Client:   client,
		wg:       sync.WaitGroup{},
		quit:     make(chan struct{}),
	}
	go poller.Start()

	for i := 0; i < 5; i++ {
		commitId := randHash()
		events <- commitId
		g.Eventually(noti.Events, 5).Should(gomega.Receive(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
			"Type":     gomega.Equal(notifier.EventTypeWorkload),
			"CommitId": gomega.Equal(commitId),
			"State":    gomega.Equal(notifier.EventStateSucceeded),
		})))
		g.Consistently(noti.Events).ShouldNot(gomega.Receive())
	}

	poller.Stop(context.TODO())
}

func TestPollTimeout(t *testing.T) {
	defer goleak.VerifyNone(t)
	g := gomega.NewGomegaWithT(t)

	log := logr.TestLogger{T: t}
	events := make(chan string)
	noti := notifier.NewMock()
	client := &flux.Mock{}

	poller := Poller{
		Log:      log,
		Notifier: noti,
		Events:   events,
		Interval: 3,
		Timeout:  10,
		Client:   client,
		wg:       sync.WaitGroup{},
		quit:     make(chan struct{}),
	}
	go poller.Start()

	client.Services = []v6.ControllerStatus{
		{
			ID:       resource.MustParseID("namespace:helmrelease/resource-name"),
			Status:   "failed",
			ReadOnly: "ReadOnlyMode",
		},
	}

	commitId := randHash()
	events <- commitId
	g.Eventually(noti.Events, 12).Should(gomega.Receive(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
		"Type":     gomega.Equal(notifier.EventTypeWorkload),
		"CommitId": gomega.Equal(commitId),
		"State":    gomega.Equal(notifier.EventStateFailed),
	})))
	g.Consistently(noti.Events).ShouldNot(gomega.Receive())

	poller.Stop(context.TODO())
}

func TestPollCancel(t *testing.T) {
	defer goleak.VerifyNone(t)
	g := gomega.NewGomegaWithT(t)

	log := logr.TestLogger{T: t}
	events := make(chan string)
	noti := notifier.NewMock()
	client := &flux.Mock{}

	poller := Poller{
		Log:      log,
		Notifier: noti,
		Events:   events,
		Interval: 1,
		Timeout:  10,
		Client:   client,
		wg:       sync.WaitGroup{},
		quit:     make(chan struct{}),
	}
	go poller.Start()

	client.Services = []v6.ControllerStatus{
		{
			ID:       resource.MustParseID("namespace:deployment/resource-name"),
			Status:   "ready",
			ReadOnly: "ReadOnlyMode",
		},
	}

	firstCommitId := randHash()
	events <- firstCommitId
	secondCommitId := randHash()
	events <- secondCommitId
	g.Eventually(noti.Events, 5).Should(gomega.Receive(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
		"Type":     gomega.Equal(notifier.EventTypeWorkload),
		"CommitId": gomega.Equal(firstCommitId),
		"State":    gomega.Equal(notifier.EventStateCanceled),
	})))
	g.Eventually(noti.Events, 5).Should(gomega.Receive(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
		"Type":     gomega.Equal(notifier.EventTypeWorkload),
		"CommitId": gomega.Equal(secondCommitId),
		"State":    gomega.Equal(notifier.EventStateSucceeded),
	})))
	g.Consistently(noti.Events).ShouldNot(gomega.Receive())

	poller.Stop(context.TODO())
}

func TestPollShutdown(t *testing.T) {
	defer goleak.VerifyNone(t)

	log := logr.TestLogger{T: t}
	events := make(chan string)
	noti := notifier.NewMock()
	client := &flux.Mock{}

	poller := Poller{
		Log:      log,
		Notifier: noti,
		Events:   events,
		Interval: 1,
		Timeout:  0,
		Client:   client,
		wg:       sync.WaitGroup{},
		quit:     make(chan struct{}),
	}
	go poller.Start()

	client.Services = []v6.ControllerStatus{
		{
			ID:       resource.MustParseID("namespace:helmrelease/resource-name"),
			Status:   "failed",
			ReadOnly: "ReadOnlyMode",
		},
	}

	events <- randHash()
	poller.Stop(context.TODO())
}

package poller

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"os"
	"sync"
	"testing"

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
	data := make([]byte, 10)
	_, err := rand.Read(data)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	commitID := fmt.Sprintf("%x", sha256.Sum256(data))
	return commitID
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
		commitID := randHash()
		events <- commitID
		g.Eventually(noti.Events, 5).Should(gomega.Receive(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
			"Type":     gomega.Equal(notifier.EventTypeWorkload),
			"CommitID": gomega.Equal(commitID),
			"State":    gomega.Equal(notifier.EventStatePending),
		})))
		g.Eventually(noti.Events, 5).Should(gomega.Receive(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
			"Type":     gomega.Equal(notifier.EventTypeWorkload),
			"CommitID": gomega.Equal(commitID),
			"State":    gomega.Equal(notifier.EventStateSucceeded),
		})))
		g.Consistently(noti.Events).ShouldNot(gomega.Receive())
	}

	err := poller.Stop(context.TODO())
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
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

	commitID := randHash()
	events <- commitID
	g.Eventually(noti.Events, 5).Should(gomega.Receive(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
		"Type":     gomega.Equal(notifier.EventTypeWorkload),
		"CommitID": gomega.Equal(commitID),
		"State":    gomega.Equal(notifier.EventStatePending),
	})))
	g.Eventually(noti.Events, 12).Should(gomega.Receive(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
		"Type":     gomega.Equal(notifier.EventTypeWorkload),
		"CommitID": gomega.Equal(commitID),
		"State":    gomega.Equal(notifier.EventStateFailed),
	})))
	g.Consistently(noti.Events).ShouldNot(gomega.Receive())

	err := poller.Stop(context.TODO())
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
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

	firstCommitID := randHash()
	events <- firstCommitID
	g.Eventually(noti.Events, 5).Should(gomega.Receive(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
		"Type":     gomega.Equal(notifier.EventTypeWorkload),
		"CommitID": gomega.Equal(firstCommitID),
		"State":    gomega.Equal(notifier.EventStatePending),
	})))
	secondCommitID := randHash()
	events <- secondCommitID
	g.Eventually(noti.Events, 5).Should(gomega.Receive(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
		"Type":     gomega.Equal(notifier.EventTypeWorkload),
		"CommitID": gomega.Equal(secondCommitID),
		"State":    gomega.Equal(notifier.EventStatePending),
	})))
	g.Eventually(noti.Events, 5).Should(gomega.Receive(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
		"Type":     gomega.Equal(notifier.EventTypeWorkload),
		"CommitID": gomega.Equal(firstCommitID),
		"State":    gomega.Equal(notifier.EventStateCanceled),
	})))
	g.Eventually(noti.Events, 5).Should(gomega.Receive(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
		"Type":     gomega.Equal(notifier.EventTypeWorkload),
		"CommitID": gomega.Equal(secondCommitID),
		"State":    gomega.Equal(notifier.EventStateSucceeded),
	})))
	g.Consistently(noti.Events).ShouldNot(gomega.Receive())

	err := poller.Stop(context.TODO())
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
}

func TestPollShutdown(t *testing.T) {
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
	g.Eventually(noti.Events, 5).Should(gomega.Receive())
	events <- randHash()
	g.Eventually(noti.Events, 5).Should(gomega.Receive())
	events <- randHash()
	g.Eventually(noti.Events, 5).Should(gomega.Receive())

	err := poller.Stop(context.TODO())
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
}

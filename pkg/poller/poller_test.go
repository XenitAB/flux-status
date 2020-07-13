package poller

import (
	"testing"

	v6 "github.com/fluxcd/flux/pkg/api/v6"
	"github.com/fluxcd/flux/pkg/resource"
	"github.com/onsi/gomega"
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

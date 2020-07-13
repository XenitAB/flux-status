package poller

import (
	"testing"

	"github.com/onsi/gomega"
	"github.com/xenitab/flux-status/pkg/flux"
)

func TestVerifyReadyDeployment(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	ss := []flux.Service{
		{
			Id:       "namespace:deployment/resource-name",
			Status:   "ready",
			ReadOnly: "ReadOnlyMode",
		},
	}
	res := verifyServices(ss)
	g.Expect(res).Should(gomega.Equal([]string{}))
}

func TestVerifyNotReadyDeployment(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	ss := []flux.Service{
		{
			Id:       "namespace:deployment/resource-name",
			Status:   "updating",
			ReadOnly: "ReadOnlyMode",
		},
	}
	res := verifyServices(ss)
	g.Expect(res).Should(gomega.Equal([]string{"namespace:deployment/resource-name"}))
}

func TestVerifyReadyHelmRelease(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	ss := []flux.Service{
		{
			Id:       "namespace:helmrelease/resource-name",
			Status:   "deployed",
			ReadOnly: "ReadOnlyMode",
		},
	}
	res := verifyServices(ss)
	g.Expect(res).Should(gomega.Equal([]string{}))
}

func TestVerifyNotReadyHelmRelease(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	ss := []flux.Service{
		{
			Id:       "namespace:helmrelease/resource-name",
			Status:   "failed",
			ReadOnly: "ReadOnlyMode",
		},
	}
	res := verifyServices(ss)
	g.Expect(res).Should(gomega.Equal([]string{"namespace:helmrelease/resource-name"}))
}

func TestVerifyHelmReleaseWithDeployment(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	ss := []flux.Service{
		{
			Id:       "namespace:helmrelease/resource-name",
			Status:   "deployed",
			ReadOnly: "ReadOnlyMode",
		},
		{
			Id:       "namespace:deployment/resource-name",
			Status:   "ready",
			ReadOnly: "NotInRepo",
		},
	}
	res := verifyServices(ss)
	g.Expect(res).Should(gomega.Equal([]string{}))
}

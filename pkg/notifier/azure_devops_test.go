package notifier

import (
	"testing"

	"github.com/onsi/gomega"
)

func TestParseAzdoUrlHttps(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	s := "https://foobar@dev.azure.com/org/proj/_git/repo"
	c, err := parseAzdoUrl(s)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(c.orgUrl).Should(gomega.Equal("https://foobar@dev.azure.com/org"))
	g.Expect(c.projectId).Should(gomega.Equal("proj"))
	g.Expect(c.repositoryId).Should(gomega.Equal("repo"))
}

func TestParseAzdoUrlSsh(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	s := "ssh://ssh.dev.azure.com/v3/org/proj/repo"
	c, err := parseAzdoUrl(s)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(c.orgUrl).Should(gomega.Equal("https://dev.azure.com/org"))
	g.Expect(c.projectId).Should(gomega.Equal("proj"))
	g.Expect(c.repositoryId).Should(gomega.Equal("repo"))
}

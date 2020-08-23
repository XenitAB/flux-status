package notifier

import (
	"testing"

	"github.com/onsi/gomega"
)

func TestParseAzdoURLHttps(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	s := "https://foobar@dev.azure.com/org/proj/_git/repo"
	c, err := parseAzdoURL(s)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(c.orgURL).Should(gomega.Equal("https://foobar@dev.azure.com/org"))
	g.Expect(c.projectID).Should(gomega.Equal("proj"))
	g.Expect(c.repositoryID).Should(gomega.Equal("repo"))
}

func TestParseAzdoURLSsh(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	s := "ssh://ssh.dev.azure.com/v3/org/proj/repo"
	c, err := parseAzdoURL(s)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(c.orgURL).Should(gomega.Equal("https://dev.azure.com/org"))
	g.Expect(c.projectID).Should(gomega.Equal("proj"))
	g.Expect(c.repositoryID).Should(gomega.Equal("repo"))
}

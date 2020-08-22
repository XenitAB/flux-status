package notifier

import (
	"testing"

	"github.com/onsi/gomega"
)

func TestParseGitlabUrlHttps(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	s := "https://gitlab.com/group/name.git"
	id, bUrl, err := parseGitlabUrl(s)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(id).Should(gomega.Equal("group/name"))
	g.Expect(bUrl).Should(gomega.Equal(""))
}

func TestParseGitlabUrlSsh(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	s := "ssh://git@gitlab.com/group/name.git"
	id, bUrl, err := parseGitlabUrl(s)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(id).Should(gomega.Equal("group/name"))
	g.Expect(bUrl).Should(gomega.Equal(""))
}

func TestParseGitlabUrlSubGroups(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	s := "https://gitlab.com/group1/group2/name.git"
	id, bUrl, err := parseGitlabUrl(s)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(id).Should(gomega.Equal("group1/group2/name"))
	g.Expect(bUrl).Should(gomega.Equal(""))
}

func TestParseGitlabUrlSelfHosted(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	s := "https://selfhostedgitlab.com/group/name.git"
	id, bUrl, err := parseGitlabUrl(s)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(id).Should(gomega.Equal("group/name"))
	g.Expect(bUrl).Should(gomega.Equal("https://selfhostedgitlab.com"))
}

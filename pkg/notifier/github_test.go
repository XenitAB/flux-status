package notifier

import (
	"testing"

	"github.com/onsi/gomega"
)

func TestParseGithubURLHttps(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	s := "https://github.com/group/name.git"
	owner, repo, err := parseGitHubURL(s)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(owner).Should(gomega.Equal("group"))
	g.Expect(repo).Should(gomega.Equal("name"))
}

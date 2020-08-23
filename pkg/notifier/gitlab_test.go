package notifier

import (
	"testing"

	"github.com/onsi/gomega"
)

func TestParseHttpURL(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	s := "https://gitlab.com/namespace/name.git"
	url, err := parseGitlabURL(s)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(url).Should(gomega.Equal("namespace/name"))
}

package exporter

import (
	"fmt"
	"testing"
)

func TestParseAzdoGitUrlHttps(t *testing.T) {
	s := "https://foobar@dev.azure.com/org/proj/_git/repo"
	c, err := parseAzdoGitUrl(s)
	if err != nil {
		t.Error(err)
		return
	}

	if c.orgUrl != "https://foobar@dev.azure.com/org" {
		fmt.Println(c.orgUrl)
		t.Fail()
		return
	}

	if c.projectId != "proj" {
		fmt.Println(c.projectId)
		t.Fail()
		return
	}

	if c.repositoryId != "repo" {
		fmt.Println(c.repositoryId)
		t.Fail()
		return
	}
}

func TestParseAzdoGitUrlSsh(t *testing.T) {
	s := "ssh://ssh.dev.azure.com/v3/org/proj/repo"
	c, err := parseAzdoGitUrl(s)
	if err != nil {
		t.Error(err)
		return
	}

	if c.orgUrl != "https://dev.azure.com/org" {
		fmt.Println(c.orgUrl)
		t.Fail()
		return
	}

	if c.projectId != "proj" {
		fmt.Println(c.projectId)
		t.Fail()
		return
	}

	if c.repositoryId != "repo" {
		fmt.Println(c.repositoryId)
		t.Fail()
		return
	}

}

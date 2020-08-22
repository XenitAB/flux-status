package notifier

import (
	"fmt"
	"testing"
)

func TestParseAzdoUrlHttps(t *testing.T) {
	s := "https://foobar@dev.azure.com/org/proj/_git/repo"
	c, err := parseAzdoUrl(s)
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

func TestParseAzdoUrlSsh(t *testing.T) {
	s := "ssh://ssh.dev.azure.com/v3/org/proj/repo"
	c, err := parseAzdoUrl(s)
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

package notifier

import (
	"errors"
	"net/url"
	"strings"

	"github.com/xanzy/go-gitlab"
)

type Gitlab struct {
	Id     string
	client *gitlab.Client
}

func NewGitlab(token string, url string) (*Gitlab, error) {
	if len(token) == 0 {
		return nil, errors.New("Gitlab token can't be empty")
	}

	id, err := parseGitlabUrl(url)
	if err != nil {
		return nil, err
	}

	client, err := gitlab.NewClient(token)
	if err != nil {
		return nil, err
	}

	gitlab := &Gitlab{
		Id:     id,
		client: client,
	}

	return gitlab, nil
}

func (g Gitlab) Send(e Event) error {
	name := e.Id + "/" + e.Event + "/" + e.Instance
	options := &gitlab.SetCommitStatusOptions{
		State:       gitlabState(e.State),
		Description: &e.Message,
		Name:        &name,
	}

	_, _, err := g.client.Commits.SetCommitStatus(g.Id, e.CommitId, options)
	if err != nil {
		return err
	}

	return nil
}

func (g Gitlab) Get(commitId string, instance string) (*Status, error) {
  return nil, nil
}

func (g Gitlab) String() string {
	return "Gitlab" + " " + g.Id
}

func gitlabState(s EventState) gitlab.BuildStateValue {
	switch s {
	case EventStateFailed:
		return gitlab.Failed
	case EventStatePending:
		return gitlab.Running
	case EventStateSucceeded:
		return gitlab.Success
	case EventStateCanceled:
		return gitlab.Canceled
	default:
		return gitlab.Pending
	}
}

func parseGitlabUrl(s string) (string, error) {
	u, err := url.Parse(s)
	if err != nil {
		return "", err
	}

	comp := strings.Split(u.Path, "/")
	id := comp[1] + "/" + strings.TrimSuffix(comp[2], ".git")
	return id, nil
}

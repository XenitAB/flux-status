package notifier

import (
	"errors"
	"net/url"
	"strings"

	"github.com/xanzy/go-gitlab"
)

type Gitlab struct {
	instance string
	id       string
	client   *gitlab.Client
}

func NewGitlab(inst string, url string, token string) (*Gitlab, error) {
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
		instance: inst,
		id:       id,
		client:   client,
	}

	return gitlab, nil
}

func (g Gitlab) Send(e Event) error {
	name := StatusId + "/" + e.Event + "/" + g.instance
	options := &gitlab.SetCommitStatusOptions{
		State:       gitlabState(e.State),
		Description: &e.Message,
		Name:        &name,
	}

	_, _, err := g.client.Commits.SetCommitStatus(g.id, e.CommitId, options)
	if err != nil {
		return err
	}

	return nil
}

func (g Gitlab) Get(commitId string) (*Status, error) {
	return nil, nil
}

func (g Gitlab) String() string {
	return "Gitlab" + " " + g.id
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

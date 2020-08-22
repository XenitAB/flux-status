package notifier

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
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

	id, bUrl, err := parseGitlabUrl(url)
	if err != nil {
		return nil, err
	}

	gitlabClient := func() (*gitlab.Client, error) {
		if len(bUrl) == 0 {
			return gitlab.NewClient(token)
		}

		return gitlab.NewClient(token, gitlab.WithBaseURL(bUrl))
	}

	client, err := gitlabClient()
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

func (g Gitlab) Send(ctx context.Context, e Event) error {
	name := fmt.Sprintf("%v/%v/%v", StatusId, g.instance, e.Type)
	options := &gitlab.SetCommitStatusOptions{
		State:       toGitlabState(e.State),
		Description: &e.Message,
		Name:        &name,
	}

	_, _, err := g.client.Commits.SetCommitStatus(g.id, e.CommitId, options)
	if err != nil {
		return err
	}

	return nil
}

func (g Gitlab) Get(commitId string, action string) (*Status, error) {
	opts := gitlab.GetCommitStatusesOptions{
		All: gitlab.Bool(true),
	}
	statuses, _, err := g.client.Commits.GetCommitStatuses(g.id, commitId, &opts)
	if err != nil {
		return nil, err
	}

	name := StatusId + "/" + action + "/" + g.instance
	for _, status := range statuses {
		if status.Name != name {
			continue
		}

		gitlabState := gitlab.BuildStateValue(status.Status)
		return &Status{
			Name:  status.Name,
			State: fromGitlabState(gitlabState),
		}, nil
	}

	return nil, errors.New("No status found")
}

func (g Gitlab) String() string {
	return "Gitlab" + " " + g.id
}

func toGitlabState(s EventState) gitlab.BuildStateValue {
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

func fromGitlabState(s gitlab.BuildStateValue) EventState {
	switch s {
	case gitlab.Pending, gitlab.Created, gitlab.Running, gitlab.Manual:
		return EventStatePending
	case gitlab.Success:
		return EventStateSucceeded
	case gitlab.Failed:
		return EventStateFailed
	case gitlab.Canceled, gitlab.Skipped:
		return EventStateCanceled
	default:
		return EventStateFailed
	}
}

func parseGitlabUrl(s string) (string, string, error) {
	u, err := url.Parse(s)
	if err != nil {
		return "", "", err
	}

	id := strings.TrimSuffix(u.Path, filepath.Ext(u.Path))
	// remove '/' prefix
	id = id[1:]

	basePath := ""
	if u.Hostname() != "gitlab.com" {
		basePath = fmt.Sprintf("%s://%s", u.Scheme, u.Host)
	}

	return id, basePath, nil
}

package notifier

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"
)

type GitHub struct {
	Instance   string
	Owner      string
	Repository string
	Client     *github.Client
}

func NewGitHub(inst string, url string, token string) (*GitHub, error) {
	if len(token) == 0 {
		return nil, errors.New("GitHub token can't be empty")
	}

	owner, repo, err := parseGitHubUrl(url)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	return &GitHub{
		Instance:   inst,
		Owner:      owner,
		Repository: repo,
		Client:     client,
	}, nil
}

func (g GitHub) Send(ctx context.Context, e Event) error {
	state, err := toGitHubState(e.State)
	if err != nil {
		return err
	}

	githubContext := fmt.Sprintf("%v/%v/%v", StatusId, g.Instance, e.Type)
	status := &github.RepoStatus{
		State:       &state,
		Description: &e.Message,
		Context:     &githubContext,
	}

	_, _, err = g.Client.Repositories.CreateStatus(ctx, g.Owner, g.Repository, e.CommitId, status)
	if err != nil {
		return err
	}

	return nil
}

func (g GitHub) Get(commitId string, action string) (*Status, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	statuses, _, err := g.Client.Repositories.ListStatuses(ctx, g.Owner, g.Repository, commitId, nil)
	if err != nil {
		return nil, err
	}

	for _, status := range statuses {
		comp := strings.Split(*status.Context, "/")

		if len(comp) < 3 {
			continue
		}

		if comp[1] != g.Instance && comp[2] != action {
			continue
		}

		state, err := fromGitHubState(*status.State)
		if err != nil {
			return nil, err
		}

		return &Status{
			Name:  *status.Context,
			State: state,
		}, nil
	}

	return nil, errors.New("No status found")
}

func (g GitHub) String() string {
	return "GitHub"
}

func parseGitHubUrl(urlStr string) (string, string, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", "", nil
	}

	comp := strings.Split(u.Path, "/")
	if len(comp) < 3 {
		return "", "", fmt.Errorf("Not enough components in path %v", u.Path)
	}

	owner := comp[1]
	repo := strings.TrimSuffix(comp[2], filepath.Ext(comp[2]))
	return owner, repo, nil
}

func toGitHubState(s EventState) (string, error) {
	switch s {
	case EventStateFailed:
		return "failure", nil
	case EventStatePending:
		return "pending", nil
	case EventStateSucceeded:
		return "success", nil
	case EventStateCanceled:
		return "error", nil
	default:
		return "", errors.New("Failed converting to GitHub state")
	}
}

func fromGitHubState(s string) (EventState, error) {
	switch s {
	case "failure":
		return EventStateFailed, nil
	case "error":
		return EventStateFailed, nil
	case "pending":
		return EventStatePending, nil
	case "success":
		return EventStateSucceeded, nil
	default:
		return "", errors.New("Failed converting to EventState")
	}
}

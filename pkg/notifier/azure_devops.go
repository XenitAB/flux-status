package notifier

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/microsoft/azure-devops-go-api/azuredevops"
	"github.com/microsoft/azure-devops-go-api/azuredevops/git"
)

// AzureDevops handles events for AzureDevops repositories.
type AzureDevops struct {
	instance     string
	client       git.Client
	repositoryID string
	projectID    string
}

// NewAzureDevops creates and returns an AzureDevops instance.
func NewAzureDevops(inst string, url string, pat string) (*AzureDevops, error) {
	azdoConfig, err := parseAzdoURL(url)
	if err != nil {
		return nil, err
	}

	var connection *azuredevops.Connection
	if len(pat) == 0 {
		connection = azuredevops.NewAnonymousConnection(azdoConfig.orgURL)
	} else {
		connection = azuredevops.NewPatConnection(azdoConfig.orgURL, pat)
	}

	client := connection.GetClientByUrl(azdoConfig.orgURL)
	gitClient := &git.ClientImpl{
		Client: *client,
	}

	azdo := &AzureDevops{
		instance:     inst,
		client:       gitClient,
		projectID:    azdoConfig.projectID,
		repositoryID: azdoConfig.repositoryID,
	}

	return azdo, nil
}

// Send sets the status for a given commit id in a AzureDevops repository.
func (azdo AzureDevops) Send(ctx context.Context, e Event) error {
	genre := StatusID
	name := fmt.Sprintf("%v/%v", azdo.instance, e.Type)
	state := toAzdoState(e.State)

	args := git.CreateCommitStatusArgs{
		Project:      &azdo.projectID,
		RepositoryId: &azdo.repositoryID,
		CommitId:     &e.CommitID,
		GitCommitStatusToCreate: &git.GitStatus{
			Description: &e.Message,
			State:       &state,
			Context: &git.GitStatusContext{
				Genre: &genre,
				Name:  &name,
			},
		},
	}
	_, err := azdo.client.CreateCommitStatus(ctx, args)
	if err != nil {
		return err
	}

	return nil
}

// Get returns the status of a given commit id in a AzureDevops repository.
func (azdo AzureDevops) Get(commitID string, action string) (*Status, error) {
	ctx := context.Background()
	args := git.GetStatusesArgs{
		Project:      &azdo.projectID,
		RepositoryId: &azdo.repositoryID,
		CommitId:     &commitID,
	}
	statuses, err := azdo.client.GetStatuses(ctx, args)
	if err != nil {
		return nil, err
	}

	name := azdo.instance + "/" + action
	for _, status := range *statuses {
		if !(*status.Context.Genre == StatusID && *status.Context.Name == name) {
			continue
		}

		return &Status{
			Name:  *status.Context.Genre + "/" + *status.Context.Name,
			State: fromAzdoState(*status.State),
		}, nil
	}

	return nil, errors.New("No status found")
}

// String returns the name of the struct.
func (AzureDevops) String() string {
	return "Azure DevOps"
}

// gitStatus returns the correct git status based on the success state.
func toAzdoState(s EventState) git.GitStatusState {
	switch s {
	case EventStateFailed:
		return git.GitStatusStateValues.Error
	case EventStatePending:
		return git.GitStatusStateValues.Pending
	case EventStateSucceeded:
		return git.GitStatusStateValues.Succeeded
	default:
		return git.GitStatusStateValues.NotSet
	}
}

func fromAzdoState(s git.GitStatusState) EventState {
	switch s {
	case git.GitStatusStateValues.Failed:
		return EventStateFailed
	case git.GitStatusStateValues.Error:
		return EventStateFailed
	case git.GitStatusStateValues.Pending:
		return EventStatePending
	case git.GitStatusStateValues.Succeeded:
		return EventStateSucceeded
	case git.GitStatusStateValues.NotSet:
		return EventStateFailed
	case git.GitStatusStateValues.NotApplicable:
		return EventStateFailed
	default:
		return EventStateFailed
	}
}

type azdoConfig struct {
	orgURL       string
	projectID    string
	repositoryID string
}

func parseAzdoURL(s string) (*azdoConfig, error) {
	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}

	components := strings.Split(u.Path, "/")

	if u.Scheme == "https" || u.Scheme == "http" {
		orgURL := u.Scheme + "://" + u.User.String() + "@" + u.Host + "/" + components[1]
		projectID := components[2]
		repositoryID := components[4]

		return &azdoConfig{
			orgURL:       orgURL,
			projectID:    projectID,
			repositoryID: repositoryID,
		}, nil
	} else if u.Scheme == "ssh" {
		host := strings.TrimPrefix(u.Host, "ssh.")
		orgURL := "https://" + host + "/" + components[2]
		projectID := components[3]
		repositoryID := components[4]

		return &azdoConfig{
			orgURL:       orgURL,
			projectID:    projectID,
			repositoryID: repositoryID,
		}, nil
	}

	return nil, fmt.Errorf("Unsuported schema: %v", u.Scheme)
}

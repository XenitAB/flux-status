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

type AzureDevops struct {
	instance     string
	client       git.Client
	repositoryId string
	projectId    string
}

func NewAzureDevops(inst string, url string, pat string) (*AzureDevops, error) {
	azdoConfig, err := parseAzdoUrl(url)
	if err != nil {
		return nil, err
	}

	var connection *azuredevops.Connection
	if len(pat) == 0 {
		connection = azuredevops.NewAnonymousConnection(azdoConfig.orgUrl)
	} else {
		connection = azuredevops.NewPatConnection(azdoConfig.orgUrl, pat)
	}

	client := connection.GetClientByUrl(azdoConfig.orgUrl)
	gitClient := &git.ClientImpl{
		Client: *client,
	}

	azdo := &AzureDevops{
		instance:     inst,
		client:       gitClient,
		projectId:    azdoConfig.projectId,
		repositoryId: azdoConfig.repositoryId,
	}

	return azdo, nil
}

// Send updates a given commit in AzureDevops with a status.
func (azdo AzureDevops) Send(e Event) error {
	genre := StatusId
	name := azdo.instance + "/" + e.Event
	state := toAzdoState(e.State)

	ctx := context.Background()
	args := git.CreateCommitStatusArgs{
		Project:      &azdo.projectId,
		RepositoryId: &azdo.repositoryId,
		CommitId:     &e.CommitId,
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

func (azdo AzureDevops) Get(commitId string) (*Status, error) {
	ctx := context.Background()
	args := git.GetStatusesArgs{
		Project:      &azdo.projectId,
		RepositoryId: &azdo.repositoryId,
		CommitId:     &commitId,
	}
	statuses, err := azdo.client.GetStatuses(ctx, args)
	if err != nil {
		return nil, err
	}

	for _, status := range *statuses {
		if *status.Context.Genre != StatusId && *status.Context.Name != azdo.instance+"workload" {
			continue
		}

		return &Status{
			Name:  *status.Context.Genre + "/" + *status.Context.Name,
			State: fromAzdoState(*status.State),
		}, nil
	}

	return nil, errors.New("Could not find a matching status")
}

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
	orgUrl       string
	projectId    string
	repositoryId string
}

func parseAzdoUrl(s string) (*azdoConfig, error) {
	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}

	components := strings.Split(u.Path, "/")

	if u.Scheme == "https" || u.Scheme == "http" {
		orgUrl := u.Scheme + "://" + u.User.String() + "@" + u.Host + "/" + components[1]
		projectId := components[2]
		repositoryId := components[4]

		return &azdoConfig{
			orgUrl:       orgUrl,
			projectId:    projectId,
			repositoryId: repositoryId,
		}, nil
	} else if u.Scheme == "ssh" {
		host := strings.TrimPrefix(u.Host, "ssh.")
		orgUrl := "https://" + host + "/" + components[2]
		projectId := components[3]
		repositoryId := components[4]

		return &azdoConfig{
			orgUrl:       orgUrl,
			projectId:    projectId,
			repositoryId: repositoryId,
		}, nil
	}

	return nil, fmt.Errorf("Unsuported schema: %v", u.Scheme)
}

package exporter

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/microsoft/azure-devops-go-api/azuredevops"
	"github.com/microsoft/azure-devops-go-api/azuredevops/git"
)

type AzureDevops struct {
	client       git.Client
	repositoryId string
	projectId    string
}

func NewAzureDevops(pat string, gitUrl string) (*AzureDevops, error) {
	azdoConfig, err := parseAzdoGitUrl(gitUrl)
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
		client:       gitClient,
		projectId:    azdoConfig.projectId,
		repositoryId: azdoConfig.repositoryId,
	}

	return azdo, nil
}

// Send updates a given commit in AzureDevops with a status.
func (azdo AzureDevops) Send(e Event) error {
	ctx := context.Background()
	args := git.CreateCommitStatusArgs{
		Project:      &azdo.projectId,
		RepositoryId: &azdo.repositoryId,
		CommitId:     &e.CommitId,
		GitCommitStatusToCreate: &git.GitStatus{
			Description: &e.Message,
			State:       gitStatus(e.State),
			Context: &git.GitStatusContext{
				Genre: stringPointer("flux-status"),
				Name:  &e.Sender,
			},
		},
	}
	_, err := azdo.client.CreateCommitStatus(ctx, args)
	if err != nil {
		return err
	}

	return nil
}

func stringPointer(v string) *string {
	return &v
}

// gitStatus returns the correct git status based on the success state.
func gitStatus(s EventState) *git.GitStatusState {
	switch s {
	case EventStateFailed:
		return &git.GitStatusStateValues.Error
	case EventStatePending:
		return &git.GitStatusStateValues.Pending
	case EventStateSucceeded:
		return &git.GitStatusStateValues.Succeeded
	default:
		return &git.GitStatusStateValues.NotSet
	}
}

type azdoConfig struct {
	orgUrl       string
	projectId    string
	repositoryId string
}

func parseAzdoGitUrl(s string) (*azdoConfig, error) {
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

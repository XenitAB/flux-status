package main

import (
	"encoding/json"
	"fmt"
	"os"

	flag "github.com/spf13/pflag"

	"github.com/xenitab/flux-status/pkg/notifier"
)

func main() {
	_ = flag.Bool("debug", false, "Enables debug mode.")
	commitId := flag.String("commit-id", "", "Id of commit to get status for.")
	instance := flag.String("instance", "default", "Id to differentiate between multiple flux-status updating the same repository.")
	gitUrl := flag.String("git-url", "", "URL for git repository, should be same as flux.")
	azdoPat := flag.String("azdo-pat", "", "Tokent to authenticate with Azure DevOps.")
	gitlabToken := flag.String("gitlab-token", "", "Token to authenticate with Gitlab.")
	flag.Parse()

	notifier, err := notifier.GetNotifier(*gitUrl, *azdoPat, *gitlabToken)
	if err != nil {
		fmt.Println("Could not create notifier")
		os.Exit(1)
	}

	status, err := notifier.Get(*commitId, *instance)
	if err != nil {
		fmt.Printf("Could not get status: %v", err)
		os.Exit(1)
	}

	b, err := json.Marshal(status)
	if err != nil {
		fmt.Println("Could not marshal json")
		os.Exit(1)
	}

	fmt.Println(string(b))
}

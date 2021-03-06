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
	commitID := flag.String("commit-id", "", "Id of commit to get status for.")
	instance := flag.String("instance", "default", "Id to differentiate between multiple flux-status updating the same repository.")
	action := flag.String("action", "workload", "Action to get status for, either sync or workload.")
	gitURL := flag.String("git-url", "", "URL for git repository, should be same as flux.")
	azdoPat := flag.String("azdo-pat", "", "Tokent to authenticate with Azure DevOps.")
	glToken := flag.String("gitlab-token", "", "Token to authenticate with Gitlab.")
	ghToken := flag.String("github-token", "", "Token to authenticate with GitHub.")
	flag.Parse()

	notifier, err := notifier.GetNotifier(*instance, *gitURL, *azdoPat, *glToken, *ghToken)
	if err != nil {
		fmt.Println("Could not create notifier")
		os.Exit(1)
	}

	status, err := notifier.Get(*commitID, *action)
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

# Flux Status
[![Go Report Card](https://goreportcard.com/badge/github.com/XenitAB/flux-status)](https://goreportcard.com/report/github.com/XenitAB/flux-status)
[![Docker Repository on Quay](https://quay.io/repository/xenitab/flux-status/status "Docker Repository on Quay")](https://quay.io/repository/xenitab/flux-status)

Flux Status synchronizes events from [Flux](https://github.com/fluxcd/flux) with the source commit.

Flux is a tool to implement a GitOps deployment flow in Kubernetes, but it lacks a UI to give
feedback regarding the status of the deployment. It is possible to read the logs from Flux but
it is not user friendly for new users. There are solutions such as [fluxcloud](https://github.com/justinbarrick/fluxcloud) that exports the deployment events to communication tools but they can easily overflow as the amount of flux instances increases.

<p align="center">
  <img src="./assets/workflow.png">
</p>

Instead Flux Status aims to give deployment feedback at the source of the deployment, the git repository. Flux Status updates the commit status with the result of the synchronization loop.
Currently there are two events that can be sent to the commit status. The result of Flux applying
the manifests to the cluster and the state of the workloads after they have been updated.

## How To
The simplest way to run Flux Status is as a sidecar in the Flux Pod, as it simplifies the life cycle
management as the container will be created with Flux.

Given that you are using the [official Flux Helm chart](https://github.com/fluxcd/helm-operator/tree/master/chart/helm-operator) use the following values. An additional token argument specific for your git provider is needed for the sidecar, for more information read about [Notifiers](## Notifiers).
```yaml
git:
  url: <git-url>

additionalArgs:
  - --connect=ws://localhost:3000

extraContainers:
  - name: "flux-status"
    image: "quay.io/xenitab/flux-status:v0.1.0-rc2"
    imagePullPolicy: "IfNotPresent"
    args:
      - --git-url=<git-url>
```

## Notifiers
Flux Status uses different notifier depending on the git provider used, and they require different
types configuration parameters depending on the notifier used. The main parameter needed is the
token used to authenticate with the different APIs.

### Azure DevOps
The Azure DevOps notifier requires a [personal access token](https://docs.microsoft.com/en-us/azure/devops/organizations/accounts/use-personal-access-tokens-to-authenticate?view=azure-devops&tabs=preview-page) to authenticate with the Azure DevOps API. The toke should be passed with the `--azdo-pat` flag.

### GitHub
The GitHub notifier requires a [personal access token](https://docs.github.com/en/github/authenticating-to-github/creating-a-personal-access-token) to authenticate with the API. The token should be passed with the `--github-token` flag. Currently the user committing the status will be the user the token belongs to. There is no way of overriding this currently, but in the future it might be possible to use an OAuth app instead.

### GitLab
The GitLab notifier requires an [access token](https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html) to authenticate with the GitLab API. The token should be passed with the `--gitlab-token` flag.

### Bitbucket
TBD

## CLI
Flux Status also has a CLI which makes the process of getting the status of a commit set by Flux Status easier. You can download the CLI binary from the [Release Page](https://github.com/XenitAB/flux-status/releases).
The configuration is similar to the Flux Status daemon. All you need is the instance name, git URL, commit id, and token to get the status.
```shell
$ flux-status-cli --instance dev  --git-url <git-url> --commit-id <commit-id> --azdo-pat <pat>
{"name":"flux-status/dev/workload","state":"succeeded"}
```

## License
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

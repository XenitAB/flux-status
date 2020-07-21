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
Currently there aare two events that can be sent to the commit status. The result of Flux applying
the manifets to the cluster and the state of the workloads after they have been updated.

## How To
The simplest way to run Flux Status is as a sidecar in the Flux Pod, as it simplifies the lifecycle
management as the container will be created with Flux.

Given that you are using the [oficial Flux Helm chart](https://github.com/fluxcd/helm-operator/tree/master/chart/helm-operator) use the following values. An additional token argument specific for your git provider is needed for the sider, for more imformation read about [Notifiers](## Notifiers).
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

### Gitlab
The Gitlab notifier requires an [access token](https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html) to authenticate with the Gitlab API. The token should be passed with the `--gitlab-token` flag.

### Azure Devops
The Azure Devops notifier requires a [personal access token]() to autheticate with the Azure Devops API. THhe toke should be passed with the `--azdo-pat` flag.

### Github
TBD

### Bitbucket
TBD

## License
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

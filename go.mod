module github.com/xenitab/flux-status

go 1.14

require (
	github.com/fluxcd/flux v1.20.0
	github.com/go-logr/logr v0.2.0
	github.com/go-logr/zapr v0.2.0
	github.com/gorilla/mux v1.7.4
	github.com/gorilla/websocket v1.4.2
	github.com/microsoft/azure-devops-go-api/azuredevops v1.0.0-b3
	github.com/onsi/gomega v1.10.1
	github.com/spf13/pflag v1.0.5
	github.com/xanzy/go-gitlab v0.33.0
	go.uber.org/zap v1.15.0
)

replace (
	github.com/docker/docker => github.com/docker/docker v0.7.3-0.20190327010347-be7ac8be2ae0
	github.com/fluxcd/flux/pkg/install => github.com/fluxcd/flux/pkg/install v0.0.0-20200205115544-4fc656b636e3
	github.com/fluxcd/helm-operator => github.com/fluxcd/helm-operator v1.0.0-rc9
	github.com/fluxcd/helm-operator/pkg/install => github.com/fluxcd/helm-operator/pkg/install v0.0.0-20200213151218-f7e487142b46
)

TAG = latest
#TAG = $(shell git describe --tags --exact-match || git describe --always --dirty)
IMG ?= xenitab/flux-status:$(TAG)

test:
	go test ./...

build:
	go build -o bin/flux-status cmd/flux-status/main.go

deploy:
	kustomize build  manifests | kubectl  apply -f -

docker-build:
	docker build -t ${IMG} .

kind-load:
	kind load docker-image $(IMG)

docker-push:
	docker push -t ${IMG} .

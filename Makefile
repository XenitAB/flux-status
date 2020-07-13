TAG = dev
#TAG = $(shell git describe --tags --exact-match || git describe --always --dirty)
IMG ?= quay.io/xenitab/flux-status:$(TAG)

fmt:
	go fmt ./...

vet:
	go vet ./...

test: fmt vet
	go test ./...

deploy:
	kustomize build  manifests | kubectl  apply -f -

docker-build:
	docker build -t ${IMG} .

kind-load:
	kind load docker-image $(IMG)

docker-push:
	docker push -t ${IMG} .

TAG = dev
IMG ?= quay.io/xenitab/flux-status:$(TAG)

assets:
	draw.io -b 10 -x -f png -p 0 -o assets/workflow.png assets/diagram.drawio
.PHONY: assets

fmt:
	go fmt ./...

lint:
	golint ./...

vet:
	go vet ./...

test: vet fmt
	go test -timeout 1m ./...

deploy:
	kustomize build  manifests | kubectl  apply -f -

docker-build:
	docker build -t $(IMG) .

kind-load:
	kind load docker-image $(IMG)

docker-push:
	docker push $(IMG)

cli:
	go build -o bin/flux-status-cli cmd/flux-status-cli/main.go

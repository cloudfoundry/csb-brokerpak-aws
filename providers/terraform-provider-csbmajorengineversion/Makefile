.DEFAULT_GOAL = help

GO-VERSION=$(shell go mod edit -json | jq -r .Go)
GOTOOLCHAIN=go$(GO-VERSION)
GO-VER = go$(GO-VERSION)
GO_OK :=  $(or $(USE_GO_CONTAINERS), $(shell which go 1>/dev/null 2>/dev/null; echo $$?))
DOCKER_OK := $(shell which docker 1>/dev/null 2>/dev/null; echo $$?)
ifeq ($(GO_OK), 0)
  GO = go
  GOFMT = gofmt
else ifeq ($(DOCKER_OK), 0)
  GO_DOCKER_OPTS = --rm \
  		-v $(PWD)/../..:/src \
  		--workdir /src/providers/terraform-provider-csbmajorengineversion \
  		--network=host \
		-e GOTOOLCHAIN \
  		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY
  GO = docker run $(GO_DOCKER_OPTS) golang:$(GO-VERSION) go
  GOFMT = docker run $(GO_DOCKER_OPTS) golang:$(GO-VERSION) gofmt
else
  $(error either Go or Docker must be installed)
endif

VERSION = 1.0.0

.PHONY: help
help: ## list Makefile targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: test
test: version download checkfmt checkimports vet ginkgo ## run all build, static analysis, and test steps

.PHONY: build
build: version download checkfmt checkimports vet build_binaries_in_cloudfoundry_namespace ## build the provider


.PHONY: build_binaries_in_cloudfoundry_namespace
build_binaries_in_cloudfoundry_namespace:
	mkdir -p ../build/cloudfoundry.org/cloud-service-broker/csbmajorengineversion/$(VERSION)/linux_amd64
	mkdir -p ../build/cloudfoundry.org/cloud-service-broker/csbmajorengineversion/$(VERSION)/darwin_amd64
	CGO_ENABLED=0 GOOS=linux $(GO) build -o ../build/cloudfoundry.org/cloud-service-broker/csbmajorengineversion/$(VERSION)/linux_amd64/terraform-provider-csbmajorengineversion_v$(VERSION)
	CGO_ENABLED=0 GOOS=darwin $(GO) build -o ../build/cloudfoundry.org/cloud-service-broker/csbmajorengineversion/$(VERSION)/darwin_amd64/terraform-provider-csbmajorengineversion_v$(VERSION)

.PHONY: clean
clean: ## clean up build artifacts
	- rm -rf ../build/cloudfoundry.org

download: ## download dependencies
	$(GO) mod download

vet: ## run static code analysis
	$(GO) vet ./...
	$(GO) run honnef.co/go/tools/cmd/staticcheck ./...

checkfmt: ## check that the code is formatted correctly
	@@if [ -n "$$(${GOFMT} -s -e -l -d .)" ]; then \
		echo "gofmt check failed: run 'make fmt'"; \
		exit 1; \
	fi

checkimports: ## check that imports are formatted correctly
	@@if [ -n "$$(${GO} run golang.org/x/tools/cmd/goimports -l -d .)" ]; then \
		echo "goimports check failed: run 'make fmt'";  \
		exit 1; \
	fi

fmt: ## format the code
	$(GOFMT) -s -e -l -w .
	$(GO) run golang.org/x/tools/cmd/goimports -l -w .

.PHONY: ginkgo
ginkgo: ## run the tests with Ginkgo
	$(GO) run github.com/onsi/ginkgo/v2/ginkgo -r

.PHONY: ginkgo-coverage
ginkgo-coverage: ## ginkgo tests coverage score
	go test -coverprofile=/tmp/csbmajorengineversion-coverage.out ./...
	go tool cover -func /tmp/csbmajorengineversion-coverage.out | grep total

.PHONY: version
version:
	@@$(GO) version
	@@if [ "$$(${GO} version | awk '{print $$3}')" != "${GO-VER}" ]; then \
		echo "Go version does not match: expected: ${GO-VER}, got $$(${GO} version | awk '{print $$3}')"; \
		exit 1; \
	fi

bump-go: ## bump go to latest version
	$(eval GO-NEW-VERSION := $(shell curl -L 'https://go.dev/VERSION?m=text' | head -n1 | cut -c3-))
	go mod edit -go $(GO-NEW-VERSION)

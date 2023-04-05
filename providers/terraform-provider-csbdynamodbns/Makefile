.DEFAULT_GOAL = help

GO-VERSION = 1.20.3
GO-VER = go$(GO-VERSION)
GO_OK :=  $(or $(USE_GO_CONTAINERS), $(shell which go 1>/dev/null 2>/dev/null; echo $$?))
DOCKER_OK := $(shell which docker 1>/dev/null 2>/dev/null; echo $$?)
ifeq ($(GO_OK), 0)
  GO=go
else ifeq ($(DOCKER_OK), 0)
  GO=docker run --rm -v $(PWD)/../..:/src -w /src/providers/terraform-provider-csbmssqldbrunfailover -e GOARCH -e GOOS -e CGO_ENABLED golang:$(GO-VERSION) go
else
  $(error either Go or Docker must be installed)
endif

VERSION = 1.0.0

SRC = $(shell find . -name "*.go" | grep -v "_test\." )

.PHONY: help
help: ## list Makefile targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: test
test: version download checkfmt checkimports vet ginkgo ## run all build, static analysis, and test steps

.PHONY: build
build: version download checkfmt checkimports vet ../build/cloudfoundry.org ## build the provider

../build/cloudfoundry.org: *.go */*.go
	mkdir -p ../build/cloudfoundry.org/cloud-service-broker/csbdynamodbns/$(VERSION)/linux_amd64
	mkdir -p ../build/cloudfoundry.org/cloud-service-broker/csbdynamodbns/$(VERSION)/darwin_amd64
	CGO_ENABLED=0 GOOS=linux $(GO) build -o ../build/cloudfoundry.org/cloud-service-broker/csbdynamodbns/$(VERSION)/linux_amd64/terraform-provider-csbdynamodbns_v$(VERSION)
	CGO_ENABLED=0 GOOS=darwin $(GO) build -o ../build/cloudfoundry.org/cloud-service-broker/csbdynamodbns/$(VERSION)/darwin_amd64/terraform-provider-csbdynamodbns_v$(VERSION)

.PHONY: clean
clean: ## clean up build artifacts
	- rm -rf ../build/cloudfoundry.org

download: ## download dependencies
	go mod download

vet: ## run static code analysis
	go vet ./...
	go run honnef.co/go/tools/cmd/staticcheck ./...

checkfmt: ## check that the code is formatted correctly
	@@if [ -n "$$(gofmt -s -e -l -d .)" ]; then \
		echo "gofmt check failed: run 'make fmt'"; \
		exit 1; \
	fi

checkimports: ## check that imports are formatted correctly
	@@if [ -n "$$(go run golang.org/x/tools/cmd/goimports -l -d .)" ]; then \
		echo "goimports check failed: run 'make fmt'";  \
		exit 1; \
	fi

fmt: ## format the code
	gofmt -s -e -l -w .
	go run golang.org/x/tools/cmd/goimports -l -w .

.PHONY: ginkgo
ginkgo: generate ## run the tests with Ginkgo
	go run github.com/onsi/ginkgo/v2/ginkgo -r

.PHONY: generate
generate: ## generate test fakes
	cd csbdynamodbns; go generate; cd ..

.PHONY: version
version:
	@@go version
	@@if [ "$$(go version | awk '{print $$3}')" != "${GO-VER}" ]; then \
		echo "Go version does not match: expected: ${GO-VER}, got $$(go version | awk '{print $$3}')"; \
		exit 1; \
	fi

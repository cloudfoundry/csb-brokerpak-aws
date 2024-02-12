.DEFAULT_GOAL = help

  GO = go
  GOFMT = gofmt

VERSION = 1.0.0

SRC = $(shell find . -name "*.go" | grep -v "_test\." )

.PHONY: help
help: ## list Makefile targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: test
test: download checkfmt checkimports vet ginkgo ## run all build, static analysis, and test steps

.PHONY: build
build: download checkfmt checkimports vet ../build/cloudfoundry.org ## build the provider

../build/cloudfoundry.org: *.go */*.go
	mkdir -p ../build/cloudfoundry.org/cloud-service-broker/csbdynamodbns/$(VERSION)/linux_amd64
	mkdir -p ../build/cloudfoundry.org/cloud-service-broker/csbdynamodbns/$(VERSION)/darwin_amd64
	CGO_ENABLED=0 GOOS=linux $(GO) build -o ../build/cloudfoundry.org/cloud-service-broker/csbdynamodbns/$(VERSION)/linux_amd64/terraform-provider-csbdynamodbns_v$(VERSION)
	CGO_ENABLED=0 GOOS=darwin $(GO) build -o ../build/cloudfoundry.org/cloud-service-broker/csbdynamodbns/$(VERSION)/darwin_amd64/terraform-provider-csbdynamodbns_v$(VERSION)

.PHONY: clean
clean: ## clean up build artifacts
	- rm -rf ../build/cloudfoundry.org
	- rm -rf /tmp/tpdynamodbns-non-fake.txt
	- rm -rf /tmp/tpdynamodbns-pkgs.txt
	- rm -rf /tmp/tpdynamodbns-coverage.out

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
ginkgo: generate ## run the tests with Ginkgo
	$(GO) run github.com/onsi/ginkgo/v2/ginkgo -r

.PHONY: ginkgo-coverage
ginkgo-coverage: ## ginkgo tests coverage score
	go list ./... | grep -v fake > /tmp/tpdynamodbns-non-fake.txt
	paste -sd "," /tmp/tpdynamodbns-non-fake.txt > /tmp/tpdynamodbns-pkgs.txt
	go test -coverpkg=`cat /tmp/tpdynamodbns-pkgs.txt` -coverprofile=/tmp/tpdynamodbns-coverage.out ./...
	go tool cover -func /tmp/tpdynamodbns-coverage.out | grep total

.PHONY: generate
generate: ## generate test fakes
	cd csbdynamodbns; $(GO) generate; cd ..


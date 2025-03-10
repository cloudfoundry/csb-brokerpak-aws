.DEFAULT_GOAL = help

  GO = go
  GOFMT = gofmt

VERSION = 1.0.0

.PHONY: help
help: ## list Makefile targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: test
test: download checkfmt checkimports vet ginkgo ## run all build, static analysis, and test steps

.PHONY: build
build: download checkfmt checkimports vet build_binaries_in_cloudfoundry_namespace ## build the provider


.PHONY: build_binaries_in_cloudfoundry_namespace
build_binaries_in_cloudfoundry_namespace:
	mkdir -p ../build/cloudfoundry.org/cloud-service-broker/csbmajorengineversion/$(VERSION)/linux_amd64
	mkdir -p ../build/cloudfoundry.org/cloud-service-broker/csbmajorengineversion/$(VERSION)/darwin_amd64
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -o ../build/cloudfoundry.org/cloud-service-broker/csbmajorengineversion/$(VERSION)/linux_amd64/terraform-provider-csbmajorengineversion_v$(VERSION)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GO) build -o ../build/cloudfoundry.org/cloud-service-broker/csbmajorengineversion/$(VERSION)/darwin_amd64/terraform-provider-csbmajorengineversion_v$(VERSION)
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GO) build -o ../build/cloudfoundry.org/cloud-service-broker/csbmajorengineversion/$(VERSION)/darwin_arm64/terraform-provider-csbmajorengineversion_v$(VERSION)

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

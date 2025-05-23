.DEFAULT_GOAL = help
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
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ../build/cloudfoundry.org/cloud-service-broker/csbdynamodbns/$(VERSION)/linux_amd64/terraform-provider-csbdynamodbns_v$(VERSION)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o ../build/cloudfoundry.org/cloud-service-broker/csbdynamodbns/$(VERSION)/darwin_amd64/terraform-provider-csbdynamodbns_v$(VERSION)
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o ../build/cloudfoundry.org/cloud-service-broker/csbdynamodbns/$(VERSION)/darwin_arm64/terraform-provider-csbdynamodbns_v$(VERSION)

.PHONY: clean
clean: ## clean up build artifacts
	- rm -rf ../build/cloudfoundry.org
	- rm -rf /tmp/tpdynamodbns-non-fake.txt
	- rm -rf /tmp/tpdynamodbns-pkgs.txt
	- rm -rf /tmp/tpdynamodbns-coverage.out

download: ## download dependencies
	go mod download

vet: ## run static code analysis
	go vet ./...
	go tool staticcheck ./...

checkfmt: ## check that the code is formatted correctly
	@@if [ -n "$$(gofmt -s -e -l -d .)" ]; then \
		echo "gofmt check failed: run 'make fmt'"; \
		exit 1; \
	fi

checkimports: ## check that imports are formatted correctly
	@@if [ -n "$$(go tool goimports -l -d .)" ]; then \
		echo "goimports check failed: run 'make fmt'";  \
		exit 1; \
	fi

fmt: ## format the code
	gofmt -s -e -l -w .
	go tool goimports -l -w .

.PHONY: ginkgo
ginkgo: generate ## run the tests with Ginkgo
	go tool ginkgo -r

.PHONY: ginkgo-coverage
ginkgo-coverage: ## ginkgo tests coverage score
	go list ./... | grep -v fake > /tmp/tpdynamodbns-non-fake.txt
	paste -sd "," /tmp/tpdynamodbns-non-fake.txt > /tmp/tpdynamodbns-pkgs.txt
	go test -coverpkg=`cat /tmp/tpdynamodbns-pkgs.txt` -coverprofile=/tmp/tpdynamodbns-coverage.out ./...
	go tool cover -func /tmp/tpdynamodbns-coverage.out | grep total

.PHONY: generate
generate: ## generate test fakes
	cd csbdynamodbns; go generate; cd ..


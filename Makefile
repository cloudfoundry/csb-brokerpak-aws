###### Help ###################################################################
.DEFAULT_GOAL = help

.PHONY: help
help: ## list Makefile targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

###### Setup ##################################################################
IAAS=aws
GO-VERSION = 1.18.2
GO-VER = go$(GO-VERSION)
CSB_VERSION := $(or $(CSB_VERSION), $(shell grep 'github.com/cloudfoundry/cloud-service-broker' go.mod | grep -v replace | awk '{print $$NF}' | sed -e 's/v//'))
CSB_RELEASE_VERSION := $(CSB_VERSION) # this doesnt work well if we did make latest-csb.

CSB_DOCKER_IMAGE := $(or $(CSB), cfplatformeng/csb:$(CSB_VERSION))
GO_OK := $(or $(USE_GO_CONTAINERS), $(shell which go 1>/dev/null 2>/dev/null; echo $$?))
DOCKER_OK := $(shell which docker 1>/dev/null 2>/dev/null; echo $$?)

####### broker environment variables
PAK_CACHE=/tmp/.pak-cache
SECURITY_USER_NAME := $(or $(SECURITY_USER_NAME), aws-broker)
SECURITY_USER_PASSWORD := $(or $(SECURITY_USER_PASSWORD), aws-broker-pw)
PARALLEL_JOB_COUNT := $(or $(PARALLEL_JOB_COUNT), 2)
GSB_PROVISION_DEFAULTS := $(or $(GSB_PROVISION_DEFAULTS), {"aws_vpc_id": "$(AWS_PAS_VPC_ID)"})

ifeq ($(GO_OK), 0)  # use local go binary
GO=go
GOFMT=gofmt
BROKER_GO_OPTS=PORT=8080 \
				DB_TYPE=sqlite3 \
				DB_PATH=/tmp/csb-db \
				SECURITY_USER_NAME=$(SECURITY_USER_NAME) \
				SECURITY_USER_PASSWORD=$(SECURITY_USER_PASSWORD) \
				AWS_ACCESS_KEY_ID='$(AWS_ACCESS_KEY_ID)' \
				AWS_SECRET_ACCESS_KEY=$(AWS_SECRET_ACCESS_KEY) \
 				PAK_BUILD_CACHE_PATH=$(PAK_CACHE) \
 				GSB_PROVISION_DEFAULTS='$(GSB_PROVISION_DEFAULTS)'

PAK_PATH=$(PWD)
RUN_CSB=$(BROKER_GO_OPTS) go run github.com/cloudfoundry/cloud-service-broker
LDFLAGS="-X github.com/cloudfoundry/cloud-service-broker/utils.Version=$(CSB_VERSION)"
GET_CSB="env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags $(LDFLAGS) github.com/cloudfoundry/cloud-service-broker"
else ifeq ($(DOCKER_OK), 0)
BROKER_DOCKER_OPTS=--rm -v $(PAK_CACHE):$(PAK_CACHE) -v $(PWD):/brokerpak -w /brokerpak --network=host  \
    -p 8080:8080 \
		-e SECURITY_USER_NAME \
		-e SECURITY_USER_PASSWORD \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		-e "DB_TYPE=sqlite3" \
		-e "DB_PATH=/tmp/csb-db" \
		-e PAK_BUILD_CACHE_PATH=$(PAK_CACHE) \
		-e GSB_PROVISION_DEFAULTS

RUN_CSB=docker run $(BROKER_DOCKER_OPTS) $(CSB_DOCKER_IMAGE)

#### running go inside a container, this is for integration tests and push-broker
# path inside the container
PAK_PATH=/brokerpak

GO_DOCKER_OPTS=--rm -v $(PAK_CACHE):$(PAK_CACHE) -v $(PWD):/brokerpak -w /brokerpak --network=host
GO=docker run $(GO_DOCKER_OPTS) golang:latest go
GOFMT=docker run $(GO_DOCKER_OPTS) golang:latest gofmt

# this doesnt work well if we did make latest-csb. We should build it instead, with go inside a container.
GET_CSB="wget -O cloud-service-broker https://github.com/cloudfoundry/cloud-service-broker/releases/download/v$(CSB_RELEASE_VERSION)/cloud-service-broker.linux && chmod +x cloud-service-broker"
else
$(error either Go or Docker must be installed)
endif

###### Targets ################################################################

.PHONY: deps-go-binary
deps-go-binary:
ifeq ($(SKIP_GO_VERSION_CHECK),)
	@@if [ "$$($(GO) version | awk '{print $$3}')" != "${GO-VER}" ]; then \
		echo "Go version does not match: expected: ${GO-VER}, got $$($(GO) version | awk '{print $$3}')"; \
		exit 1; \
	fi
endif

.PHONY: build
build: deps-go-binary $(IAAS)-services-*.brokerpak ## build brokerpak

$(IAAS)-services-*.brokerpak: *.yml terraform/*/*/*.tf terraform/*/*/*/*.tf | $(PAK_CACHE)
	$(RUN_CSB) pak build

###### Run ###################################################################
.PHONY: run
run: build aws_access_key_id aws_secret_access_key ## start broker with this brokerpak
	$(RUN_CSB) serve

###### docs ###################################################################

.PHONY: docs
docs: build brokerpak-user-docs.md ## build docs

brokerpak-user-docs.md: *.yml
	$(RUN_CSB) pak docs $(PAK_PATH)/$(shell ls *.brokerpak) > $@

###### examples ###################################################################

.PHONY: examples
examples: ## display available examples
	 $(RUN_CSB) client examples

###### run-examples ###################################################################
PARALLEL_JOB_COUNT := $(or $(PARALLEL_JOB_COUNT), 10000)

.PHONY: run-examples
run-examples: ## run examples in yml files. Runs examples for all services by default. Set service_name and example_name to run all examples for a specific service or an specific example.
	$(RUN_CSB) client run-examples --service-name="$(service_name)" --example-name="$(example_name)" -j $(PARALLEL_JOB_COUNT)

###### test ###################################################################

.PHONY: test
test: run-integration-tests ## run the tests

.PHONY: lint run-integration-tests
run-integration-tests: latest-csb ## run integration tests for this brokerpak
	cd ./integration-tests && go run github.com/onsi/ginkgo/v2/ginkgo -r .

###### info ###################################################################

.PHONY: info ## show brokerpak info
info: build
	$(RUN_CSB) pak info $(PAK_PATH)/$(shell ls *.brokerpak)

###### validate ###################################################################

.PHONY: validate
validate: build ## validate pak syntax
	$(RUN_CSB) pak validate $(PAK_PATH)/$(shell ls *.brokerpak)

###### push-broker ###################################################################

# fetching bits for cf push broker
cloud-service-broker: go.mod ## build or fetch CSB binary
	$(shell "$(GET_CSB)")

APP_NAME := $(or $(APP_NAME), cloud-service-broker-aws)
DB_TLS := $(or $(DB_TLS), skip-verify)

.PHONY: push-broker
push-broker: cloud-service-broker build aws_access_key_id aws_secret_access_key aws_pas_vpc_id ## push the broker with this brokerpak
	MANIFEST=cf-manifest.yml APP_NAME=$(APP_NAME) DB_TLS=$(DB_TLS) GSB_PROVISION_DEFAULTS='$(GSB_PROVISION_DEFAULTS)' ./scripts/push-broker.sh

.PHONY: push-local-broker
push-local-broker: local-cloud-service-broker build aws_access_key_id aws_secret_access_key aws_pas_vpc_id ## push the broker with this brokerpak
	MANIFEST=cf-manifest.yml APP_NAME=$(APP_NAME) DB_TLS=$(DB_TLS) GSB_PROVISION_DEFAULTS='$(GSB_PROVISION_DEFAULTS)' ./scripts/push-broker.sh

.PHONY: aws_access_key_id
aws_access_key_id:
ifndef AWS_ACCESS_KEY_ID
	$(error variable AWS_ACCESS_KEY_ID not defined)
endif

.PHONY: aws_secret_access_key
aws_secret_access_key:
ifndef AWS_SECRET_ACCESS_KEY
	$(error variable AWS_SECRET_ACCESS_KEY not defined)
endif

.PHONY: aws_pas_vpc_id
aws_pas_vpc_id:
ifndef AWS_PAS_VPC_ID
	$(error variable AWS_PAS_VPC_ID not defined - must be VPC ID for PAS foundation)
endif

###### clean ###################################################################

.PHONY: clean
clean: ## delete build files
	- rm -f $(IAAS)-services-*.brokerpak
	- rm -f ./cloud-service-broker
	- rm -f ./brokerpak-user-docs.md
	- rm -rf ../aws-released
	- rm -rf $(PAK_CACHE)

$(PAK_CACHE):
	@echo "Folder $(PAK_CACHE) does not exist. Creating it..."
	mkdir -p $@

.PHONY: latest-csb
latest-csb: ## point to the very latest CSB on GitHub
	$(GO) get -d github.com/cloudfoundry/cloud-service-broker@main
	$(GO) mod tidy

.PHONY: local-csb
local-csb: ## point to a local CSB repo
	echo "replace \"github.com/cloudfoundry/cloud-service-broker\" => \"$$PWD/../cloud-service-broker\"" >>go.mod
	$(GO) mod tidy

###### lint ###################################################################

.PHONY: lint
lint: checkformat checkimports format vet staticcheck ## checks format, imports and vet

checkformat: ## checks that the code is formatted correctly
	@@if [ -n "$$(${GOFMT} -s -e -l -d .)" ]; then       \
		echo "gofmt check failed: run 'make format'"; \
		exit 1;                                       \
	fi

checkimports: ## checks that imports are formatted correctly
	@@if [ -n "$$(${GO} run golang.org/x/tools/cmd/goimports -l -d .)" ]; then \
		echo "goimports check failed: run 'make format'";                      \
		exit 1;                                                                \
	fi

vet: ## runs go vet
	${GO} vet ./...

staticcheck: ## runs staticcheck
	${GO} run honnef.co/go/tools/cmd/staticcheck ./...

.PHONY: format
format: ## format the source
	${GOFMT} -s -e -l -w .
	${GO} run golang.org/x/tools/cmd/goimports -l -w .
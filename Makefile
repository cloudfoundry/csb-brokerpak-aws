###### Help ###################################################################
.DEFAULT_GOAL = help

.PHONY: help
help: ## list Makefile targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

###### Setup ##################################################################
IAAS=aws
CSB_VERSION := $(or $(CSB_VERSION), $(shell grep 'github.com/cloudfoundry/cloud-service-broker' go.mod | grep -v replace | awk '{print $$NF}' | sed -e 's/v//'))
CSB_RELEASE_VERSION := $(CSB_VERSION) # this doesnt work well if we did make latest-csb.

####### broker environment variables
SECURITY_USER_NAME := $(or $(SECURITY_USER_NAME), aws-broker)
SECURITY_USER_PASSWORD := $(or $(SECURITY_USER_PASSWORD), aws-broker-pw)
GSB_PROVISION_DEFAULTS := $(or $(GSB_PROVISION_DEFAULTS), {"aws_vpc_id": "$(AWS_PAS_VPC_ID)"})

BROKER_GO_OPTS=PORT=8080 \
				SECURITY_USER_NAME=$(SECURITY_USER_NAME) \
				SECURITY_USER_PASSWORD=$(SECURITY_USER_PASSWORD) \
				AWS_ACCESS_KEY_ID='$(AWS_ACCESS_KEY_ID)' \
				AWS_SECRET_ACCESS_KEY=$(AWS_SECRET_ACCESS_KEY) \
				DB_TYPE=sqlite3 \
				DB_PATH=/tmp/csb-db \
				BROKERPAK_UPDATES_ENABLED=$(BROKERPAK_UPDATES_ENABLED) \
				PAK_BUILD_CACHE_PATH=$(PAK_BUILD_CACHE_PATH) \
				GSB_PROVISION_DEFAULTS='$(GSB_PROVISION_DEFAULTS)' \
				GSB_SERVICE_CSB_AWS_S3_BUCKET_PLANS='$(GSB_SERVICE_CSB_AWS_S3_BUCKET_PLANS)' \
				GSB_SERVICE_CSB_AWS_POSTGRESQL_PLANS='$(GSB_SERVICE_CSB_AWS_POSTGRESQL_PLANS)' \
				GSB_SERVICE_CSB_AWS_AURORA_POSTGRESQL_PLANS='$(GSB_SERVICE_CSB_AWS_AURORA_POSTGRESQL_PLANS)' \
				GSB_SERVICE_CSB_AWS_AURORA_MYSQL_PLANS='$(GSB_SERVICE_CSB_AWS_AURORA_MYSQL_PLANS)' \
				GSB_SERVICE_CSB_AWS_MYSQL_PLANS='$(GSB_SERVICE_CSB_AWS_MYSQL_PLANS)' \
				GSB_SERVICE_CSB_AWS_REDIS_PLANS='$(GSB_SERVICE_CSB_AWS_REDIS_PLANS)' \
				GSB_SERVICE_CSB_AWS_SQS_PLANS='$(GSB_SERVICE_CSB_AWS_SQS_PLANS)' \
				GSB_COMPATIBILITY_ENABLE_BETA_SERVICES='$(GSB_COMPATIBILITY_ENABLE_BETA_SERVICES)'

PAK_PATH=$(PWD)
RUN_CSB=$(BROKER_GO_OPTS) go run github.com/cloudfoundry/cloud-service-broker
LDFLAGS="-X github.com/cloudfoundry/cloud-service-broker/utils.Version=$(CSB_VERSION)"
GET_CSB="env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags $(LDFLAGS) github.com/cloudfoundry/cloud-service-broker"

###### Targets ################################################################

.PHONY: build
build: $(IAAS)-services-*.brokerpak ## build brokerpak

$(IAAS)-services-*.brokerpak: *.yml terraform/*/*/*.tf terraform/*/*/*/*.tf providers | $(PAK_BUILD_CACHE_PATH)
	$(RUN_CSB) pak build


.PHONY: providers
providers: providers/build/cloudfoundry.org/cloud-service-broker/csbdynamodbns providers/build/cloudfoundry.org/cloud-service-broker/csbmajorengineversion ## build custom providers

providers/build/cloudfoundry.org/cloud-service-broker/csbdynamodbns:
	cd providers/terraform-provider-csbdynamodbns; $(MAKE) build

providers/build/cloudfoundry.org/cloud-service-broker/csbmajorengineversion:
	cd providers/terraform-provider-csbmajorengineversion; $(MAKE) build

###### Run ###################################################################
.PHONY: run
run: aws_access_key_id aws_secret_access_key ## start broker with this brokerpak
	$(RUN_CSB) pak build --target current
	$(RUN_CSB) serve

###### docs ###################################################################

.PHONY: docs
docs: build brokerpak-user-docs.md ## build docs

brokerpak-user-docs.md: *.yml
	$(RUN_CSB) pak docs $(PAK_PATH)/$(shell ls *.brokerpak) > $@

###### examples ###################################################################

.PHONY: examples
examples: ## display available examples
	 $(RUN_CSB) examples

###### run-examples ###################################################################
.PHONY: run-examples
run-examples: providers ## run examples in yml files. Runs examples for all services by default. Set service_name and/or example_name.
	$(RUN_CSB) run-examples --service-name="$(service_name)" --example-name="$(example_name)"

###### test ###################################################################

.PHONY: test-coverage
test-coverage: ## test coverage score
	- cd providers/terraform-provider-csbdynamodbns; $(MAKE) ginkgo-coverage
	- cd providers/terraform-provider-csbmajorengineversion; $(MAKE) ginkgo-coverage

.PHONY: test
test: lint run-integration-tests ## run the tests

.PHONY: run-integration-tests
run-integration-tests: run-provider-tests ## run integration tests for this brokerpak
	cd ./integration-tests && go run github.com/onsi/ginkgo/v2/ginkgo -r .

.PHONY: run-terraform-tests
run-terraform-tests: providers custom.tfrc ## run terraform tests for this brokerpak
	cd ./terraform-tests && TF_CLI_CONFIG_FILE="$(PWD)/custom.tfrc" go run github.com/onsi/ginkgo/v2/ginkgo -r --label-filter="${LABEL_FILTER}" .

.PHONY: run-modified-tests
run-modified-tests: providers custom.tfrc
	TF_CLI_CONFIG_FILE="$(PWD)/custom.tfrc" go run github.com/onsi/ginkgo/v2/ginkgo -r --label-filter="${LABEL_FILTER}" --timeout=3h --focus-file none $$(git diff --name-only HEAD | awk '{printf(" --focus-file  %s", $$0)}')

.PHONY: run-provider-tests
run-provider-tests:  ## run the integration tests associated with providers
	cd providers/terraform-provider-csbdynamodbns; $(MAKE) test

custom.tfrc:
	sed "s#BROKERPAK_PATH#$(PWD)#" custom.tfrc.template > $@

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
	"$(GET_CSB)"

APP_NAME := $(or $(APP_NAME), cloud-service-broker-aws)
DB_TLS := $(or $(DB_TLS), skip-verify)

.PHONY: push-broker
push-broker: cloud-service-broker build aws_access_key_id aws_secret_access_key aws_pas_vpc_id ## push the broker with this brokerpak
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
	- cd providers/terraform-provider-csbdynamodbns; $(MAKE) clean
	- cd providers/terraform-provider-csbmajorengineversion; $(MAKE) clean

$(PAK_BUILD_CACHE_PATH):
	@echo "Folder $(PAK_BUILD_CACHE_PATH) does not exist. Creating it..."
	mkdir -p $@

.PHONY: latest-csb
latest-csb: ## point to the very latest CSB on GitHub
	go get -d github.com/cloudfoundry/cloud-service-broker@main
	go mod tidy

.PHONY: local-csb
local-csb: ## point to a local CSB repo
	echo "replace \"github.com/cloudfoundry/cloud-service-broker\" => \"$$PWD/../cloud-service-broker\"" >>go.mod
	go mod tidy

###### lint ###################################################################

.PHONY: lint
lint: checkgoformat checkgoimports checktfformat vet staticcheck ## checks format, imports and vet

checktfformat: ## checks that Terraform HCL is formatted correctly
	@@if [ "$$(terraform fmt -recursive --check)" ]; then \
		echo "terraform fmt check failed: run 'make format'"; \
		exit 1; \
	fi

checkgoformat: ## checks that the Go code is formatted correctly
	@@if [ -n "$$(gofmt -s -e -l -d .)" ]; then       \
		echo "gofmt check failed: run 'make format'"; \
		exit 1;                                       \
	fi

checkgoimports: ## checks that Go imports are formatted correctly
	@@if [ -n "$$(go run golang.org/x/tools/cmd/goimports -l -d .)" ]; then \
		echo "goimports check failed: run 'make format'";                      \
		exit 1;                                                                \
	fi

vet: ## runs go vet
	go vet ./...

staticcheck: ## runs staticcheck
	go run honnef.co/go/tools/cmd/staticcheck ./...

.PHONY: format
format: ## format the source
	gofmt -s -e -l -w .
	go run golang.org/x/tools/cmd/goimports -l -w .
	terraform fmt --recursive

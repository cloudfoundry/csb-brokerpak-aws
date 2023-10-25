#/bin/sh
set -euo pipefail

######################################################################
# Ideally, we could rewrite this script in Go since it would remain  #
# mostly the same no matter how many subfolder tests we add and its  #
# logic it's the cornerstone of this new type of tests.              #
# By doing so, we can also integrate it in the existing Ginkgo suite #
######################################################################

export TF_VAR_region="us-west-2"
export TF_VAR_access_key="${AWS_ACCESS_KEY_ID}"
export TF_VAR_secret_key="${AWS_SECRET_ACCESS_KEY}"

function run_test {
	test_folder="$1"

	mkdir -p "$test_folder/exec/bind" "$test_folder/exec/provision"
	cp "$test_folder/test.tf"       "$test_folder/exec/"
	cp -R $test_folder/.bind/*      "$test_folder/exec/bind/"
	cp -R $test_folder/.provision/* "$test_folder/exec/provision/"

	pushd "$test_folder/exec/"
		terraform init
		terraform apply -auto-approve

		terraform output -json "bind"      > "bind/test_bind.tfvars.json"
		terraform output -json "provision" > "provision/test_provision.tfvars.json"
	popd

	pushd "$test_folder/exec/provision/"
		terraform init
		terraform apply -auto-approve -var-file="test_provision.tfvars.json"
		terraform output > "../bind/provision_outputs.tfvars"
	popd

	pushd "$test_folder/exec/bind/"
		terraform init
		terraform apply -auto-approve -var-file="test_bind.tfvars.json" -var-file="provision_outputs.tfvars"
	popd

	pushd "$test_folder"
		sleep 10
		"./run_test.sh"
	popd
}

run_test "s3_sse_kms"
run_test "s3_no_sse"

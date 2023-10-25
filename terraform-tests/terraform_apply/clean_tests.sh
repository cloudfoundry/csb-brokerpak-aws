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

function clean_test {
	test_folder="$1"

	pushd "$test_folder"
		./clean_test.sh || echo "This test doesn't require any custom clean steps."
	popd

	pushd "$test_folder/exec/bind"
		sed -i 's/prevent_destroy = true/prevent_destroy = false/g' *.tf
		terraform apply -destroy -auto-approve -var-file="test_bind.tfvars.json" -var-file="provision_outputs.tfvars"
	popd

	pushd "$test_folder/exec/provision"
		sed -i 's/prevent_destroy = true/prevent_destroy = false/g' *.tf
		terraform apply -destroy -auto-approve -var-file="test_provision.tfvars.json"
	popd

	pushd "$test_folder/exec/"
		sed -i 's/prevent_destroy = true/prevent_destroy = false/g' *.tf
		terraform apply -destroy -auto-approve
	popd

	rm -Rf "$test_folder/exec/"
}

clean_test "s3_sse_kms"
clean_test "s3_no_sse"

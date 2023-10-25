#/bin/sh
set -euo pipefail

pushd "exec/bind/"
	export AWS_ACCESS_KEY_ID="$(terraform output -raw "access_key_id")"
	export AWS_SECRET_ACCESS_KEY="$(terraform output -raw "secret_access_key")" 
popd

pushd "exec/provision/"
	export AWS_DEFAULT_REGION="$(terraform output -raw "region")"
	export BUCKET_NAME="$(terraform output -raw "bucket_name")"
popd

pushd "exec/"
	aws s3 rm "s3://${BUCKET_NAME}/" --recursive
popd

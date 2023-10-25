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
	echo "some content from test_no_sse" > "test_no_sse_file_upload"
	aws s3 cp "test_no_sse_file_upload" "s3://${BUCKET_NAME}/"
	CONTENT="$(aws s3 cp "s3://${BUCKET_NAME}/test_no_sse_file_upload" -)"
	echo "${CONTENT}" | grep "some content from test_no_sse" || (echo "Received content is wrong: \"${CONTENT}\"" ; exit 1)
popd

#!/usr/bin/env bash

set -o pipefail
set -o nounset
set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

. "${SCRIPT_DIR}/../functions.sh"



SERVICE_NAME=postgres-$$



RESULT=1
if   create_service "csb-aws-postgresql" small "${SERVICE_NAME}" "{\"use_tls\":false}"; then

    (cd "${SCRIPT_DIR}/postgres_validator" && cf push --no-start)
    if cf bind-service postgres-validator "${SERVICE_NAME}"; then
        if cf start postgres-validator; then
            RESULT=0
            echo "postgres-validator success"
        else
            echo "postgres-validator failed"
            cf logs postgres-validator --recent
        fi
    fi
fi

if [ ${RESULT} -eq 0 ]; then
    echo "$0 SUCCESS"
else
    echo "$0 FAILED"
fi



cf delete -f -r postgres-validator
delete_service ${SERVICE_NAME}

exit ${RESULT}

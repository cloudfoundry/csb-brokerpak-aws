#!/usr/bin/env bash

set +x # Hide secrets
set -o errexit
set -o pipefail
set -e

if [[ -z ${MANIFEST} ]]; then
  MANIFEST=manifest.yml
fi

if [[ -z ${APP_NAME} ]]; then
  APP_NAME=cloud-service-broker
fi

if [[ -z ${SECURITY_USER_NAME} ]]; then
  echo "Missing SECURITY_USER_NAME variable"
  exit 1
fi

if [[ -z ${SECURITY_USER_PASSWORD} ]]; then
  echo "Missing SECURITY_USER_PASSWORD variable"
  exit 1
fi

cfmf="/tmp/cf-manifest.$$.yml"
touch "$cfmf"
trap "rm -f $cfmf" EXIT
chmod 600 "$cfmf"
cat "$MANIFEST" >$cfmf

echo "  env:" >>$cfmf
echo "    SECURITY_USER_PASSWORD: ${SECURITY_USER_PASSWORD}" >>$cfmf
echo "    SECURITY_USER_NAME: ${SECURITY_USER_NAME}" >>$cfmf
echo "    BROKERPAK_UPDATES_ENABLED: ${BROKERPAK_UPDATES_ENABLED:-true}" >>$cfmf
echo "    TERRAFORM_UPGRADES_ENABLED: ${TERRAFORM_UPGRADES_ENABLED:-true}" >>$cfmf
echo "    CSB_DISABLE_TF_UPGRADE_PROVIDER_RENAMES: ${CSB_DISABLE_TF_UPGRADE_PROVIDER_RENAMES:-false}" >>$cfmf
echo "    GSB_COMPATIBILITY_ENABLE_BETA_SERVICES: ${GSB_COMPATIBILITY_ENABLE_BETA_SERVICES:-true}" >>$cfmf

if [[ ${GSB_PROVISION_DEFAULTS} ]]; then
  echo "    GSB_PROVISION_DEFAULTS: $(echo "$GSB_PROVISION_DEFAULTS" | jq @json)" >>$cfmf
fi

if [[ ${AWS_ACCESS_KEY_ID} ]]; then
  echo "    AWS_ACCESS_KEY_ID: ${AWS_ACCESS_KEY_ID}" >>$cfmf
fi

if [[ ${AWS_SECRET_ACCESS_KEY} ]]; then
  echo "    AWS_SECRET_ACCESS_KEY: ${AWS_SECRET_ACCESS_KEY}" >>$cfmf
fi

if [[ ${GSB_BROKERPAK_BUILTIN_PATH} ]]; then
  echo "    GSB_BROKERPAK_BUILTIN_PATH: ${GSB_BROKERPAK_BUILTIN_PATH}" >>$cfmf
fi

if [[ ${CH_CRED_HUB_URL} ]]; then
  echo "    CH_CRED_HUB_URL: ${CH_CRED_HUB_URL}" >>$cfmf
fi

if [[ ${CH_UAA_URL} ]]; then
  echo "    CH_UAA_URL: ${CH_UAA_URL}" >>$cfmf
fi

if [[ ${CH_UAA_CLIENT_NAME} ]]; then
  echo "    CH_UAA_CLIENT_NAME: ${CH_UAA_CLIENT_NAME}" >>$cfmf
fi

if [[ ${CH_UAA_CLIENT_SECRET} ]]; then
  echo "    CH_UAA_CLIENT_SECRET: ${CH_UAA_CLIENT_SECRET}" >>$cfmf
fi

if [[ ${CH_SKIP_SSL_VALIDATION} ]]; then
  echo "    CH_SKIP_SSL_VALIDATION: ${CH_SKIP_SSL_VALIDATION}" >>$cfmf
fi

if [[ ${DB_TLS} ]]; then
  echo "    DB_TLS: ${DB_TLS}" >>$cfmf
fi

if [[ -z "$GSB_SERVICE_CSB_AWS_S3_BUCKET_PLANS" ]]; then
  echo "Missing GSB_SERVICE_CSB_AWS_S3_BUCKET_PLANS variable"
  exit 1
fi
echo "    GSB_SERVICE_CSB_AWS_S3_BUCKET_PLANS: $(echo "$GSB_SERVICE_CSB_AWS_S3_BUCKET_PLANS" | jq @json)" >>$cfmf

if [[ -z "$GSB_SERVICE_CSB_AWS_POSTGRESQL_PLANS" ]]; then
  echo "Missing GSB_SERVICE_CSB_AWS_POSTGRESQL_PLANS variable"
  exit 1
fi
echo "    GSB_SERVICE_CSB_AWS_POSTGRESQL_PLANS: $(echo "$GSB_SERVICE_CSB_AWS_POSTGRESQL_PLANS" | jq @json)" >>$cfmf

if [[ -z "$GSB_SERVICE_CSB_AWS_AURORA_POSTGRESQL_PLANS" ]]; then
  echo "Missing GSB_SERVICE_CSB_AWS_AURORA_POSTGRESQL_PLANS variable"
  exit 1
fi
echo "    GSB_SERVICE_CSB_AWS_AURORA_POSTGRESQL_PLANS: $(echo "$GSB_SERVICE_CSB_AWS_AURORA_POSTGRESQL_PLANS" | jq @json)" >>$cfmf

if [[ -z "$GSB_SERVICE_CSB_AWS_AURORA_MYSQL_PLANS" ]]; then
  echo "Missing GSB_SERVICE_CSB_AWS_AURORA_MYSQL_PLANS variable"
  exit 1
fi
echo "    GSB_SERVICE_CSB_AWS_AURORA_MYSQL_PLANS: $(echo "$GSB_SERVICE_CSB_AWS_AURORA_MYSQL_PLANS" | jq @json)" >>$cfmf

if [[ -z "$GSB_SERVICE_CSB_AWS_MYSQL_PLANS" ]]; then
  echo "Missing GSB_SERVICE_CSB_AWS_MYSQL_PLANS variable"
  exit 1
fi
echo "    GSB_SERVICE_CSB_AWS_MYSQL_PLANS: $(echo "$GSB_SERVICE_CSB_AWS_MYSQL_PLANS" | jq @json)" >>$cfmf

if [[ -z "$GSB_SERVICE_CSB_AWS_REDIS_PLANS" ]]; then
  echo "Missing GSB_SERVICE_CSB_AWS_REDIS_PLANS variable"
  exit 1
fi
echo "    GSB_SERVICE_CSB_AWS_REDIS_PLANS: $(echo "$GSB_SERVICE_CSB_AWS_REDIS_PLANS" | jq @json)" >>$cfmf

if [[ -z "$GSB_SERVICE_CSB_AWS_MSSQL_PLANS" ]]; then
  echo "Missing GSB_SERVICE_CSB_AWS_MSSQL_PLANS variable"
  exit 1
fi
echo "    GSB_SERVICE_CSB_AWS_MSSQL_PLANS: $(echo "$GSB_SERVICE_CSB_AWS_MSSQL_PLANS" | jq @json)" >>$cfmf

if [[ -z "$GSB_SERVICE_CSB_AWS_SQS_PLANS" ]]; then
  echo "Missing GSB_SERVICE_CSB_AWS_SQS_PLANS variable"
  exit 1
fi
echo "    GSB_SERVICE_CSB_AWS_SQS_PLANS: $(echo "$GSB_SERVICE_CSB_AWS_SQS_PLANS" | jq @json)" >>$cfmf

cf push --no-start -f "${cfmf}" --var app=${APP_NAME}

if [[ -z ${MSYQL_INSTANCE} ]]; then
  MSYQL_INSTANCE=csb-sql
fi

cf bind-service "${APP_NAME}" "${MSYQL_INSTANCE}"

cf start "${APP_NAME}"

if [[ -z ${BROKER_NAME} ]]; then
  BROKER_NAME=csb-$USER
fi

cf create-service-broker "${BROKER_NAME}" "${SECURITY_USER_NAME}" "${SECURITY_USER_PASSWORD}" https://$(LANG=EN cf app "${APP_NAME}" | grep 'routes:' | cut -d ':' -f 2 | xargs) --space-scoped --update-if-exists

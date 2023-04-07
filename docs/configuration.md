# Broker Configuration
The broker can be configured though environment variables or configuration files or a combo of both.

## Configuration File
A configuration file can be provided at run time to the broker.
```bash
cloud-service-broker serve --config <config file name>
```

A configuration file can be YAML or JSON. Config file values that are `.` delimited represent hierarchy in the config file.

Example:
```
db:
  host: hostname
```
represents a config file value of `db.host`

## Database Configuration Properties

Connection details for the backing database for the service broker.

You can configure the following values:

| Environment Variable      | Config File Value | Type   | Description                                                         |
|---------------------------|-------------------|--------|---------------------------------------------------------------------|
| <tt>DB_HOST</tt> <b>*</b> | db.host           | string | <p>Database host </p>                                               |
| <tt>DB_USERNAME</tt>      | db.user           | string | <p>Database username </p>                                           |
| <tt>DB_PASSWORD</tt>      | db.password       | secret | <p>Database password </p>                                           |
| <tt>DB_PORT</tt> <b>*</b> | db.port           | string | <p>Database port (defaults to 3306)  Default: <code>3306</code></p> |
| <tt>DB_NAME</tt> <b>*</b> | db.name           | string | <p>Database name  Default: <code>servicebroker</code></p>           |
| <tt>CA_CERT</tt>          | db.ca.cert        | text   | <p>Server CA cert </p>                                              |
| <tt>CLIENT_CERT</tt>      | db.client.cert    | text   | <p>Client cert </p>                                                 |
| <tt>CLIENT_KEY</tt>       | db.client.key     | text   | <p>Client key </p>                                                  |

## Broker Service Configuration

Broker service configuration values:

| Environment Variable                     | Config File Value | Type   | Description                           |
|------------------------------------------|-------------------|--------|---------------------------------------|
| <tt>SECURITY_USER_NAME</tt> <b>*</b>     | api.user          | string | <p>Broker authentication username</p> |
| <tt>SECURITY_USER_PASSWORD</tt> <b>*</b> | api.password      | string | <p>Broker authentication password</p> |
| <tt>PORT</tt>                            | api.port          | string | <p>Port to bind broker to</p>         |
| <tt>HOST</tt>                            | api.host          | string | <p>Host to bind broker to</p>         |

## Credhub Configuration
The broker supports passing credentials to apps via 
[credhub references](https://github.com/cloudfoundry/credhub/blob/main/docs/secure-service-credentials.md#service-brokers),
thus keeping them private to the application (they won't show up in `cf env app_name` output.)

| Environment Variable   | Config File Value           | Type    | Description                                                                                                 |
|------------------------|-----------------------------|---------|-------------------------------------------------------------------------------------------------------------|
| CH_CRED_HUB_URL        | credhub.url                 | URL     | credhub service URL - usually `https://credhub.service.cf.internal:8844`                                    |
| CH_UAA_URL             | credhub.uaa_url             | URL     | uaa service URL - usually `https://uaa.service.cf.internal:8443`                                            |
| CH_UAA_CLIENT_NAME     | credhub.uaa_client_name     | string  | uaa username - usually `credhub_admin_client`                                                               |
| CH_UAA_CLIENT_SECRET   | credhub.uaa_client_secret   | string  | uaa client secret - "*Credhub Admin Client Credentials*" from *Operations Manager > PAS > Credentials* tab. |
| CH_SKIP_SSL_VALIDATION | credhub.skip_ssl_validation | boolean | skip SSL validation if true                                                                                 |
| CH_CA_CERT_FILE        | credhub.ca_cert_file        | path    | path to cert file                                                                                           |

## Brokerpak Configuration

Brokerpak configuration values:

| Environment Variable                                   | Config File Value                         | Type   | Description                                                          |
|--------------------------------------------------------|-------------------------------------------|--------|----------------------------------------------------------------------|
| <tt>GSB_BROKERPAK_BUILTIN_PATH</tt>                    | brokerpak.builtin.path                    | string | <p>Path to search for .brokerpak files, default: <code>./</code></p> |
| <tt>GSB_BROKERPAK_CONFIG</tt>                          | brokerpak.config                          | string | JSON global config for broker pak services                           |
| <tt>GSB_PROVISION_DEFAULTS</tt>                        | provision.defaults                        | string | JSON global provision defaults                                       |
| <tt>GSB_SERVICE_*SERVICE_NAME*_PROVISION_DEFAULTS</tt> | service.*service-name*.provision.defaults | string | JSON provision defaults override for *service-name*                  |
| <tt>GSB_SERVICE_*SERVICE_NAME*_PLANS</tt>              | service.*service-name*.plans              | string | JSON plan collection to augment plans for *service-name*             |
| <tt>GSB_COMPATIBILITY_ENABLE_BETA_SERVICES</tt>        | compatibility.enable-beta-services        | bool   | Enable services tagged with `beta`. Default: `false`                 |

## AWS Configuration

The AWS brokerpak supports default values for access key id and secret access key credentials.

| Environment Variable  | Config File Value     | Type   | Description       |
|-----------------------|-----------------------|--------|-------------------|
| AWS_ACCESS_KEY_ID     | aws.access_key_id     | string | access key id     |
| AWS_SECRET_ACCESS_KEY | aws.secret_access_key | string | secret access key |

### Config Example

```yml
aws:
  access_key_id: access key id
  secret_access_key: secret access key
db:
  host: your mysql host
  password: your mysql password
  user: your mysql username
api:
  user: someusername
  password: somepassword
  host: ...
  port: ...
credhub:
  url: ...
  uaa_url: ...
  uaa_client_name: ...
  uaa_client_secret: ...
 ```

### Global Config Example

Services for a given IaaS should have common parameter names for service wide platform resources (like regions)

AWS services support global region and VPC ID:

```yaml
provision:
  defaults: '{
    "region": "us-west-1",
    "aws_vpc_id": "vpc-093f61a410460f34c"
  }'
```

### Plans Example

Plans can be added to the brokerpak configuration on top of the ones provided:

```yaml
service:
  csb-aws-postgresql:
    plans: '[
      {
        "name":"small",
        "id":"85b27a04-8695-11ea-818a-274131861b81",
        "description":"additional plan with 4 cores, version 11 and 10 gb",
        "display_name":"small",
        "cores":"4",
        "postgres_version": 11,
        "storage_gb":10
      }
    ]'
  ```

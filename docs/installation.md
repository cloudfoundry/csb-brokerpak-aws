# Installing the broker on AWS

The broker service and the AWS brokerpak can be pushed and registered on a foundation running on AWS.

Documentation for broker configuration can be found [here](./configuration.md).

## Requirements

### CloudFoundry running on AWS.
The AWS brokerpak services are provisioned with firewall rules that only allow internal connectivity.
This allows `cf push`ed applications access, while denying any public access.

### AWS Service Credentials

The services need to be provisioned in the same AWS account that the foundation is running in.
To do this, the broker needs the following service principal credentials to manage resources within that account:
- access key id
- secret access key

#### Required IAM Policies
The AWS account represented by the access key needs the following permission policies:
```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": [
                "dynamodb:CreateTable",
                "dynamodb:CreateTableReplica",
                "dynamodb:DeleteTable",
                "dynamodb:DescribeBackup",
                "dynamodb:DescribeContinuousBackups",
                "dynamodb:DescribeTable",
                "dynamodb:DescribeTimeToLive",
                "dynamodb:ListTables",
                "dynamodb:ListTagsOfResource",
                "dynamodb:TagResource",
                "dynamodb:UntagResource",
                "ec2:AuthorizeSecurityGroupIngress",
                "ec2:AuthorizeSecurityGroupEgress",
                "ec2:CreateSecurityGroup",
                "ec2:DeleteSecurityGroup",
                "ec2:DescribeNetworkInterfaces",
                "ec2:DescribeRouteTables",
                "ec2:DescribeSecurityGroups",
                "ec2:DescribeSecurityGroupRules",
                "ec2:DescribeSubnets",
                "ec2:DescribeVpcAttribute",
                "ec2:DescribeVpcs",
                "ec2:RevokeSecurityGroupEgress",
                "ec2:RevokeSecurityGroupIngress",
                "elasticache:AddTagsToResource",
                "elasticache:RemoveTagsFromResource",
                "elasticache:ListTagsForResource",
                "elasticache:CreateCacheSubnetGroup",
                "elasticache:CreateReplicationGroup",
                "elasticache:DeleteCacheSubnetGroup",
                "elasticache:DeleteReplicationGroup",
                "elasticache:DescribeCacheClusters",
                "elasticache:DescribeCacheSubnetGroups",
                "elasticache:DescribeReplicationGroups",
                "elasticache:IncreaseReplicaCount",
                "elasticache:DecreaseReplicaCount",
                "elasticache:ModifyReplicationGroup",
                "elasticache:ModifyReplicationGroupShardConfiguration",
                "iam:CreateAccessKey",
                "iam:CreateUser",
                "iam:DeleteAccessKey",
                "iam:DeleteUser",
                "iam:DeleteUserPolicy",
                "iam:GetAccountAuthorizationDetails",
                "iam:GetPolicy",
                "iam:GetUser",
                "iam:GetUserPolicy",
                "iam:ListAccessKeys",
                "iam:ListAttachedUserPolicies",
                "iam:ListGroupsForUser",
                "iam:ListPolicies",
                "iam:ListUserPolicies",
                "iam:PutUserPolicy",
                "kms:GenerateDataKey",
                "kms:Encrypt",
                "kms:DescribeKey",
                "kms:Decrypt",
                "kms:CreateGrant",
                "kms:RevokeGrant",
                "logs:CreateLogDelivery",
                "logs:CreateLogGroup",
                "logs:DescribeResourcePolicies",
                "logs:DescribeLogGroups",
                "logs:GetLogDelivery",
                "logs:DeleteLogDelivery",
                "logs:DeleteLogGroup",
                "logs:ListLogDeliveries",
                "logs:ListTagsLogGroup",
                "logs:ListTagsForResource",
                "logs:UntagLogGroup",
                "logs:UntagResource",
                "logs:TagResource",
                "logs:TagLogGroup",
                "logs:ListTagsForResource",
                "logs:PutResourcePolicy",
                "logs:PutRetentionPolicy",
                "logs:UpdateLogDelivery",
                "rds:AddTagsToResource",
                "rds:RemoveTagsFromResource",
                "rds:ListTagsForResource",
                "rds:CreateDBCluster",
                "rds:CreateDBClusterParameterGroup",
                "rds:CreateDBInstance",
                "rds:CreateDBParameterGroup",
                "rds:CreateDBSubnetGroup",
                "rds:DeleteDBCluster",
                "rds:DeleteDBClusterParameterGroup",
                "rds:DeleteDBInstance",
                "rds:DeleteDBParameterGroup",
                "rds:DeleteDBSnapshot",
                "rds:DeleteDBSubnetGroup",
                "rds:DescribeDBClusterParameters",
                "rds:DescribeDBClusters",
                "rds:DescribeDBClusterParameterGroups",
                "rds:DescribeDBInstances",
                "rds:DescribeDBParameters",
                "rds:DescribeDBParameterGroups",
                "rds:DescribeDBSnapshots",
                "rds:DescribeDBSubnetGroups",
                "rds:DescribeGlobalClusters",
                "rds:ModifyDBClusterParameterGroup",
                "rds:ModifyDBCluster",
                "rds:ModifyDBInstance",
                "rds:ModifyDBParameterGroup",
                "rds:DescribeDBEngineVersions",
                "secretsmanager:CancelRotateSecret",
                "secretsmanager:CreateSecret",
                "secretsmanager:DeleteSecret",
                "secretsmanager:DescribeSecret",
                "secretsmanager:GetSecretValue",
                "secretsmanager:PutSecretValue",
                "secretsmanager:RotateSecret",
                "secretsmanager:TagResource",
                "secretsmanager:UpdateSecret",
                "secretsmanager:UntagResource",
                "s3:BypassGovernanceRetention",
                "s3:BypassGovernanceRetention",
                "s3:CreateAccessPoint",
                "s3:CreateAccessPoint",
                "s3:CreateBucket",
                "s3:DeleteAccessPointPolicy",
                "s3:DeleteBucket",
                "s3:DeleteBucketPolicy",
                "s3:DeleteObject",
                "s3:GetAccelerateConfiguration",
                "s3:GetAccountPublicAccessBlock",
                "s3:GetBucketAcl",
                "s3:GetBucketCORS",
                "s3:GetBucketLogging",
                "s3:GetBucketObjectLockConfiguration",
                "s3:GetBucketOwnershipControls",
                "s3:GetBucketPolicy",
                "s3:GetBucketPolicyStatus",
                "s3:GetBucketPublicAccessBlock",
                "s3:GetBucketRequestPayment",
                "s3:GetBucketTagging",
                "s3:GetBucketVersioning",
                "s3:GetBucketWebsite",
                "s3:GetEncryptionConfiguration",
                "s3:GetLifecycleConfiguration",
                "s3:GetObject",
                "s3:GetObjectLegalHold",
                "s3:GetObjectRetention",
                "s3:GetReplicationConfiguration",
                "s3:ListBucket",
                "s3:PutAccessPointPolicy",
                "s3:PutAccountPublicAccessBlock",
                "s3:PutBucketAcl",
                "s3:PutBucketLogging",
                "s3:PutBucketObjectLockConfiguration",
                "s3:PutBucketOwnershipControls",
                "s3:PutBucketOwnershipControls",
                "s3:PutBucketPolicy",
                "s3:PutBucketPublicAccessBlock",
                "s3:PutBucketRequestPayment",
                "s3:PutBucketTagging",
                "s3:PutBucketVersioning",
                "s3:PutEncryptionConfiguration",
                "s3:PutObject",
                "s3:PutObjectAcl",
                "s3:PutObjectLegalHold",
                "s3:PutObjectRetention",
                "sqs:CreateQueue",
                "sqs:DeleteQueue",
                "sqs:ListQueueTags",
                "sqs:TagQueue",
                "sqs:UntagQueue",
                "sqs:GetQueueAttributes",
                "sqs:SetQueueAttributes",
                "sqs:GetQueueUrl",
                "sqs:ListQueues",
                "dms:CreateReplicationSubnetGroup"
            ],
            "Effect": "Allow",
            "Resource": "*"
        }
    ]
}
```

##### IAM Policies in Enhanced Monitoring

To enable the Enhanced Monitoring feature for Amazon RDS, it's necessary to grant additional permissions.
This feature requires permission to act on your behalf to send OS metric information to CloudWatch Logs.
You grant Enhanced Monitoring permissions using an AWS Identity and Access Management (IAM) role.

To configure the ARN for the IAM role that permits RDS to send enhanced monitoring metrics to CloudWatch Logs,
the user that you want to access Enhanced Monitoring needs a policy that includes a statement that allows the 
user to pass the role, like the following. Use your _account number_ and replace the _role name_ with 
the name of your role.

```json
{
        "Sid": "PolicyStatementToAllowUserToPassOneSpecificRole",
        "Effect": "Allow",
        "Action": [ "iam:PassRole" ],
        "Resource": "arn:aws:iam::account-id:role/RDS-Monitoring-Role-Name"
    }
```

To read about setting up and enabling Enhanced Monitoring see the
[AWS Documentation](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_Monitoring.OS.Enabling.html).

To read about granting a user permission to pass a role to an AWS service, see the
[AWS Documentation](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_use_passrole.html).

### MySQL Database for Broker State
The broker keeps service instance and binding information in a MySQL database. 

#### Binding a MySQL Database
If there is an existing broker in the foundation that can provision a MySQL instance use `cf create-service`
to create a new MySQL instance. Then use `cf bind-service` to bind that instance to the service broker.

#### Manually Provisioning a MySQL Database

The AWS Service Broker stores the state of provisioned resources in a MySQL database.
You may use any database compatible with the MySQL protocol.

If a MySQL instance needs to be manually provisioned, it must be accessible to applications running within the
foundation so that the `cf push`ed broker can access it.

The following configuration parameters will be needed:
- `DB_HOST`
- `DB_USERNAME`
- `DB_PASSWORD`

It is also necessary to create a database named `servicebroker` or the name set in the parameter `DB_NAME`
within that server (use your favorite tool to connect to the MySQL server and issue `CREATE DATABASE servicebroker;`).

## Step By Step From a Pre-build Release with a Bound MySQL Instance

Fetch a pre-built broker and brokerpak and bind it to a `cf create-service` managed MySQL.

### Requirements

The following tools are needed on your workstation:
- [cf cli](https://docs.cloudfoundry.org/cf-cli/install-go-cli.html)

### Assumptions

The `cf` CLI has been used to authenticate with a foundation (`cf api` and `cf login`,) and an org and space
have been targeted (`cf target`).

### Fetch A Broker and AWS Brokerpak

Download a Cloud Service Broker release from [GitHub](https://github.com/cloudfoundry/cloud-service-broker/releases).
Find the latest release matching the name pattern `vX.X.X`.
Change filename `cloud-service-broker.linux` to `cloud-service-broker`.
Add execution permissions `chmod +x cloud-service-broker`

Download an AWS Brokerpak release from [GitHub](https://github.com/cloudfoundry/csb-brokerpak-aws/releases).
Find the latest release matching the name pattern `X.X.X`.

Put the `cloud-service-broker` and `aws-services-X.X.X.brokerpak` into the same directory on your workstation.

### Create a MySQL instance with AWS broker

If there is an existing AWS broker in the foundation that can provision a MySQL instance use `cf create-service`
to create a new MySQL instance.
Then use `cf bind-service` to bind that instance to the service broker app.

The following command will create a basic MySQL database instance named `csb-sql`

```bash
cf create-service <MySQL_SERVICE_OFFERING_NAME> <PLAN_NAME> csb-sql [-b <SERVICE_BROKER_NAME>] 
```

### Build Config File
To avoid putting any sensitive information in environment variables, a config file can be used.

Create a file named `config.yml` in the same directory the broker and brokerpak have been downloaded to. Its contents should be:

```yaml
aws:
  access_key_id: your access key id
  secret_access_key: your secret access key

api:
  user: someusername
  password: somepassword
```

Add your custom plans to the `config.yml` file, for example, plans for MySQL

```yaml
service:
  csb-aws-mysql:
    plans: '[{"name":"default","id":"0f3522b2-f040-443b-bc53-4aed25284840","description":"Default MySQL plan","display_name":"default","instance_class":"db.m6g.large","mysql_version":"8.0","storage_gb":100}]'
```

### Push and Register the Broker

Push the broker as a binary application:

```bash
SECURITY_USER_NAME=someusername
SECURITY_USER_PASSWORD=somepassword
APP_NAME=cloud-service-broker

chmod +x cloud-service-broker
cf push "${APP_NAME}" -c './cloud-service-broker serve --config config.yml' -b binary_buildpack --random-route --no-start
```

Bind the MySQL database and start the service broker:
```bash
cf bind-service cloud-service-broker csb-sql
cf start "${APP_NAME}"
```

Register the service broker:
```bash
BROKER_NAME=csb-$USER

cf create-service-broker "${BROKER_NAME}" "${SECURITY_USER_NAME}" "${SECURITY_USER_PASSWORD}" https://$(LANG=EN cf app "${APP_NAME}" | grep 'routes:' | cut -d ':' -f 2 | xargs) --space-scoped || cf update-service-broker "${BROKER_NAME}" "${SECURITY_USER_NAME}" "${SECURITY_USER_PASSWORD}" https://$(LANG=EN cf app "${APP_NAME}" | grep 'routes:' | cut -d ':' -f 2 | xargs)
```

Once this completes, the output from `cf marketplace` should include:

```
csb-aws-mysql    default            Default MySQL plan
```

## Step By Step From a Pre-built Release with a Manually Provisioned MySQL Instance

Fetch a pre-built broker and brokerpak and configure with a manually provisioned MySQL instance.

Requirements and assumptions are the same as above. Follow instructions above to [fetch the broker and brokerpak](#Fetch-A-Broker-and-AWS-Brokerpak)

### Create a MySQL Database

It's an exercise for the reader to create a MySQL server somewhere that a `cf push`ed app can access.
The database connection values (hostname, username and password) will be needed in the next step.
It is also necessary to create a database named `servicebroker` within that server (use your favorite tool to
connect to the MySQL server and issue `CREATE DATABASE servicebroker;`).

### Build Config File
To avoid putting any sensitive information in environment variables, a config file can be used.

Create a file named `config.yml` in the same directory the broker and brokerpak have been downloaded to. Its contents should be:

```yaml
aws:
  access_key_id: your access key id
  secret_access_key: your secret access key

db:
  host: your mysql host
  password: your mysql password
  user: your mysql username

api:
  user: someusername
  password: somepassword

service:
  csb-aws-mysql:
    plans: '[{"name":"default","id":"0f3522b2-f040-443b-bc53-4aed25284840","description":"Default MySQL plan","display_name":"default","instance_class":"db.m6g.large","mysql_version":"8.0","storage_gb":100}]'
```

Add your custom plans to the `config.yml` file, for example, plans for MySQL

Push and Register the Broker, see [previous section](#Push-and-Register-the-Broker)

Once these steps are complete, the output from `cf marketplace` should resemble the same as above.

## Step By Step From Source with Bound MySQL
Grab the source code, build and deploy.

### Requirements

The following tools are needed on your workstation:
- [The latest GoLang version](https://golang.org/dl/)
- [make](https://www.gnu.org/software/make/)
- [cf cli](https://docs.cloudfoundry.org/cf-cli/install-go-cli.html)

The Cloud Service Broker for AWS must be installed in your foundation.

### Assumptions

The `cf` CLI has been used to authenticate with a foundation (`cf api` and `cf login`,) and an org and space
have been targeted (`cf target`).

### Clone the Repo

The following commands will clone the service broker repository and cd into the resulting directory:
```bash
git clone https://github.com/cloudfoundry/cloud-service-broker.git
cd cloud-service-broker
```
### Set Required Environment Variables

Collect the AWS service credentials for your account and set them as environment variables:
```bash
export AWS_SECRET_ACCESS_KEY=your secret access key
export AWS_ACCESS_KEY_ID=your access key id

```
Generate username and password for the broker - Cloud Foundry will use these credentials to authenticate API calls to
the service broker.
```bash
export SECURITY_USER_NAME=someusername
export SECURITY_USER_PASSWORD=somepassword
```
### Create a MySQL instance

It's an exercise for the reader to create a MySQL server somewhere that a `cf push`ed app can access.
If there is an existing AWS broker in the foundation that can provision a MySQL instance use `cf create-service`
to create a new MySQL instance.
Then use `cf bind-service` to bind that instance to the service broker app.

The following command will create a basic MySQL database instance named `csb-sql`

```bash
cf create-service <MySQL_SERVICE_OFFERING_NAME> <PLAN_NAME> csb-sql [-b <SERVICE_BROKER_NAME>] 
```

### Use the Makefile to Deploy the Broker
There is a make target that will build the broker and brokerpak and deploy to and register with Cloud Foundry
as a space scoped broker. This will be local and private to the org and space your `cf` CLI is targeting.

```bash
make push-broker
```

Once these steps are complete, the output from `cf marketplace` should resemble the same as above.

## Step By Step Slightly Harder Way

Requirements and assumptions are the same as above.
Follow instructions for the first two steps above ([Clone the Repo](#Clone-the-Repo) and
[Set Required Environment Variables](#Set-Required-Environment-Variables)).

### Create a MySQL Database

It's an exercise for the reader to create a MySQL server somewhere that a `cf push`ed app can access.
The database connection values (hostname, username and password) will be needed in the next step.
It is also necessary to create a database named `servicebroker` within that server (use your favorite tool to
connect to the MySQL server and issue `CREATE DATABASE servicebroker;`).
Set the following environment variables with information about that MySQL instance:

```bash
export DB_HOST=mysql server host
export DB_USERNAME=mysql server username
export DB_PASSWORD=mysql server password
```

### Build the Broker and Brokerpak

Use the makefile to build the broker executable and brokerpak.
```bash
make cloud-service-broker
make build
```

### Pushing the Broker

All the steps to push and register the broker:
```bash
APP_NAME=cloud-service-broker

cf push --no-start

cf set-env "${APP_NAME}" SECURITY_USER_PASSWORD "${SECURITY_USER_PASSWORD}"
cf set-env "${APP_NAME}" SECURITY_USER_NAME "${SECURITY_USER_NAME}"

cf set-env "${APP_NAME}" AWS_ACCESS_KEY_ID "${AWS_ACCESS_KEY_ID}"
cf set-env "${APP_NAME}" AWS_SECRET_ACCESS_KEY "${AWS_SECRET_ACCESS_KEY}"

cf set-env "${APP_NAME}" DB_HOST "${DB_HOST}"
cf set-env "${APP_NAME}" DB_USERNAME "${DB_USERNAME}"
cf set-env "${APP_NAME}" DB_PASSWORD "${DB_PASSWORD}"

cf set-env "${APP_NAME}" GSB_BROKERPAK_BUILTIN_PATH ./

cf start "${APP_NAME}"

BROKER_NAME=csb-$USER

cf create-service-broker "${BROKER_NAME}" "${SECURITY_USER_NAME}" "${SECURITY_USER_PASSWORD}" https://$(LANG=EN cf app "${APP_NAME}" | grep 'routes:' | cut -d ':' -f 2 | xargs) --space-scoped || cf update-service-broker "${BROKER_NAME}" "${SECURITY_USER_NAME}" "${SECURITY_USER_PASSWORD}" https://$(LANG=EN cf app "${APP_NAME}" | grep 'routes:' | cut -d ':' -f 2 | xargs)
```

Once these steps are complete, the output from `cf marketplace` should resemble the same as above.

## Uninstalling the Broker
First, make sure there are all service instances created with `cf create-service` have been destroyed
with `cf delete-service` otherwise removing the broker will fail.

### Unregister the Broker
```bash
cf delete-service-broker csb-$USER
```

### Uninstall the Broker
```bash
cf delete cloud-service-broker
```



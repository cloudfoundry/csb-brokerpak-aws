# PostgreSQL Plans and Config

These are the default plans and configuration options for PostgreSQL across the supported cloud platforms (AWS, Azure and GCP.)

## Plans

| Plan | Version | CPUs | Memory Size | Disk Size |
|------|---------|------|-------------|-----------|
|small | 11      | 2    | min 4GB     | 5GB       |
|medium| 11      | 4    | min 8GB     | 10GB      |
|large | 11      | 8    | min 16GB    | 20GB      |


## Configuration Options

The following options can be configured across all supported platforms. Notes below document any platform specific information for mapping that might be required.

| Option Name | Values | Default |
|-------------|--------|---------|
| postgres_version | 9.5, 9.6, 10, 11 | 11    |
| storage_gb  | 5 - 4096| 5      |
| cores       | 1,2,4,8,16,32,64 | 1      |
| db_name     | | csb-db |

### AWS Notes - applies to *csb-aws-postgresql*

CPU/memory size mapped into [AWS DB instance types](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.DBInstanceClass.html) as follows:

| Plan  | Instance class |
|-------|----------|
| small | db.t2.medium |
| medium | db.m4.xlarge |
| large | db.m4.2xlarge |
| subsume | existing posgresql db |

#### Core to instance class mapping

| Cores | Instance class |
|-------|---------------|
| 1     | db.t2.small  |
| 2     | db.t3.medium  |
| 4     | db.m5.xlarge  |
| 8     | db.m5.2xlarge |
| 16    | db.m5.4xlarge |
| 32    | db.m5.8xlarge |
| 64    | db.m5.16xlarge|

#### AWS specific config parameters

The following parameters (as well as those above) may be configured during service provisioning (`cf create-service csb-aws-postgresql ... -c '{...}'`

| Parameter | Type | Description | Default |
|-----------|------|------|---------|
| instance_name | string | name of AWS instance to create | csb-mysql-*instance_id* |
| region  | string | [AWS region](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html#concepts-available-regions) to deploy service  | us-west-2 |
| aws_vpc_id | string | The VPC to connect the instance to | the default vpc |
| aws_access_key_id | string | ID of AWS access key for instance | config file value `aws.access_key_id` |
| aws_secret_access_key | string | ID of AWS access key secret for instance | config file value `aws.secret_access_key` |
| instance_class | string | explicit [instance class](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.DBInstanceClass.html) *overrides* `cores` conversion into instance class per table above | | 
| multi-az | boolean | If `true`, create multi-az ([HA](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.MultiAZ.html)) instance | `false` | 
| publicly_accessible | boolean | If `true`, make instance available to public connections | `false ` |
| storage_autoscale | boolean | If `true`, storage will autoscale to max of *storage_autoscale_limit_gb* | `false` |
| storage_autoscale_limit_gb | number | if *storage_autoscale* is `true`, max size storage will scale up to ||
| storage_encrypted | boolean | If `true`, DB storage will be encrypted | `false`|
| parameter_group_name | string | PostgreSQL parameter group name for instance | `default.postgres.<postgres version>` |
| rds_subnet_group | string | Name of subnet to attach DB instance to, overrides *aws_vpc_id* | |
| rds_vpc_security_group_ids | comma delimited string | Security group ID's to assign to DB instance | |
| use_tls | boolean |Use TLS for DB connections | `true` |
| allow_major_version_upgrade | bool | Indicates that major version upgrades are allowed. | `true` |
| auto_minor_version_upgrade  | bool | Indicates that minor engine upgrades will be applied automatically to the DB instance during the maintenance window| `true` |
| maintenance_day | integer | Day of week for maintenance window | See the [AWS documentation](http://docs.aws.amazon.com/cli/latest/reference/rds/create-db-instance.html) |
| maintenance_start_hour | integer | Start hour for maintenance window | See the [AWS documentation](http://docs.aws.amazon.com/cli/latest/reference/rds/create-db-instance.html)|
| maintenance_start_min | integer | Start minute for maintenance window | See the [AWS documentation](http://docs.aws.amazon.com/cli/latest/reference/rds/create-db-instance.html)|
| maintenance_end_hour | integer | End hour for maintenance window | See the [AWS documentation](http://docs.aws.amazon.com/cli/latest/reference/rds/create-db-instance.html)|
| maintenance_end_min | integer | End minute for maintenance window | See the [AWS documentation](http://docs.aws.amazon.com/cli/latest/reference/rds/create-db-instance.html)|


#### Subsume Parameters
| Parameter | Type | Description |
|-----------|------|------|
| aws_db_id | string | The AWS resource ID for the postgresql DB to subsume |



## Binding Credentials

The binding credentials for PostgreSQL have the following shape:

```json
{
    "name" : "database name",
    "hostname" : "database server host",
    "port" : "database server port",
    "username" : "authentication user name",
    "password" : "authentication password",
    "uri" : "database connection URI",
    "jdbcUrl" : "jdbc format connection URI"
}
```

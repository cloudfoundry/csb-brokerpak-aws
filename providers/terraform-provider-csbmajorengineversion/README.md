# terraform-provider-csbmajorengineversion

Terraform provider designed to get the RDS major version given an engine and a reference version.

For example, suppose we want to know the best version of our RDS instance, knowing that we are using
the `aurora-mysql` engine, and specifically, `aurora-mysql` engine version `8.0.mysql_aurora.3.03.1`;
the provider will return version the Major Version `8.0`.

```terraform

provider "csbmajorengineversion" {
  engine            = "aurora-mysql"
  access_key_id     = "XXXXXXXXX"
  secret_access_key = "XXXXXXXXX"
}

data "csbmajorengineversion" "major_version" {
  engine_version = "8.0.mysql_aurora.3.03.1"
}

# Result

data "csbmajorengineversion" "major_version" {
  engine_version = "8.0.mysql_aurora.3.03.1"
  major_version  = "8.0"
}
```

## Argument Reference

The following arguments are supported:

* `access_key_id`: (Required) AWS access key 
* `secret_access_key`: (Required) AWS secret key
* `engine`: (Required) The database engine to use. For supported values, see the Engine parameter in
  [API action CreateDBInstance](https://docs.aws.amazon.com/AmazonRDS/latest/APIReference/API_CreateDBInstance.html).
* `engine_version`: (Required) The engine version of your current RDS instance.

In addition to all arguments above, the following attributes are exported:

* `major_version`: The major engine version.

## Mandatory Permissions

* `rds:DescribeDBEngineVersions`: Grants permission to return a list of the available DB engines.
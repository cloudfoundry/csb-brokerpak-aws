variable "aws_access_key_id" { type = string }
variable "aws_secret_access_key" { type = string }
variable "labels" { type = map(any) }

variable "region" {
  type = string
}

variable "instance_name" { type = string }
variable "instance_class" {
  type = string
}

variable "db_name" {
  type = string
}

variable "aws_vpc_id" {
  type = string
}

variable "rds_subnet_group" {
  type = string
}

variable "rds_vpc_security_group_ids" {
  type = string
}

variable "mssql_version" {
  type = string
}

variable "engine" {
  type = string

  validation {
    condition = anytrue([
      var.engine == "sqlserver-ee",
      var.engine == "sqlserver-ex",
      var.engine == "sqlserver-se",
      var.engine == "sqlserver-web",
    ])
    error_message = "engine not in the list of supported values: sqlserver-ee, sqlserver-ex, sqlserver-se, sqlserver-web"
  }
}

variable "storage_gb" {
  type = number

  validation {
    condition     = var.storage_gb >= 20
    error_message = "storage_gb must be >= 20"
  }
}

variable "aws_access_key_id" { type = string }
variable "aws_secret_access_key" { type = string }
variable "labels" { type = map(any) }

variable "region" {
  type    = string
  default = "us-west-2"
}

variable "instance_name" { type = string }
variable "instance_class" {
  type    = string
  default = ""
}

variable "db_name" {
  type    = string
  default = "vsbdb"
}

variable "aws_vpc_id" {
  type    = string
  default = ""
}

variable "rds_subnet_group" {
  type    = string
  default = ""
}

variable "rds_vpc_security_group_ids" {
  type    = string
  default = ""
}

variable "mssql_version" {
  type    = string
  default = null
}

variable "engine" {
  type    = string
  default = "sqlserver-ee"

  validation {
    condition     = contains(["sqlserver-ee", "sqlserver-ex", "sqlserver-se", "sqlserver-web"], var.engine)
    error_message = "engine not in the list of supported values: sqlserver-ee, sqlserver-ex, sqlserver-se, sqlserver-web"
  }
}

variable "storage_gb" {
  type    = number
  default = 20

  validation {
    condition     = var.storage_gb >= 20 && var.storage_gb <= 4096
    error_message = "storage_gb must be >= 20 and <= 4096"
  }
}

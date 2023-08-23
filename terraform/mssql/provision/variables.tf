variable "aws_access_key_id" { type = string }
variable "aws_secret_access_key" { type = string }
variable "region" { type = string }
variable "instance_name" { type = string }
variable "storage_encrypted" { type = bool }
variable "kms_key_id" { type = string }
variable "db_name" { type = string }
variable "labels" { type = map(any) }
variable "instance_class" { type = string }
variable "aws_vpc_id" { type = string }
variable "rds_subnet_group" { type = string }
variable "rds_vpc_security_group_ids" { type = string }
variable "mssql_version" { type = string }
variable "engine" { type = string }
variable "storage_gb" { type = number }
variable "deletion_protection" { type = bool }
variable "max_allocated_storage" {
  type     = number
  nullable = true
}
variable "storage_type" { type = string }
variable "iops" { type = number }
variable "option_group_name" { type = string }
variable "publicly_accessible" { type = bool }

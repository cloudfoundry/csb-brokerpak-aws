variable "aws_access_key_id" {
  type      = string
  sensitive = true
}
variable "aws_secret_access_key" {
  type      = string
  sensitive = true
}
variable "region" { type = string }
variable "name" { type = string }
variable "reader_endpoint" { type = bool }
variable "hostname" { type = string }
variable "reader_hostname" { type = string }
variable "admin_username" { type = string }
variable "admin_password" {
  type      = string
  sensitive = true
}
variable "port" { type = number }
variable "use_managed_admin_password" { type = bool }
variable "managed_admin_credentials_arn" { type = string }
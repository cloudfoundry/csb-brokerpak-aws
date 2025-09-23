variable "aws_access_key_id" {
  type      = string
  sensitive = true
}
variable "aws_secret_access_key" {
  type      = string
  sensitive = true
}
variable "region" { type = string }
variable "db_name" {
  type = string
}
variable "hostname" {
  type = string
}
variable "admin_username" {
  type = string
}
variable "admin_password" {
  type      = string
  sensitive = true
}
variable "require_ssl" {
  type = bool
}
variable "use_managed_admin_password" {
  type = string
}
variable "managed_admin_credentials_arn" {
  type = string
}
variable "port" {
  type = number
}
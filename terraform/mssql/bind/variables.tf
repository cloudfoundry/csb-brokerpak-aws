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

locals {
  port = 1433
}
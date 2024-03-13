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
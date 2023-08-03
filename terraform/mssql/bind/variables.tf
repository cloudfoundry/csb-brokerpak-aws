variable "db_name" {
  type = string
  default = "my-db"
}
variable "hostname" {
  type = string
  default = "database-azucgar.crvbjnvu3aun.us-west-2.rds.amazonaws.com"
}
variable "admin_username" {
  type = string
  default = "admin"
}
variable "admin_password" {
  type = string
  default = "12345678"
}

locals {
  port = 1433
}
variable "aws_access_key_id" {
  type      = string
  sensitive = true
}

variable "aws_secret_access_key" {
  type      = string
  sensitive = true
}
variable "region" { type = string }
variable "user_name" { type = string }
variable "prefix" { type = string }
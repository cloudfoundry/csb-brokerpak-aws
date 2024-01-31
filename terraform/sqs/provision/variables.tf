variable "aws_access_key_id" {
  type      = string
  sensitive = true
}
variable "aws_secret_access_key" {
  type      = string
  sensitive = true
}
variable "region" { type = string }

variable "instance_name" { type = string }
variable "labels" { type = map(any) }

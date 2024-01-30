variable "aws_access_key_id" { type = string }
variable "aws_secret_access_key" { type = string }
variable "region" { type = string }

variable "instance_name" { type = string }
variable "labels" { type = map(any) }

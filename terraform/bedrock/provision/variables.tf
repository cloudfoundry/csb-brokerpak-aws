variable "instance_name" { type = string }
variable "region" { type = string }
variable "ttl_hours" { type = number }
variable "models" { type = string }
variable "labels" { type = map(any) }

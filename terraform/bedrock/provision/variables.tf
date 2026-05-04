variable "instance_name" { type = string }
variable "region" { type = string }
variable "ttl_hours" { type = number }
variable "models" { type = string }
variable "budget_amount" { type = number }
variable "budget_alert_email" {
	type    = string
	default = ""
}
variable "labels" { type = map(any) }

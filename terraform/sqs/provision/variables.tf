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
variable "fifo" { type = bool }
variable "visibility_timeout_seconds" { type = number }
variable "message_retention_seconds" { type = number }
variable "max_message_size" { type = number }
variable "delay_seconds" { type = number }
variable "receive_wait_time_seconds" { type = number }
variable "dlq_arn" { type = string }
variable "max_receive_count" { type = number }
variable "deduplication_scope" { type = string }
variable "fifo_throughput_limit" { type = string }


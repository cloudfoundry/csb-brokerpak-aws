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
variable "content_based_deduplication" { type = bool }
variable "deduplication_scope" { type = string }
variable "fifo_throughput_limit" { type = string }
variable "sqs_managed_sse_enabled" { type = bool }
variable "kms_master_key_id" { type = string }
variable "kms_data_key_reuse_period_seconds" { type = number }
variable "kms_extra_key_ids" { type = string }
resource "aws_sqs_queue" "queue" {
  name                       = var.fifo ? "${var.instance_name}.fifo" : var.instance_name
  fifo_queue                 = var.fifo
  visibility_timeout_seconds = var.visibility_timeout_seconds
  message_retention_seconds  = var.message_retention_seconds
  max_message_size           = var.max_message_size
  delay_seconds              = var.delay_seconds
  receive_wait_time_seconds  = var.receive_wait_time_seconds

  redrive_policy = var.dlq_arn != "" ? jsonencode({
    deadLetterTargetArn = var.dlq_arn,
    maxReceiveCount     = var.max_receive_count
  }) : null

  deduplication_scope   = var.deduplication_scope
  fifo_throughput_limit = var.fifo_throughput_limit

  # Server-side encryption settings
  kms_master_key_id                 = var.kms_master_key_id == "" ? null : var.kms_master_key_id
  kms_data_key_reuse_period_seconds = var.kms_data_key_reuse_period_seconds

  sqs_managed_sse_enabled = !var.sqs_managed_sse_enabled ? null : var.sqs_managed_sse_enabled

  lifecycle {
    prevent_destroy = true
  }
}
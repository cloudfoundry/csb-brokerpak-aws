resource "aws_sqs_queue" "queue" {
  name       = var.fifo ? "${var.instance_name}.fifo" : var.instance_name
  fifo_queue = var.fifo

  redrive_policy = var.dlq_arn != "" ? jsonencode({
    deadLetterTargetArn = var.dlq_arn,
    maxReceiveCount     = var.max_receive_count
  }) : null

  lifecycle {
    prevent_destroy = true
  }
}
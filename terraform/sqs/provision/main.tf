resource "aws_sqs_queue" "queue" {
  name       = var.fifo ? "${var.instance_name}.fifo" : var.instance_name
  fifo_queue = var.fifo
  tags       = var.labels
  visibility_timeout_seconds = var.visibility_timeout_seconds 

  lifecycle {
    prevent_destroy = true
  }
}
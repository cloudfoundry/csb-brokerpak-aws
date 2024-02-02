resource "aws_sqs_queue" "queue" {
  name       = var.fifo ? "${var.instance_name}.fifo" : var.instance_name
  fifo_queue = var.fifo
  tags       = var.labels

  lifecycle {
    prevent_destroy = true
  }
}
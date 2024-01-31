resource "aws_sqs_queue" "queue" {
  name       = var.instance_name
  fifo_queue = false
  tags       = var.labels
}
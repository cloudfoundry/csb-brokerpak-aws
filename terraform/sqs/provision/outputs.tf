output "arn" { value = aws_sqs_queue.queue.arn }
output "region" { value = var.region }
output "queue_url" { value = aws_sqs_queue.queue.id }
output "queue_name" { value = aws_sqs_queue.queue.name }
output "arn" { value = aws_sqs_queue.queue.arn }
output "url" { value = aws_sqs_queue.queue.id }
output "name" { value = aws_sqs_queue.queue.name }
output "region" { value = var.region }
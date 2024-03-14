output "arn" { value = aws_sqs_queue.queue.arn }
output "region" { value = var.region }
output "queue_url" { value = aws_sqs_queue.queue.id }
output "queue_name" { value = aws_sqs_queue.queue.name }
output "dlq_arn" { value = var.dlq_arn }
output "kms_all_key_ids" { value = local.kms_all_key_ids }
output "status" {
  value = format(
    "created SQS queue: %s (ARN: %s)",
    aws_sqs_queue.queue.id,
    aws_sqs_queue.queue.arn
  )
}

# Execution instructions

Working folder: `~/workspace/csb/csb-brokerpak-aws/terraform/sqs/provision/dlq`

1. Check AWS environment variables are set:
```shell
env | grep AWS
AWS_ACCESS_KEY_ID=XXXXXXXX
AWS_SECRET_ACCESS_KEY=XXXXXX
```

3. Create Standard and DLQ queues
```shell
terraform init
```

```shell
bash ./create_infra.sh
```
Output:
```shell                        
Terraform used the selected providers to generate the following execution plan. Resource actions are indicated with the following symbols:
  + create

Terraform will perform the following actions:

  # aws_sqs_queue.my_dlq will be created
  + resource "aws_sqs_queue" "my_dlq" {
      ....
      + visibility_timeout_seconds        = 30
    }

  # aws_sqs_queue.my_queue will be created
  + resource "aws_sqs_queue" "my_queue" {
      ...
      + redrive_policy                    = (known after apply)
      ...
    }

Plan: 2 to add, 0 to change, 0 to destroy.

Changes to Outputs:
  + my_dlq_url   = (known after apply)
  + my_queue_url = (known after apply)
aws_sqs_queue.my_dlq: Creating...
aws_sqs_queue.my_dlq: Still creating... [10s elapsed]
aws_sqs_queue.my_dlq: Creation complete after 27s [id=https://sqs.us-west-2.amazonaws.com/XXXXXXXX/my-dlq]
aws_sqs_queue.my_queue: Creating...
aws_sqs_queue.my_queue: Still creating... [10s elapsed]
aws_sqs_queue.my_queue: Creation complete after 27s [id=https://sqs.us-west-2.amazonaws.com/XXXXXXXX/my-queue]

Apply complete! Resources: 2 added, 0 changed, 0 destroyed.

Outputs:

my_dlq_url = "https://sqs.us-west-2.amazonaws.com/XXXXXXXX/my-dlq"
my_queue_url = "https://sqs.us-west-2.amazonaws.com/XXXXXXXX/my-queue"

```

3. Execute GoLang app

```shell
bash ./execute.sh
```
Output
```shell
2024/02/12 16:12:43 Message received: Incorrect format message
2024/02/12 16:12:43 Message format incorrect, leaving in queue for DLQ
2024/02/12 16:12:43 Message sent: This is a correctly formatted message.
2024/02/12 16:12:43 Message received: This is a correctly formatted message.
2024/02/12 16:12:44 Message processed and deleted.
2024/02/12 16:12:44 Message received: Incorrect format message
2024/02/12 16:12:44 Message format incorrect, leaving in queue for DLQ
2024/02/12 16:12:44 Message sent: Incorrect format message
2024/02/12 16:12:44 Message received: Incorrect format message
2024/02/12 16:12:44 Message format incorrect, leaving in queue for DLQ
2024/02/12 16:12:45 Message received: Incorrect format message
2024/02/12 16:12:45 Message format incorrect, leaving in queue for DLQ
2024/02/12 16:12:45 Message received: Incorrect format message
2024/02/12 16:12:45 Message format incorrect, leaving in queue for DLQ
2024/02/12 16:12:46 Message received: Incorrect format message
2024/02/12 16:12:46 Message format incorrect, leaving in queue for DLQ
2024/02/12 16:12:46 Message received: Incorrect format message
2024/02/12 16:12:46 Message format incorrect, leaving in queue for DLQ
2024/02/12 16:12:47 DLQ Message logged: Incorrect format message
2024/02/12 16:12:47 DLQ Subscriber stopped due to context cancellation.
2024/02/12 16:12:47 Application stopped.
```
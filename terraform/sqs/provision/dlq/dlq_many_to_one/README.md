
# Execution instructions

Working folder: `~/workspace/csb/csb-brokerpak-aws/terraform/sqs/provision/dlq-several-subscribers`

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
....


Plan: 3 to add, 0 to change, 0 to destroy.

Changes to Outputs:
  + my_dlq_url       = (known after apply)
  + my_queue_two_url = (known after apply)
  + my_queue_url     = (known after apply)
.....

Apply complete! Resources: 3 added, 0 changed, 0 destroyed.

Outputs:

my_dlq_url = "https://sqs.us-west-2.amazonaws.com/XXXXXXXXX/my-dlq-several-subscribers"
my_queue_two_url = "https://sqs.us-west-2.amazonaws.com/XXXXXXXXX/my-queue-two-several-subscribers"
my_queue_url = "https://sqs.us-west-2.amazonaws.com/XXXXXXXXX/my-queue-several-subscribers"


```

3. Execute GoLang app

```shell
bash ./execute.sh
```
Output
```shell
2024/02/13 13:53:57 DLQ Message logged: Incorrect format message - queue name: queue_two
2024/02/13 13:53:57 Message sent: This is a correctly formatted message.
2024/02/13 13:53:57 Message sent: This is a correctly formatted message.
2024/02/13 13:53:57 Message received: This is a correctly formatted message. - queue queue_one
2024/02/13 13:53:57 Message received: This is a correctly formatted message. - queue queue_two
2024/02/13 13:53:57 Message processed and deleted.
2024/02/13 13:53:57 DLQ Message logged: Incorrect format message - queue name: queue_one
2024/02/13 13:53:57 Message processed and deleted.
2024/02/13 13:53:57 DLQ Message logged: Incorrect format message - queue name: queue_one
2024/02/13 13:53:57 DLQ Message logged: Incorrect format message - queue name: queue_two
2024/02/13 13:53:58 Message sent: Incorrect format message
2024/02/13 13:53:58 Message received: Incorrect format message - queue queue_one
2024/02/13 13:53:58 Message format incorrect, leaving in queue for DLQ
2024/02/13 13:53:58 Message sent: Incorrect format message
2024/02/13 13:53:58 Message received: Incorrect format message - queue queue_two
2024/02/13 13:53:58 Message format incorrect, leaving in queue for DLQ
2024/02/13 13:53:59 Message received: Incorrect format message - queue queue_one
2024/02/13 13:53:59 Message format incorrect, leaving in queue for DLQ
2024/02/13 13:53:59 Message received: Incorrect format message - queue queue_two
2024/02/13 13:53:59 Message format incorrect, leaving in queue for DLQ
2024/02/13 13:54:00 Message received: Incorrect format message - queue queue_one
2024/02/13 13:54:00 Message format incorrect, leaving in queue for DLQ
2024/02/13 13:54:00 Message received: Incorrect format message - queue queue_two
2024/02/13 13:54:00 Message format incorrect, leaving in queue for DLQ
2024/02/13 13:54:01 Message received: Incorrect format message - queue queue_one
2024/02/13 13:54:01 Message format incorrect, leaving in queue for DLQ
2024/02/13 13:54:01 Message received: Incorrect format message - queue queue_two
2024/02/13 13:54:01 Message format incorrect, leaving in queue for DLQ
2024/02/13 13:54:02 Message received: Incorrect format message - queue queue_one
2024/02/13 13:54:02 Message format incorrect, leaving in queue for DLQ
2024/02/13 13:54:02 Message received: Incorrect format message - queue queue_two
2024/02/13 13:54:02 Message format incorrect, leaving in queue for DLQ
2024/02/13 13:54:03 DLQ Message logged: Incorrect format message - queue name: queue_one
2024/02/13 13:54:03 DLQ Message logged: Incorrect format message - queue name: queue_two
2024/02/13 13:54:12 No messages found
2024/02/13 13:54:12 No messages found
2024/02/13 13:54:16 Application stopped due to timeout.
```
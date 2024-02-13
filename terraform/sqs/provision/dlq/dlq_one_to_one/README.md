
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

....

Apply complete! Resources: 5 added, 0 changed, 0 destroyed.

Outputs:

my_dlq_url = "https://sqs.us-west-2.amazonaws.com/XXXXXXXX/my-dlq"
my_queue_url = "https://sqs.us-west-2.amazonaws.com/XXXXXXXX/my-queue"
user_access_key_id = <sensitive>
user_secret_access_key = <sensitive>

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
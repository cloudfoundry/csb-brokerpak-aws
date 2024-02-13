# Dead Letter Queue (DLQ) in Amazon SQS

- **Main Objective**: Capture unprocessed messages for later analysis.
  - **Error Handling and Debugging**
  - **Preventing Message Loss**
  - **Monitoring and Alerting**

- **Legacy Broker**: 
  - **DLQ**: No explicit mention of DLQ
  - **Sample Plan**: Empty redrive_policy
  - **Provisioning**: Among the properties accepted by the AWS API is RedrivePolicy, therefore, if it is possible to configure the queue as DLQ.
  - **Create Service Key**: Creates an IAM user and accepts a `policy_name`. This policy is associated with the user. So, we could configure such a poliza to have read access to the DLQ.


- **Is it possible to use a DLQ with multiple queues?**
  - **Yes. Multiple standard or FIFO queues can be configured to direct their failed messages to a single DLQ.**
  - **It adds a common point of failure**
  - **It centralizes monitoring and managing errors across these queues**
  - **It necessary consider configuration, permissions, error analysis, DLQ thresholds, scalability**:
    - **Permissions and Configuration**: 
        - Ensure DLQ has appropriate IAM policies for access from multiple queues.
        - Configure each standard queue to point to the same DLQ.
        - Adjust visibility timeouts and retry policies according to the workload.
        - Monitor and manage DLQ to prevent overaccumulation of messages.
    - **Policy Example**: [see example here](#sqs-dlq-policy-example)

- **Comparison between a DLQ used by muliple queues vs one queue**:
  - **Single Queue**: 
    - Easier to set up and maintain.
    - Delivered to the end user as a single service
    - Functionality to support multiple s-queues can be expanded later if required

  - **Multiple Queues**: 
    - If it is delivered to the end user as a single service:
      - It requires appropriate configuration to generate N queues at the same time as generating the DLQ.
    - If it is delivered as an **independent** service in CSB
      - Requires synchronization with the Standard Queue/s when creating the redrive policy.
      - Requires at least two bindings to connect to the Standard Queue and the DLQ. It depends on the number of s-queues.
    - This capability allows for a flexible and centralized approach to managing failed message processing across multiple sources.
    - Complexity at the first stage of service development



- **SQS DLQ Policy Example**:
  - **Principal**: "*". We can adjust the Principal section to include specific user IDs, roles, or AWS services.
  - **Condition**:
    - **SourceArn**: ARN of the source queues
  - **Objective**: This policy allows sourceQueue1 and sourceQueue2 to send messages to the DLQ.
  - **Example**:
    ```json
    {
    "Version": "2012-10-17",
    "Id": "PolicyID",
    "Statement": [
        {
            "Sid": "StmtID",
            "Effect": "Allow",
            "Principal": "*",
            "Action": "sqs:SendMessage",
            "Resource": "arn:aws:sqs:region:account-id:DLQName",
            "Condition": {
                "ArnEquals": {
                "aws:SourceArn": ["arn:aws:sqs:region:account-id:sourceQueue1", "arn:aws:sqs:region:account-id:sourceQueue2"]
                }
            }
        }
    ]
    }
    ```
  - **Redrive Policy**: : Each source queue that send messages to a DLQ must have a redrive policy. 
      - **ARN of the DLQ**
      - **Maximum number of times a message can be received before being moved to the DLQ (maxReceiveCount)**
        - **Example**:
            ```json
            {
              "deadLetterTargetArn": "arn:aws:sqs:region:account-id:DLQName",
              "maxReceiveCount": 5
            }
            ```
- **One Standard Queue To One DLQ Example**: [See code here](./dlq_one_to_one/README.md)
- **Two Standard Queues To One DLQ Example**: [See code here](./dlq_many_to_one/README.md)
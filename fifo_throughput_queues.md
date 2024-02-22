# FIFO (First-In-First-Out) queues 

Features: 

* Ensure the order of messages is preserved
* Messages are delivered exactly once

## FIFO throughput limit and deduplication scope

To optimize 
* Performance
* Behavior

### FIFO Throughput Limit

- **Purpose**: Controls the number of messages per second that can be sent or received from a FIFO queue. Modes:
    * Standard throughput
    * High throughput.

- **Standard vs. High Throughput**: By default, FIFO queues support up to 300 transactions per second (TPS) per action (send, receive, or delete).
With high throughput mode, this limit can be increased significantly to 3,000 TPS per action if batching is used, allowing for 10 actions per batch.

- **Configuration**: Enabling high throughput mode involves configuring the `fifo_throughput_limit` property.
The available settings typically include `perQueue` (the default mode with standard throughput limits) and `perMessageGroupId` (for high throughput mode).

### Deduplication Scope

- **Purpose**: Ensures that messages within the FIFO queue are **unique** based on the criteria defined by the deduplication scope.
It prevents the queue from processing the same message multiple times within the deduplication interval.

- **Scope**: Could be set to either the entire queue or at the message group level.
This is particularly important in FIFO queues where the order of message processing is critical.

- **Configuration**: The `deduplication_scope` property defines the scope. Values:
    * `messageGroup`: deduplication is performed within each message group
    * `queue`: deduplication across the entire queue:

### Relationship and Combination

- **Related but Independent**: While both settings are crucial for FIFO queues, they serve different purposes.

- **Combination Use**: In practice, these settings can be combined to optimize the performance and reliability of FIFO queues.
For instance, using high throughput mode (`fifo_throughput_limit` set to `perMessageGroupId`) in combination with message group-based
deduplication (`deduplication_scope` set to `messageGroup`) allows for high-volume, ordered processing with
deduplication checks to ensure messages within each group are unique.

### API limitations

* When setting FIFO throughput limit to perMessageGroupId, that is to say, High throughput Mode ON, the value for Deduplication Scope must be `messageGroup` or the operation fails.

```shell
Error: creating SQS Queue (csb-sqs-7468726f-7567-6870-7574-332e2e2e2e2e.fifo): operation error SQS: 
CreateQueue, https response error StatusCode: 400
InvalidAttributeValue: Invalid value for the parameter FifoThroughputLimit.
Reason: To set FifoThroughputLimit to perMessageGroupId, the DeduplicationScope must be messageGroup.
```

**Note**: the opposite example is allowed - Deduplication per Message Group with High Throughput Mode OFF


* When setting Deduplication Scope to messageGroup in a Standard Queue

Params: `'{"deduplication_scope":"messageGroup","fifo_throughput_limit":"perMessageGroupId"}'`

```shell
Error: creating SQS Queue (csb-sqs-7468726f-7567-6870-7574-352e2e2e2e2e): operation error SQS:
CreateQueue, https response error StatusCode: 400
InvalidAttributeName: You can specify the DeduplicationScope only when FifoQueue is set to true
```


* When setting FIFO throughput limit to perMessageGroupId in a Standard Queue

```shell
Error: creating SQS Queue (csb-sqs-7468726f-7567-6870-7574-362e2e2e2e2e): operation error SQS:
CreateQueue, https response error StatusCode: 400
InvalidAttributeName: You can specify the FifoThroughputLimit only when FifoQueue is set to true.
```
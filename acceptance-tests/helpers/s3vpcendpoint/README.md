# VPC Endpoint Management

This document provides instructions on how to create and delete a VPC endpoint with specific policies.

## Prerequisites

Ensure that you have Go installed on your machine and that you have set up your AWS credentials correctly.

## Environment Variables

The following environment variables are required to run the commands:

* AWS_ACCESS_KEY_ID: Your AWS access key ID
* AWS_SECRET_ACCESS_KEY: Your AWS secret access key

## Creating a VPC Endpoint

To create a VPC endpoint, you can use the `create` command followed by the VPC ID. The VPC ID is added to the endpoint
policy to allow connecting users from it whose username starts with `csb`. The policy also adds a statement to allow
connecting the main principal from within the VPC.

Run the following command in your terminal:

```shell
$ go run main.go create <VPC_ID> <USER_ARN>

```

Replace <VPC_ID> with your actual VPC ID and <USER_ARN> with the ARN of the main user principal. For example:

```shell
$ go run main.go create vpc-034198f0e9f2eacaa arn:aws:iam::XXXXXXX:user/automation-user
```

The output of this command will be the VPC endpoint ID.

## Deleting a VPC Endpoint

To delete a VPC endpoint, you can use the delete command followed by the VPC endpoint ID that was returned when you
created the endpoint. Run the following command in your terminal:

Run the following command in your terminal:

```shell
$ go run main.go delete <VPC_ENDPOINT_ID>
```

Replace <VPC_ENDPOINT_ID> with your actual VPC endpoint ID. For example:

```shell
$ go run main.go delete vpce-0176a37c22505c40f
```

This will delete the specified VPC endpoint.

> **Note:**
> Please replace `<VPC_ID>`, `<USER_ARN>` and `<VPC_ENDPOINT_ID>` with your actual VPC ID, user ARN, and VPC endpoint ID
> respectively.
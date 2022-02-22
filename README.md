# AWS Brokerpak

A brokerpak for the [Cloud Service Broker](https://github.com/pivotal/cloud-service-broker) that provides support for AWS services.

## Development Requirements

* [Docker](https://docs.docker.com/get-docker/) - tooling provided as docker images
* make - covers development lifecycle steps

A docker container for the cloud service broker binary is available at *cfplatformeng/csb*

## AWS account information

To provision services, the brokerpak currently requires AWS access key id and secret. The brokerpak expects them in environment variables:

* AWS_ACCESS_KEY_ID
* AWS_SECRET_ACCESS_KEY

## Development Tools

A Makefile supports the full local development lifecycle for the brokerpak.

Make targets will run with the *cfplatformeng/csb* docker image. Alternatively, a custom image can be specified by setting the `CSB` environment variable

Available make targets can be listed by running `make`.

There is a make target to push the broker and brokerpak into a CloudFoundry foundation. It will be necessary to manually configure a few items for the broker to work.

- `make push-broker` will `cf push` the broker into CloudFoundry. Requires the `cf` cli to be installed.

The broker gets pushed into CloudFoundry as *cloud-service-broker-aws*  It will be necessary to bind a MySQL database to the broker to provide broker state storage. See [AWS Installation](./docs/aws-installation.md) docs for more info.

## Broker
The version of Cloud Service Broker to use with this brokerpak is encoded in the `go.mod` file.
The make targets will use this version by default.
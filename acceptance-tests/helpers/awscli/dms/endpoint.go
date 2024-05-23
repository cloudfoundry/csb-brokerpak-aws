package dms

import (
	"csbbrokerpakaws/acceptance-tests/helpers/awscli"
	"csbbrokerpakaws/acceptance-tests/helpers/random"
	"fmt"
)

type EndpointType string

const (
	Source EndpointType = "source"
	Target EndpointType = "target"
)

type Endpoint struct {
	arn    string
	region string
}

type CreateEndpointParams struct {
	EndpointType    EndpointType
	EnvironmentName string
	Username        string
	Password        string
	Server          string
	DatabaseName    string
	Region          string
	Engine          string
	Port            int
}

func CreateEndpoint(params CreateEndpointParams) *Endpoint {
	id := random.Name(random.WithPrefix(params.EnvironmentName))

	var receiver struct {
		ARN string `jsonry:"Endpoint.EndpointArn"`
	}

	awscli.AWSToJSON(
		&receiver,
		"dms",
		"create-endpoint",
		"--endpoint-identifier", id,
		"--endpoint-type", string(params.EndpointType),
		"--engine-name", params.Engine,
		"--username", params.Username,
		fmt.Sprintf("--password=%s", params.Password), // workaround for passwords starting with dashes as suggested in https://github.com/aws/aws-cli/issues/1135#issuecomment-77254897
		"--port", fmt.Sprintf("%d", params.Port),
		"--server-name", params.Server,
		"--database-name", params.DatabaseName,
		"--region", params.Region,
	)

	return &Endpoint{
		arn:    receiver.ARN,
		region: params.Region,
	}
}

func (e *Endpoint) Cleanup() {
	awscli.AWS("dms", "delete-endpoint", "--region", e.region, "--endpoint-arn", e.arn)
}

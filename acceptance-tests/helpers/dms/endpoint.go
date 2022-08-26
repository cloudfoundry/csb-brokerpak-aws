package dms

import (
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

func CreateEndpoint(endpointType EndpointType, username, password, server, dbname, region string, port int) *Endpoint {
	id := random.Name()

	var receiver struct {
		ARN string `jsonry:"Endpoint.EndpointArn"`
	}

	AWSToJSON(
		&receiver,
		"dms",
		"create-endpoint",
		"--endpoint-identifier", id,
		"--endpoint-type", string(endpointType),
		"--engine-name", "postgres",
		"--username", username,
		"--password", password,
		"--port", fmt.Sprintf("%d", port),
		"--server-name", server,
		"--database-name", dbname,
		"--region", region,
	)

	return &Endpoint{
		arn:    receiver.ARN,
		region: region,
	}
}

func (e *Endpoint) Cleanup() {
	AWS("dms", "delete-endpoint", "--region", e.region, "--endpoint-arn", e.arn)
}

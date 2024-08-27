package dynaclient

import (
	"context"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	smithyendpoints "github.com/aws/smithy-go/endpoints"
)

// Fail fast if the interface is not implemented
var _ dynamodb.EndpointResolverV2 = endpointResolverV2{}

type endpointResolverV2 struct {
	endpoint string
}

func (e endpointResolverV2) ResolveEndpoint(ctx context.Context, params dynamodb.EndpointParameters) (smithyendpoints.Endpoint, error) {
	u, err := url.Parse(e.endpoint)
	if err != nil {
		return smithyendpoints.Endpoint{}, err
	}

	return smithyendpoints.Endpoint{URI: *u}, nil
}

// Package vpcendpoint provides test helpers for setting up an AWS VPC Endpoint
package vpcendpoint

import (
	"csbbrokerpakaws/acceptance-tests/helpers/awscli"
	"fmt"
)

type CallerIdentity struct {
	Arn string `json:"Arn"`
}

type RouteTable struct {
	RouteTableID string `json:"RouteTableId"`
}

type RouteTables struct {
	RouteTables []RouteTable `json:"RouteTables"`
}

type VpcEndpoint struct {
	VpcEndpointID string `json:"VpcEndpointId"`
}

type VpcEndpointResponse struct {
	VpcEndpoint VpcEndpoint `json:"VpcEndpoint"`
}

func CreateEndpoint(allowedVPCID, defaultRegion string) string {

	// Get the ARN of the current user
	getCallerIdentityCommand := []string{
		"sts",
		"get-caller-identity",
		"--output",
		"json",
	}

	var callerIdentity CallerIdentity
	awscli.AWSToJSON(&callerIdentity, getCallerIdentityCommand...)
	allowedUserARN := callerIdentity.Arn

	policyDocument := fmt.Sprintf(`{
		"Statement": [
			{
				"Action": "*",
				"Effect": "Allow",
				"Resource": "*",
				"Principal": "*",
				"Condition": {
					"StringEquals": {
						"aws:sourceVpc": %[1]q
					},
					"StringLike": {
						"aws:username": "csb-*"
					}
				}				
			},
			{
				"Action": "*",
				"Effect": "Allow",
				"Resource": "*",
				"Principal": {
					"AWS": %[2]q
				},
				"Condition": {
					"StringEquals": {
						"aws:sourceVpc":  %[1]q
					}
				}
			}
		]
	}`, allowedVPCID, allowedUserARN)

	describeRoutesTablesCommand := []string{
		"ec2",
		"describe-route-tables",
		"--filters",
		"Name=vpc-id,Values=" + allowedVPCID,
		"--output",
		"json",
	}

	var routesTables RouteTables
	awscli.AWSToJSON(&routesTables, describeRoutesTablesCommand...)

	routeTableIDs := make([]string, len(routesTables.RouteTables))
	for i, rt := range routesTables.RouteTables {
		routeTableIDs[i] = rt.RouteTableID
	}

	createEndpointCommand := []string{
		"ec2", "create-vpc-endpoint",
		"--vpc-id", allowedVPCID,
		"--service-name", fmt.Sprintf("com.amazonaws.%s.s3", defaultRegion),
		"--vpc-endpoint-type", "Gateway",
		"--route-table-ids",
	}

	createEndpointCommand = append(createEndpointCommand, routeTableIDs...)
	createEndpointCommand = append(createEndpointCommand, []string{"--policy-document", policyDocument, "--output", "json"}...)

	var response VpcEndpointResponse
	awscli.AWSToJSON(&response, createEndpointCommand...)
	return response.VpcEndpoint.VpcEndpointID
}

func DeleteVPCEndpoint(vpcEndpointID string) {
	deleteEndpointCommand := []string{
		"ec2", "delete-vpc-endpoints",
		"--vpc-endpoint-ids", vpcEndpointID,
	}

	awscli.AWS(deleteEndpointCommand...)
}

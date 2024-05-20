package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type createParams struct {
	allowedVPCID   string
	allowedUserARN string
}

func main() {
	createCommand := flag.NewFlagSet("create", flag.ExitOnError)
	deleteCommand := flag.NewFlagSet("delete", flag.ExitOnError)

	if len(os.Args) < 2 {
		fmt.Println("expected 'create' or 'delete' subcommands")
		os.Exit(1)
	}

	// Environment variables AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY are required
	svc := createAWSClientOrPanic()

	switch os.Args[1] {
	case "create":
		params := getCreateParamsFromArgsOrFail(createCommand)

		vpcEndpointID := createVPCEndpoint(svc, params)
		// vpcEndpointID to the stdout to be able to pipe it to the next step
		fmt.Println(vpcEndpointID)
	case "delete":
		vpcEndpointID := getVPCEndpointIDFromArgsOrFail(deleteCommand)
		deleteVPCEndpoint(svc, vpcEndpointID)
	default:
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func getVPCEndpointIDFromArgsOrFail(deleteCommand *flag.FlagSet) string {
	err := deleteCommand.Parse(os.Args[2:])
	if err != nil {
		fmt.Printf("error parsing 'delete' subcommand %s\n", err)
		os.Exit(1)
	}

	if len(deleteCommand.Args()) != 1 {
		fmt.Println("expected VPC Endpoint ID argument for 'delete' subcommand")
		os.Exit(1)
	}

	vpcEndpointID := deleteCommand.Args()[0]
	return vpcEndpointID
}

func getCreateParamsFromArgsOrFail(createCommand *flag.FlagSet) createParams {
	err := createCommand.Parse(os.Args[2:])
	if err != nil {
		fmt.Printf("error parsing 'create' subcommand %s\n", err)
		os.Exit(1)
	}

	if len(createCommand.Args()) != 2 {
		fmt.Println("expected allowed VPC ID argument for 'create' subcommand")
		os.Exit(1)
	}

	return createParams{
		allowedVPCID:   createCommand.Args()[0],
		allowedUserARN: createCommand.Args()[1],
	}
}

func createAWSClientOrPanic() *ec2.Client {
	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")

	if accessKey == "" || secretKey == "" {
		panic("Environment variables AWS_ACCESS_KEY_ID and/or AWS_SECRET_ACCESS_KEY are not set")
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		panic(err)
	}

	return ec2.NewFromConfig(cfg)
}

func createVPCEndpoint(svc *ec2.Client, params createParams) string {
	// Describe route tables
	rtInput := &ec2.DescribeRouteTablesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: []string{params.allowedVPCID},
			},
		},
	}
	rtOutput, err := svc.DescribeRouteTables(context.TODO(), rtInput)
	if err != nil {
		panic(err)
	}

	// Get all route table IDs
	routeTableIDs := make([]string, 0, len(rtOutput.RouteTables))
	for _, rt := range rtOutput.RouteTables {
		routeTableIDs = append(routeTableIDs, aws.ToString(rt.RouteTableId))
	}

	// Create VPC endpoint
	input := &ec2.CreateVpcEndpointInput{
		VpcId:           aws.String(params.allowedVPCID),
		ServiceName:     aws.String("com.amazonaws.us-west-2.s3"),
		VpcEndpointType: types.VpcEndpointTypeGateway,
		RouteTableIds:   routeTableIDs,
		PolicyDocument:  aws.String(createPolicyDocument(params)),
	}

	result, err := svc.CreateVpcEndpoint(context.TODO(), input)
	if err != nil {
		panic(err)
	}

	return *result.VpcEndpoint.VpcEndpointId
}

func createPolicyDocument(params createParams) string {
	return fmt.Sprintf(`{
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
							"aws:sourceVpc": %[1]q
						}
					}
				}
            ]
        }`, params.allowedVPCID, params.allowedUserARN)
}

func deleteVPCEndpoint(svc *ec2.Client, vpcEndpointID string) {
	input := &ec2.DeleteVpcEndpointsInput{
		VpcEndpointIds: []string{vpcEndpointID},
	}

	_, err := svc.DeleteVpcEndpoints(context.TODO(), input)
	if err != nil {
		panic(err)
	}

	fmt.Println("VPC Endpoint deleted: ", vpcEndpointID)
}

// arn:aws:iam::649758297924:user/automation-admin-user

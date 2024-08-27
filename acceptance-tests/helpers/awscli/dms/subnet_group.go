package dms

import (
	"csbbrokerpakaws/acceptance-tests/helpers/awscli"
	"csbbrokerpakaws/acceptance-tests/helpers/random"
	"fmt"
)

type replicationSubnetGroup struct {
	id     string
	region string
}

func createReplicationSubnetGroup(vpc, envName, region string) *replicationSubnetGroup {
	subnets := listSubnets(vpc, region)
	id := random.Name(random.WithPrefix(envName))
	awscli.AWS(append([]string{"dms", "create-replication-subnet-group", "--region", region, "--replication-subnet-group-identifier", id, "--replication-subnet-group-description", id, "--subnet-ids"}, subnets...)...)

	return &replicationSubnetGroup{
		id:     id,
		region: region,
	}
}

func (r *replicationSubnetGroup) cleanup() {
	awscli.AWS("dms", "delete-replication-subnet-group", "--replication-subnet-group-identifier", r.id, "--region", r.region)
}

func listSubnets(vpc, region string) []string {
	var receiver struct {
		Subnets []string `jsonry:"Subnets.SubnetId"`
	}
	awscli.AWSToJSON(&receiver, "ec2", "describe-subnets", "--filters", fmt.Sprintf("Name=vpc-id,Values=%s", vpc), "--region", region)
	return receiver.Subnets
}

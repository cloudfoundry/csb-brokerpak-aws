package dms

import (
	"csbbrokerpakaws/acceptance-tests/helpers/awscli"
	"csbbrokerpakaws/acceptance-tests/helpers/random"
	"fmt"
	"time"

	"github.com/onsi/ginkgo/v2"
)

const instancePollPeriod = 30 * time.Second

type ReplicationInstance struct {
	arn                    string
	region                 string
	replicationSubnetGroup *replicationSubnetGroup
}

func CreateReplicationInstance(vpc, envName, region string) *ReplicationInstance {
	subnetGroup := createReplicationSubnetGroup(vpc, envName, region)
	arn := createReplicationInstance(subnetGroup.id, envName, region)
	return &ReplicationInstance{
		arn:                    arn,
		region:                 region,
		replicationSubnetGroup: subnetGroup,
	}
}

func (r ReplicationInstance) Wait() {
	for start := time.Now(); time.Since(start) < time.Hour; {
		switch getReplicationInstanceState(r.arn, r.region) {
		case "creating": // wait
		case "available":
			return
		default:
			ginkgo.Fail("bad state")
		}
		time.Sleep(instancePollPeriod)
	}
	ginkgo.Fail("timed out")
}

func (r ReplicationInstance) Cleanup() {
	deleteReplicationInstance(r.arn, r.region)
	r.replicationSubnetGroup.cleanup()
}

func createReplicationInstance(replicationSubnetGroupID, envName, region string) string {
	var receiver struct {
		ARN string `jsonry:"ReplicationInstance.ReplicationInstanceArn"`
	}
	awscli.AWSToJSON(&receiver, "dms", "create-replication-instance", "--replication-instance-identifier", random.Name(random.WithPrefix(envName)), "--replication-instance-class", "dms.t3.micro", "--region", region, "--replication-subnet-group-identifier", replicationSubnetGroupID)

	return receiver.ARN
}

func getReplicationInstanceState(arn, region string) string {
	var receiver struct {
		StatusStrings []string `jsonry:"ReplicationInstances.ReplicationInstanceStatus"`
	}
	awscli.AWSToJSON(&receiver, "dms", "describe-replication-instances", "--region", region, "--filters", fmt.Sprintf("Name=replication-instance-arn,Values=%s", arn))
	switch len(receiver.StatusStrings) {
	default:
		ginkgo.Fail("matched more than one instance")
		fallthrough // unreachable
	case 0:
		return ""
	case 1:
		return receiver.StatusStrings[0]
	}
}

func deleteReplicationInstance(arn, region string) {
	awscli.AWS("dms", "delete-replication-instance", "--replication-instance-arn", arn, "--region", region)

	for start := time.Now(); time.Since(start) < time.Hour; {
		if !replicationInstanceExists(arn, region) {
			return
		}
		time.Sleep(instancePollPeriod)
	}

	ginkgo.Fail("timed out")
}

func replicationInstanceExists(arn, region string) bool {
	var receiver struct {
		ARNs []string `jsonry:"ReplicationInstances.ReplicationInstanceArn"`
	}
	awscli.AWSToJSON(&receiver, "dms", "describe-replication-instances", "--region", region)

	for _, a := range receiver.ARNs {
		if a == arn {
			return true
		}
	}
	return false
}

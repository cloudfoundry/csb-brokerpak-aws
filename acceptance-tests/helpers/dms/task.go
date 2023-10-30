package dms

import (
	"csbbrokerpakaws/acceptance-tests/helpers/random"
	"fmt"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

const taskPollPeriod = 30 * time.Second

func RunReplicationTask(replicationInstance *ReplicationInstance, source, target *Endpoint, region, schema string) {
	taskID := random.Name()

	var taskReceiver struct {
		ARN string `jsonry:"ReplicationTask.ReplicationTaskArn"`
	}

	AWSToJSON(
		&taskReceiver,
		"dms",
		"create-replication-task",
		"--replication-task-identifier", taskID,
		"--source-endpoint-arn", source.arn,
		"--target-endpoint-arn", target.arn,
		"--replication-instance-arn", replicationInstance.arn,
		"--migration-type", "full-load",
		"--table-mappings", fmt.Sprintf(`{"rules":[{"rule-type":"selection","rule-id":"1","rule-name":"1","object-locator":{"schema-name":"%s","table-name":"%%"},"rule-action":"include","filters":[]}]}`, schema),
		"--region", region,
	)

	defer replicationTaskDeletion(taskReceiver.ARN, region)

	waitForReplicationTaskCreation(taskID, region)

	AWS(
		"dms",
		"start-replication-task",
		"--replication-task-arn", taskReceiver.ARN,
		"--region", region,
		"--start-replication-task-type", "start-replication",
	)

	waitForReplicationTaskCompletion(taskID, region)
}

func waitForReplicationTaskCompletion(taskID, region string) {
	for start := time.Now(); time.Since(start) < 10*time.Minute; {
		var statusReceiver struct {
			Status      []string `jsonry:"ReplicationTasks.Status"`
			Percentages []int    `jsonry:"ReplicationTasks.ReplicationTaskStats.FullLoadProgressPercent"`
		}

		AWSToJSON(
			&statusReceiver,
			"dms",
			"describe-replication-tasks",
			"--region", region,
			"--filters", fmt.Sprintf("Name=replication-task-id,Values=%s", taskID),
		)

		gomega.Expect(statusReceiver.Status).To(gomega.HaveLen(1))

		switch statusReceiver.Status[0] {
		case "starting", "running":
		case "stopped":
			gomega.Expect(statusReceiver.Percentages).To(gomega.HaveLen(1))
			gomega.Expect(statusReceiver.Percentages[0]).To(gomega.Equal(100))
			return
		default:
			ginkgo.Fail(fmt.Sprintf("unexpected status: %q", statusReceiver.Status[0]))
		}

		time.Sleep(taskPollPeriod)
	}

	ginkgo.Fail("timed out")
}

func waitForReplicationTaskCreation(taskID, region string) {
	for start := time.Now(); time.Since(start) < 10*time.Minute; {
		var statusReceiver struct {
			Status []string `jsonry:"ReplicationTasks.Status"`
		}

		AWSToJSON(
			&statusReceiver,
			"dms",
			"describe-replication-tasks",
			"--region", region,
			"--filters", fmt.Sprintf("Name=replication-task-id,Values=%s", taskID),
		)

		gomega.Expect(statusReceiver.Status).To(gomega.HaveLen(1))

		switch statusReceiver.Status[0] {
		case "creating":
			time.Sleep(taskPollPeriod)
		case "ready":
			return
		default:
			ginkgo.Fail("unexpected status")
		}
	}

	ginkgo.Fail("timed out")
}

func replicationTaskDeletion(taskARN, region string) {
	AWS(
		"dms",
		"delete-replication-task",
		"--region", region,
		"--replication-task-arn", taskARN,
	)

	for start := time.Now(); time.Since(start) < time.Hour; {
		if !replicationTaskExists(taskARN, region) {
			return
		}
		time.Sleep(taskPollPeriod)
	}

	ginkgo.Fail("timed out")
}

func replicationTaskExists(arn, region string) bool {
	var receiver struct {
		ARNs []string `jsonry:"ReplicationTasks.ReplicationTaskArn"`
	}
	AWSToJSON(&receiver, "dms", "describe-replication-tasks", "--region", region)

	for _, a := range receiver.ARNs {
		if a == arn {
			return true
		}
	}
	return false
}

package util

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestParserMqClusterMessage(t *testing.T) {
	Convey("Given some rabbitmq cluster message", t, func() {
		rabbitMessage := `{"timestamp": 1445239796, "clusterId": 4, "type": "slaveState", "attached": "attached string", "nodeId":"0d7ba959af6940c99a369697cf912b2b", "message": "message string"}`
		Convey("Parse the rabbitmq cluster message", func() {
			var x = ParserMqClusterMessage([]byte(rabbitMessage))

			Convey("Test the rabbitmq cluster message should equal {rabMessage}", func() {
				rabMessage := RabbitMqMessage{
					Timestamp: 1445239796,
					ClusterId: 4,
					Type:      "slaveState",
					Attached:  "attached string",
					NodeId:    "0d7ba959af6940c99a369697cf912b2b",
					Message:   "message string"}
				So(*x, ShouldResemble, rabMessage)
			})
		})
	})
}

func TestMasterMetricsJson(t *testing.T) {
	Convey("Given some RabbitMqMessage.Message", t, func() {
		rabbitMessage := RabbitMqMessage{
			Timestamp: 1445239796,
			ClusterId: 4,
			Type:      "slaveState",
			Attached:  "attached string",
			NodeId:    "0d7ba959af6940c99a369697cf912b2b",
			Message:   "{\"master\\/cpus_percent\":0,\"master\\/cpus_total\":2,\"master\\/cpus_used\":0,\"master\\/disk_total\":0,\"master\\/disk_used\":0,\"master\\/elected\":1,\"master\\/mem_total\":1024,\"master\\/mem_used\":0}"}
		Convey("Parse the RabbitMqMessage.Message", func() {
			var x = MasterMetricsJson(rabbitMessage)
			Convey("Test the RabbitMqMessage.Message should equal {rabMessage}", func() {
				rabMessage := MasterMetricsMar{
					CpuPercent: 0.0,
					CpuShare:   0.0,
					CpuTotal:   2.0,
					MemTotal:   1024,
					MemUsed:    0,
					DiskTotal:  0,
					DiskUsed:   0,
					Timestamp:  1445239796,
					Leader:     1,
					ClusterId:  "4",
				}
				So(x, ShouldResemble, rabMessage)
			})
		})
	})
}

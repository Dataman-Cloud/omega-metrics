package util

//import (
//	. "github.com/smartystreets/goconvey/convey"
//	"testing"
//)
//
//func TestParserMqClusterMessage(t *testing.T) {
//	Convey("Given some rabbitmq cluster message", t, func() {
//		rabbitMessage := `{"timestamp": 1445239796, "clusterId": 4, "type": "slaveState", "attached": "attached string", "nodeId":"0d7ba959af6940c99a369697cf912b2b", "message": "message string"}`
//
//		Convey("Parse the rabbitmq cluster message", func() {
//			var x = ParserMqClusterMessage([]byte(rabbitMessage))
//
//			Convey("Test the rabbitmq cluster message should equal {rabMessage}", func() {
//				rabMessage := RabbitMqMessage{
//					Timestamp: 1445239796,
//					ClusterId: 4,
//					Type:      "slaveState",
//					Attached:  "attached string",
//					NodeId:    "0d7ba959af6940c99a369697cf912b2b",
//					Message:   "message string"}
//				So(*x, ShouldEqual, rabMessage)
//			})
//		})
//	})
//}

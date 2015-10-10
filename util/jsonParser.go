package util

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"strconv"
)

const (
	Deployment_success      = "deployment_success"
	Deployment_failed       = "deployment_failed"
	Deployment_info         = "deployment_info"
	Deployment_step_success = "deployment_step_success"
	Deployment_step_failure = "deployment_step_failure"
)

type Message struct {
	Code string      `json:"code"`
	Data interface{} `json:"data"`
	Err  string      `json:"error"`
}

type RabbitMqMessage struct {
	ClusterId int    `json:"clusterId"`
	NodeId    string `json:"nodeId"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

type MasterMetrics struct {
	CpuPercent float64 `json:"master/cpus_percent"`
	DiskUsed   int     `json:"master/disk_used"`
	DiskTotal  int     `json:"master/disk_total"`
	MemUsed    int     `json:"master/mem_used"`
	MemTotal   int     `json:"master/mem_total"`
	Leader     int     `json:"master/elected"`
}

type MasterMetricsMar struct {
	CpuPercent float64 `json:"cpuPercent"`
	MemTotal   int     `json:"memTotal"`
	MemUsed    int     `json:"memUsed"`
	DiskTotal  int     `json:"diskTotal"`
	DiskUsed   int     `json:"diskUsed"`
	Timestamp  int64   `json:"timestamp"`
	Leader     int     `json:"leader"`
}

type SlaveMetrics struct {
	NodeId     string  `json:"nodeId"`
	CpuPercent float64 `json:"slave/cpus_total"`
	Disk_used  int     `json:"slave/disk_used"`
	Disk_total int     `json:"slave/disk_total"`
	Mem_used   int     `json:"slave/mem_used"`
	Mem_total  int     `json:"slave/mem_total"`
}

type MarathonEvent struct {
	EventType   string      `json:"eventType"`
	Timestamp   string      `json:"timestamp"`
	Id          string      `json:"id,omitempty"`
	Plan        plan        `json:"plan,omitempty"`
	CurrentStep currentStep `json:"currentStep,omitempty"`
}

type currentStep struct {
	Actions []actions
}

type actions struct {
	Type string
	App  string
}

type plan struct {
	Id string
}

func Handler(routingKey string, messageBody []byte) {
	switch routingKey {
	case Master_metrics_routing:
		nodeId, leader, json := MasterMetricsJson(string(messageBody))
		log.Infof("received message nodeId:%s leader:%d json:%s", nodeId, leader, json)
	case Slave_metrics_routing:
		nodeId, json := SlaveMetricsJson(string(messageBody))
		log.Infof("received message nodeId:%s json:%s", nodeId, json)
	}
}

func ReturnMessage(code string, strs []string, errMessage string) interface{} {
	var s Message
	var ss MasterMetricsMar
	var ls []interface{}
	s.Code = code
	s.Err = errMessage
	for _, str := range strs {
		json.Unmarshal([]byte(str), &ss)
		ls = append(ls, ss)
	}
	s.Data = ls
	_, err := json.Marshal(s)
	if err != nil {
		log.Error("[ReturnMessage] failed: ", err)
		return ""
	}
	return s
}

func MasterMetricsJson(str string) (string, int, string) {
	var mmm RabbitMqMessage
	var mm MasterMetrics
	var ss MasterMetricsMar
	json.Unmarshal([]byte(str), &mmm)
	clusterId := strconv.Itoa(mmm.ClusterId)
	json.Unmarshal([]byte(mmm.Message), &mm)

	ss.CpuPercent = mm.CpuPercent
	ss.MemTotal = mm.MemTotal
	ss.MemUsed = mm.MemUsed
	ss.DiskUsed = mm.DiskUsed
	ss.DiskTotal = mm.DiskTotal
	ss.Leader = mm.Leader
	ss.Timestamp = mmm.Timestamp

	ll, err := json.Marshal(ss)
	if err != nil {
		log.Error("Master Metrics parse failed", err)
		return "", 0, ""
	}
	return clusterId, mm.Leader, string(ll)
}

func SlaveMetricsJson(str string) (string, string) {
	var s SlaveMetrics
	json.Unmarshal([]byte(str), &s)
	nodeId := s.NodeId
	ll, err := json.Marshal(s)
	if err != nil {
		log.Error("Slave Metrics parse failed", err)
		return "", ""
	}
	return nodeId, string(ll)
}

func MarathonEventJson(str string) (string, string, string, string) {
	var rmm RabbitMqMessage
	var me MarathonEvent
	json.Unmarshal([]byte(str), &rmm)
	clusterId := strconv.Itoa(rmm.ClusterId)
	json.Unmarshal([]byte(rmm.Message), &me)
	switch me.EventType {
	case Deployment_info:
		fmt.Println("&&&&&&&&&&&& deployment info: ", rmm.Message)
		return me.EventType, clusterId, me.Plan.Id, me.CurrentStep.Actions[0].App
	case Deployment_success:
		fmt.Println("&&&&&&&&&&&& deployment success: ", rmm.Message)
		return me.EventType, clusterId, me.Id, me.Timestamp
	case Deployment_failed:
		fmt.Println("&&&&&&&&&&&& deployment failed: ", rmm.Message)
		return me.EventType, clusterId, me.Id, me.Timestamp
	case Deployment_step_success:
		fmt.Println("&&&&&&&&&&&& deployment step success: ", rmm.Message)
		return me.EventType, clusterId, me.CurrentStep.Actions[0].App, me.Timestamp
	case Deployment_step_failure:
		fmt.Println("&&&&&&&&&&&& deployment step failure: ", rmm.Message)
		return me.EventType, clusterId, me.CurrentStep.Actions[0].App, me.Timestamp
	}
	return "", "", "", ""
}

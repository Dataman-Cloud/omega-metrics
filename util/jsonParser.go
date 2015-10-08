package util

import (
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"strconv"
)

type Message struct {
	Code string      `json:"code"`
	Data interface{} `json:"data"`
	Err  string      `json:"error"`
}

type MasterMetricsMessage struct {
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
	var mmm MasterMetricsMessage
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

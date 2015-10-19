package util

import (
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
	"strings"

	log "github.com/cihub/seelog"
)

const (
	Deployment_success      = "deployment_success"
	Deployment_failed       = "deployment_failed"
	Deployment_info         = "deployment_info"
	Deployment_step_success = "deployment_step_success"
	Deployment_step_failure = "deployment_step_failure"
	Status_update_event     = "status_update_event"
	Destroy_app             = "destroy_app"
)

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

type SlaveMetrics struct {
	NodeId     string  `json:"nodeId"`
	CpuPercent float64 `json:"slave/cpus_total"`
	Disk_used  int     `json:"slave/disk_used"`
	Disk_total int     `json:"slave/disk_total"`
	Mem_used   int     `json:"slave/mem_used"`
	Mem_total  int     `json:"slave/mem_total"`
}

type StatusUpdate struct {
	EventType  string `json:"eventType"`
	Timestamp  string `json:"timestamp"`
	AppId      string `json:"appId"`
	Host       string `json:"host"`
	Ports      []int  `json:"ports"`
	TaskStatus string `json:"taskStatus"`
}

type DestroyApp struct {
	EventType string `json:"eventType"`
	Timestamp string `json:"timestamp"`
	AppId     string `json:"appId"`
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

var parserTypeMappings map[string]reflect.Type

func init() {
	recognizedTypes := []interface{}{
		MarathonEventMar{},
		MasterMetricsMar{},
	}

	parserTypeMappings = make(map[string]reflect.Type)
	for _, recognizedType := range recognizedTypes {
		parserTypeMappings[reflect.TypeOf(recognizedType).Name()] = reflect.TypeOf(recognizedType)
	}
}

func NewOfType(typ string) (interface{}, bool) {
	rtype, ok := parserTypeMappings[typ]
	if !ok {
		return nil, false
	}

	return reflect.New(rtype).Interface(), true
}

func ReturnMessage(typ string, strs []string) (*[]interface{}, error) {
	var monitorDatas []interface{}
	for _, str := range strs {
		monitorType, ok := NewOfType(typ)
		if !ok {
			return nil, errors.New(typ + " is not support type")
		}
		json.Unmarshal([]byte(str), &monitorType)
		monitorDatas = append(monitorDatas, monitorType)
	}
	return &monitorDatas, nil
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

func MarathonEventMarshal(eventType, timestamp, idOrApp, currentType, taskId string) string {
	var mem MarathonEventMar
	mem.EventType = eventType
	mem.Timestamp = timestamp
	mem.IdOrApp = idOrApp
	mem.CurrentType = currentType
	mem.TaskId = taskId

	ll, err := json.Marshal(mem)
	if err != nil {
		log.Error("[MarathonEventMar] struct marshal failed", err)
		return ""
	}
	return string(ll)
}

func MarathonEventJson(str string) (string, string, string, string, string, string) {
	var rmm RabbitMqMessage
	var me MarathonEvent

	var su StatusUpdate
	json.Unmarshal([]byte(str), &rmm)
	clusterId := strconv.Itoa(rmm.ClusterId)
	json.Unmarshal([]byte(rmm.Message), &me)
	log.Debugf("marathon event type: [%s] message %s", me.EventType, rmm.Message)
	switch me.EventType {
	case Deployment_info:
		return me.EventType, clusterId, me.Plan.Id, me.Timestamp, me.CurrentStep.Actions[0].App, ""
	case Deployment_success:
		return me.EventType, clusterId, me.Id, me.Timestamp, "", ""
	case Deployment_failed:
		return me.EventType, clusterId, me.Id, me.Timestamp, "", ""
	case Deployment_step_success:
		return me.EventType, clusterId, me.CurrentStep.Actions[0].App, me.Timestamp, me.CurrentStep.Actions[0].Type, ""
	case Deployment_step_failure:
		return me.EventType, clusterId, me.CurrentStep.Actions[0].App, me.Timestamp, me.CurrentStep.Actions[0].Type, ""
	case Status_update_event:
		json.Unmarshal([]byte(rmm.Message), &su)
		var portArray []string
		for _, v := range su.Ports {
			j := strconv.Itoa(v)
			portArray = append(portArray, j)
		}
		portstr := strings.Join(portArray, ",")
		appId := su.Host + ":" + portstr
		return me.EventType, clusterId, su.AppId, su.Timestamp, su.TaskStatus, appId
	case Destroy_app:
		var da DestroyApp
		json.Unmarshal([]byte(rmm.Message), &da)
		return me.EventType, clusterId, da.AppId, da.Timestamp, da.EventType, ""
	}
	return "", clusterId, "", "", "", ""
}

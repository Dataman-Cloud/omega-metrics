package util

import (
	"encoding/json"
	"errors"
	"reflect"
	"strconv"

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

var parserTypeMappings map[string]reflect.Type

func init() {
	recognizedTypes := []interface{}{
		MasterMetricsMar{},
		SlaveStateMar{},
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
		err := json.Unmarshal([]byte(str), &monitorType)
		if err != nil {
			log.Error("[ReturnMessage] unmarshal monitorType error ", err)
			return nil, err
		}
		monitorDatas = append(monitorDatas, monitorType)
	}
	return &monitorDatas, nil
}

func ParserMqClusterMessage(messgae []byte) *RabbitMqMessage {
	var mqMessage *RabbitMqMessage = &RabbitMqMessage{}
	err := json.Unmarshal(messgae, mqMessage)
	if err != nil {
		log.Error("Parser mq message has error: ", err)
		return nil
	}
	return mqMessage
}

func ReturnData(typ, str string) (*interface{}, error) {
	monitorType, ok := NewOfType(typ)
	if !ok {
		return nil, errors.New(typ + " is not support type")
	}
	err := json.Unmarshal([]byte(str), &monitorType)
	if err != nil {
		log.Error("[ReturnData] unmarshal monitorType error ", err)
		return nil, err
	}
	return &monitorType, nil
}

func MasterMetricsJson(rabbitMessage RabbitMqMessage) MasterMetricsMar {
	var masMet MasterMetrics
	var masMetMar MasterMetricsMar
	clusterId := strconv.Itoa(rabbitMessage.ClusterId)
	err := json.Unmarshal([]byte(rabbitMessage.Message), &masMet)
	if err != nil {
		log.Error("[MasterMetrics] unmarshal MasterMetrics error ", err)
		return masMetMar
	}

	return MasterMetricsMar{
		CpuPercent: masMet.CpuPercent * 100,
		CpuShare:   masMet.CpuShare,
		CpuTotal:   masMet.CpuTotal,
		MemTotal:   masMet.MemTotal,
		MemUsed:    masMet.MemUsed,
		DiskUsed:   masMet.DiskUsed,
		DiskTotal:  masMet.DiskTotal,
		Leader:     masMet.Leader,
		Timestamp:  rabbitMessage.Timestamp,
		ClusterId:  clusterId,
	}
}

func MasterStateJson(rabbitMessage RabbitMqMessage) MasterStateMar {
	var masSta MasterState
	var masStaMar MasterStateMar
	clusterId := strconv.Itoa(rabbitMessage.ClusterId)
	err := json.Unmarshal([]byte(rabbitMessage.Message), &masSta)
	if err != nil {
		log.Error("[MasterState] unmarshal MasterState error ", err)
		return masStaMar
	}
	masStaMar.Timestamp = rabbitMessage.Timestamp
	masStaMar.ClusterId = clusterId

	if len(masSta.Frameworks) == 0 {
		masStaMar.Leader = 0
		return masStaMar
	}
	for _, v := range masSta.Frameworks {
		if v.Name == "marathon" {
			for _, task := range v.Tasks {
				var apps AppAndTasks
				apps.TaskId = task.Id
				apps.AppName = task.Name
				masStaMar.AppAndTasks = append(masStaMar.AppAndTasks, apps)
			}
		}
	}
	masStaMar.Leader = 1
	return masStaMar
}

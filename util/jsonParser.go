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

var parserTypeMappings map[string]reflect.Type

func init() {
	recognizedTypes := []interface{}{
		MarathonEventMar{},
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

	ss.CpuPercent = mm.CpuPercent * 100
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

func parseMesosPorts(str string) (string, error) {
	if str == "" {
		return "", nil
	}
	str1 := strings.Replace(str, "[", "", -1)
	str2 := strings.Replace(str1, "]", "", -1)
	arr := strings.Split(str2, "-")
	start, err := strconv.Atoi(arr[0])
	if err != nil {
		return "", errors.New("string to int error: " + arr[0])
	}
	end, err := strconv.Atoi(arr[1])
	if err != nil {
		return "", errors.New("string to int error: " + arr[1])
	}
	var portsArr []string
	for i := start; i <= end; i++ {
		portsArr = append(portsArr, strconv.Itoa(i))
	}
	return strings.Join(portsArr, ","), nil

}

func SlaveStateJson(str string) []SlaveStateMar {
	var js RabbitMqMessage
	var message SlaveState
	var s map[string]ContainerInfo
	var array []SlaveStateMar

	json.Unmarshal([]byte(str), &js)
	clusterId := strconv.Itoa(js.ClusterId)
	// parse "message"
	json.Unmarshal([]byte(js.Message), &message)
	ip := message.Flags.Ip
	m := make(map[string]appNameAndId)
	for _, v := range message.Frameworks {
		if v.Name == "marathon" {
			for _, exec := range v.Executors {
				slaveId := strings.Split(exec.Directory, "/")[4]
				key := "mesos-" + slaveId + "." + exec.Container
				var value appNameAndId
				lastindex := strings.LastIndex(exec.Id, ".")
				value.AppName = exec.Id[:lastindex]
				portstring, err := parseMesosPorts(exec.Resources.Ports)
				if err != nil {
					log.Error("parseMessosPorts error: ", err)
				}
				value.AppId = ip + ":" + portstring
				m[key] = value

			}
		}
	}

	// parse "attached"
	json.Unmarshal([]byte(js.Attached), &s)
	for _, value := range s {
		if len(value.Stats) != 2 {
			log.Error("[slave state] length of Stats isn't 2, can't calc cpurate")
			continue
		}
		var conInfo SlaveStateMar

		flag := false
		var app appNameAndId
		var containerId string
		for _, aliase := range value.Aliases {
			_, ok := m[aliase]
			if ok {
				flag = true
				containerId = aliase
				app = m[aliase]
			}
		}
		if flag == false {
			continue
		}
		conInfo.ClusterId = clusterId
		conInfo.App = app
		conInfo.ContainerId = containerId
		//      conInfo.Timestamp = value.Stats[1].Timestamp
		conInfo.CpuUsed = int64(value.Stats[1].Cpu.Usage.Total - value.Stats[0].Cpu.Usage.Total)
		conInfo.CpuTotal = (value.Stats[1].Timestamp.Sub(value.Stats[0].Timestamp).Nanoseconds())
		conInfo.MemoryUsed = value.Stats[1].Memory.Usage
		conInfo.MemoryTotal = value.Spec.Memory.Limit
		ls, _ := json.Marshal(conInfo)
		log.Debugf("AppMetrics: ", string(ls))
		array = append(array, conInfo)
	}
	return array
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

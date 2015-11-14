package util

import (
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
	"strings"
	"time"

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
		err := json.Unmarshal([]byte(str), &monitorType)
		if err != nil {
			log.Error("[ReturnMessage] unmarshal monitorType error ", err)
			return nil, err
		}
		monitorDatas = append(monitorDatas, monitorType)
	}
	return &monitorDatas, nil
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

func MasterMetricsJson(str string) MasterMetricsMar {
	var rabbitMessage RabbitMqMessage
	var masMet MasterMetrics
	var masMetMar MasterMetricsMar
	err := json.Unmarshal([]byte(str), &rabbitMessage)
	if err != nil {
		log.Error("[MasterMetrics] unmarshal RabbitMqMessage error ", err)
		return masMetMar
	}
	clusterId := strconv.Itoa(rabbitMessage.ClusterId)
	err = json.Unmarshal([]byte(rabbitMessage.Message), &masMet)
	if err != nil {
		log.Error("[MasterMetrics] unmarshal MasterMetrics error ", err)
		return masMetMar
	}
	masMetMar.CpuPercent = masMet.CpuPercent * 100
	masMetMar.CpuShare = masMet.CpuShare
	masMetMar.CpuTotal = masMet.CpuTotal
	masMetMar.MemTotal = masMet.MemTotal
	masMetMar.MemUsed = masMet.MemUsed
	masMetMar.DiskUsed = masMet.DiskUsed
	masMetMar.DiskTotal = masMet.DiskTotal
	masMetMar.Leader = masMet.Leader
	masMetMar.Timestamp = rabbitMessage.Timestamp
	masMetMar.ClusterId = clusterId
	return masMetMar
}

func MasterStateJson(str string) MasterStateMar {
	var rabbitMessage RabbitMqMessage
	var masSta MasterState
	var masStaMar MasterStateMar
	err := json.Unmarshal([]byte(str), &rabbitMessage)
	if err != nil {
		log.Error("[MasterState] unmarshal RabbitMqMessage error ", err)
		return masStaMar
	}
	clusterId := strconv.Itoa(rabbitMessage.ClusterId)
	err = json.Unmarshal([]byte(rabbitMessage.Message), &masSta)
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
	var rabbitMessage RabbitMqMessage
	var message SlaveState
	var cadInfo map[string]ContainerInfo
	var array []SlaveStateMar

	err := json.Unmarshal([]byte(str), &rabbitMessage)
	if err != nil {
		log.Error("[SlaveState] unmarshal RabbitMqMessage error ", err)
		return array
	}
	clusterId := strconv.Itoa(rabbitMessage.ClusterId)
	// parse "message"
	err = json.Unmarshal([]byte(rabbitMessage.Message), &message)
	if err != nil {
		log.Error("[SlaveState] unmarshal SlaveState error ", err)
		return array
	}
	ip := message.Flags.Ip
	m := make(map[string]appInfo)
	for _, v := range message.Frameworks {
		if v.Name == "marathon" {
			var num int = 0
			for _, exec := range v.Executors {
				slaveId := exec.Tasks[0].Slave_id
				key := "mesos-" + slaveId + "." + exec.Container
				var value appInfo
				value.Task_id = exec.Id
				value.Slave_id = slaveId
				value.AppName = exec.Tasks[0].Name
				value.Resources = exec.Tasks[0].Resources
				portstring, err := parseMesosPorts(exec.Tasks[0].Resources.Ports)
				if err != nil {
					log.Error("parseMessosPorts error: ", err)
				}
				if portstring == "" {
					value.AppId = ip + "-" + strconv.Itoa(num)
					num += 1
				} else {
					value.AppId = ip + ":" + portstring
				}
				m[key] = value

			}
		}
	}

	// parse "attached"
	err = json.Unmarshal([]byte(rabbitMessage.Attached), &cadInfo)
	if err != nil {
		log.Error("[SlaveState] unmarshal cadvisor containerInfo error ", err)
		return array
	}
	for _, value := range cadInfo {
		if len(value.Stats) != 2 {
			log.Error("[slave state] length of Stats isn't 2, can't calc cpurate")
			continue
		}
		var conInfo SlaveStateMar

		flag := false
		var app appInfo
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
		cpuUsed := float64(value.Stats[1].Cpu.Usage.Total - value.Stats[0].Cpu.Usage.Total)
		cpuTotal := float64(value.Stats[1].Timestamp.Sub(value.Stats[0].Timestamp).Nanoseconds())
		cpuCores := float64(len(value.Stats[1].Cpu.Usage.PerCpu))
		conInfo.CpuUsedCores = cpuUsed / cpuTotal * cpuCores

		conInfo.CpuShareCores = app.Resources.Cpus
		conInfo.MemoryUsed = value.Stats[1].Memory.Usage / (1024 * 1024)
		conInfo.MemoryTotal = app.Resources.Mem
		ls, _ := json.Marshal(conInfo)
		log.Debugf("AppMetrics: ", string(ls))
		array = append(array, conInfo)
	}
	return array
}

func marathonEventMarshal(timestamp string) string {
	// 改变时间戳格式"2015-10-21T07:16:31.802Z" 为 "2015-10-21 07:16:31"
	layout := "2006-01-02T15:04:05.999Z"
	t, err := time.Parse(layout, timestamp)
	if err != nil {
		log.Error("[marathon event] timestamp parse error", err)
		return timestamp
	}
	nt := t.Add(time.Hour * 8)
	return nt.Format("2006-01-02 15:04:05")
}

func MarathonEventJson(str string) MarathonEventMar {
	var rabbitMessage RabbitMqMessage
	var marEvent MarathonEvent
	var marEventMar MarathonEventMar

	var statusUpdate StatusUpdate
	err := json.Unmarshal([]byte(str), &rabbitMessage)
	if err != nil {
		log.Error("[MarathonEvent] unmarshal RabbitMqMessage error ", err)
		return marEventMar
	}
	marEventMar.ClusterId = strconv.Itoa(rabbitMessage.ClusterId)
	err = json.Unmarshal([]byte(rabbitMessage.Message), &marEvent)
	if err != nil {
		log.Error("[MarathonEvent] unmarshal MarathonEvent error ", err)
		return marEventMar
	}
	log.Debugf("marathon event type: [%s] message %s", marEvent.EventType, rabbitMessage.Message)
	switch marEvent.EventType {
	case Deployment_info:
		marEventMar.EventType = marEvent.EventType
		marEventMar.App.AppId = marEvent.Plan.Id
		marEventMar.App.AppName = strings.Replace(marEvent.CurrentStep.Actions[0].App, "/", "", 1)
		marEventMar.Timestamp = marathonEventMarshal(marEvent.Timestamp)
		return marEventMar
	case Deployment_success:
		marEventMar.EventType = marEvent.EventType
		marEventMar.App.AppId = marEvent.Id
		marEventMar.Timestamp = marathonEventMarshal(marEvent.Timestamp)
		return marEventMar
	case Deployment_failed:
		marEventMar.EventType = marEvent.EventType
		marEventMar.App.AppId = marEvent.Id
		marEventMar.Timestamp = marathonEventMarshal(marEvent.Timestamp)
		return marEventMar
	case Deployment_step_success:
		marEventMar.EventType = marEvent.EventType
		marEventMar.App.AppName = strings.Replace(marEvent.CurrentStep.Actions[0].App, "/", "", 1)
		marEventMar.Timestamp = marathonEventMarshal(marEvent.Timestamp)
		marEventMar.CurrentType = marEvent.CurrentStep.Actions[0].Type
		return marEventMar
	case Deployment_step_failure:
		marEventMar.EventType = marEvent.EventType
		marEventMar.App.AppName = strings.Replace(marEvent.CurrentStep.Actions[0].App, "/", "", 1)
		marEventMar.Timestamp = marathonEventMarshal(marEvent.Timestamp)
		marEventMar.CurrentType = marEvent.CurrentStep.Actions[0].Type
		return marEventMar
	case Status_update_event:
		err = json.Unmarshal([]byte(rabbitMessage.Message), &statusUpdate)
		if err != nil {
			log.Error("[MarathonEvent] unmarshal StatusUpdate error ", err)
			return marEventMar
		}
		var portArray []string
		for _, v := range statusUpdate.Ports {
			j := strconv.Itoa(v)
			portArray = append(portArray, j)
		}
		portstr := strings.Join(portArray, ",")
		appId := statusUpdate.Host + ":" + portstr
		marEventMar.EventType = marEvent.EventType
		marEventMar.App.AppName = strings.Replace(statusUpdate.AppId, "/", "", 1)
		marEventMar.Timestamp = marathonEventMarshal(statusUpdate.Timestamp)
		marEventMar.CurrentType = statusUpdate.TaskStatus
		marEventMar.TaskId = appId
		return marEventMar
	case Destroy_app:
		var destroyApp DestroyApp
		err = json.Unmarshal([]byte(rabbitMessage.Message), &destroyApp)
		if err != nil {
			log.Error("[MarathonEvent] unmarshal DestroyApp error ", err)
			return marEventMar
		}
		marEventMar.EventType = marEvent.EventType
		marEventMar.App.AppName = strings.Replace(destroyApp.AppId, "/", "", 1)
		marEventMar.Timestamp = marathonEventMarshal(destroyApp.Timestamp)
		marEventMar.CurrentType = destroyApp.EventType
		return marEventMar
	}
	return marEventMar
}

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

func SlaveStateJson(rabbitMessage RabbitMqMessage) []SlaveStateMar {

	var message SlaveState
	var cadInfo map[string]ContainerInfo
	var array []SlaveStateMar

	clusterId := strconv.Itoa(rabbitMessage.ClusterId)
	// parse "message"
	err := json.Unmarshal([]byte(rabbitMessage.Message), &message)
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
				if len(exec.Tasks) == 0 {
					log.Debug("[SlaveState] Executors.Tasks length is 0, Message is : ", message)
					continue
				}
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
		if len(value.Stats) < 2 {
			log.Error("[SlaveState] length of Stats isn't larger than 2, can't calc cpurate")
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
		deltatime := value.Stats[1].Timestamp.Sub(value.Stats[0].Timestamp)
		cpuUsed := float64(value.Stats[1].Cpu.Usage.Total - value.Stats[0].Cpu.Usage.Total)
		cpuTotal := float64(deltatime.Nanoseconds())
		//cpuCores := float64(len(value.Stats[1].Cpu.Usage.PerCpu))
		conInfo.CpuUsedCores = cpuUsed / cpuTotal

		conInfo.CpuShareCores = float64(app.Resources.Cpus)
		conInfo.MemoryUsed = value.Stats[1].Memory.Usage / (1024 * 1024)
		conInfo.MemoryTotal = app.Resources.Mem

		if value.Spec.HasNetwork {
			conInfo.NetworkReceviedByteRate = (float64(value.Stats[1].Network.RxBytes) -
				float64(value.Stats[0].Network.RxBytes)) / deltatime.Seconds()
			conInfo.NetworkSentByteRate = (float64(value.Stats[1].Network.TxBytes) -
				float64(value.Stats[0].Network.TxBytes)) / deltatime.Seconds()
		}

		if value.Spec.HasDiskIo && (len(value.Stats[1].DiskIo.IoServiceBytes) > 0) {
			DiskIOReadBytes, ok := value.Stats[1].DiskIo.IoServiceBytes[0].Stats["Read"]
			if ok {
				log.Infof("[SlaveState] Get the disk io read bytes")
				conInfo.DiskIOReadBytes = float64(DiskIOReadBytes)
			} else {
				log.Error("[SlaveState] Failed to get the disk io read bytes")
			}
			DiskIOWriteBytes, ok := value.Stats[1].DiskIo.IoServiceBytes[0].Stats["Write"]
			if ok {
				log.Infof("[SlaveState] Get the disk io write bytes")
				conInfo.DiskIOWriteBytes = float64(DiskIOWriteBytes)
			} else {
				log.Error("[SlaveState] Failed to get the disk io write bytes")
			}
		}

		ls, _ := json.Marshal(conInfo)
		log.Debugf("AppMetrics: ", string(ls))
		array = append(array, conInfo)
	}
	return array
}

package slave

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/Dataman-Cloud/omega-metrics/cache"
	"github.com/Dataman-Cloud/omega-metrics/config"
	"github.com/Dataman-Cloud/omega-metrics/db"
	"github.com/Dataman-Cloud/omega-metrics/util"
	log "github.com/cihub/seelog"
)

// hanlder slave state message
// slave state contain mesos-slave info, container monitor and haproxy sessions status
func SlaveStateHandler(message *util.RabbitMqMessage) {
	if message == nil {
		return
	}

	clusterId := strconv.Itoa(message.ClusterId)
	// parse "message"-> mesos-slave info
	var slaveState util.SlaveState
	if err := json.Unmarshal([]byte(message.Message), &slaveState); err != nil {
		log.Error("[SlaveState] unmarshal SlaveState error: ", err)
		return
	}

	if slaveState.Id == "" {
		log.Error("Salve id is null")
		return
	}

	appInfo, err := ParseSlaveState(clusterId, slaveState)
	if err != nil {
		return
	}

	if err := ParseAppMonitorData(&message.Attached, appInfo); err != nil {
		log.Error("[Slave state] Parse app monitor info got error:  ", err.Error())
	}

	if sessionInfo, ok := message.Tags["session"]; ok {
		sessionKey := clusterId + ":" + slaveState.Id
		if err := ParseSessionData(sessionInfo, sessionKey); err != nil {
			log.Error("Slave state] Parse session data got error: ", err.Error())
		}
	}

}

// construction app container name by executor and slave id
// get the resrouce the mesos distribution to every container
func ParseSlaveState(clusterId string, slaveState util.SlaveState) (map[string]util.AppInfo, error) {
	slaveAppMap := make(map[string]util.AppInfo)
	slaveId := slaveState.Id
	slaveIp := slaveState.Flags.Ip

	for _, v := range slaveState.Frameworks {
		if v.Name == "marathon" {
			for index, exec := range v.Executors {
				if len(exec.Tasks) == 0 {
					log.Warnf("[SlaveState] Executors.Tasks length is 0, Message is : %+v", exec)
					continue
				}
				key := "mesos-" + slaveId + "." + exec.Container
				var value util.AppInfo
				value.TaskId = exec.Id
				value.SlaveId = slaveId
				value.AppName = exec.Tasks[0].Name
				value.Resources = exec.Tasks[0].Resources
				portstring, err := parseMesosPorts(exec.Tasks[0].Resources.Ports)
				if err != nil {
					log.Error("parseMessosPorts error: ", err)
				}
				var appId string
				if portstring == "" {
					appId = slaveIp + "-" + strconv.Itoa(index)
				} else {
					appId = slaveIp + ":" + portstring
				}
				appInfo := util.AppInfo{
					TaskId:    exec.Id,
					ClusterId: clusterId,
					SlaveId:   slaveId,
					AppName:   exec.Tasks[0].Name,
					Resources: exec.Tasks[0].Resources,
					AppId:     appId,
				}
				slaveAppMap[key] = appInfo
			}
		}
	}

	return slaveAppMap, nil
}

// parse app monitor data to get the resourse seizure and mate container and app
func ParseAppMonitorData(message *string, slaveInfo map[string]util.AppInfo) error {
	if message == nil {
		return errors.New("[Slave attach] messgae is null")
	}

	var cadInfo map[string]util.ContainerInfo
	if err := json.Unmarshal([]byte(*message), &cadInfo); err != nil {
		log.Error("[Slave attach] unmarshal cadvisor containerInfo error ", err)
		return err
	}

	for _, value := range cadInfo {
		if len(value.Stats) < 2 {
			log.Error("[SlaveState] length of Stats isn't larger than 2, can't calc cpurate")
			continue
		}

		var conInfo util.SlaveStateMar

		flag := false
		var app util.AppInfo
		var containerId string
		for _, aliase := range value.Aliases {
			_, ok := slaveInfo[aliase]
			if ok {
				flag = true
				containerId = aliase
				app = slaveInfo[aliase]
			}
		}

		// if contianer name not mactch in mesos slave info continue
		if flag == false {
			continue
		}

		conInfo.App = app
		conInfo.ContainerId = containerId
		conInfo.Timestamp = value.Stats[1].Timestamp

		deltatime := value.Stats[1].Timestamp.Sub(value.Stats[0].Timestamp)

		// calculate cpu  use present
		cpuUsed := float64(value.Stats[1].Cpu.Usage.Total - value.Stats[0].Cpu.Usage.Total)
		cpuTotal := float64(deltatime.Nanoseconds())
		conInfo.CpuUsedCores = cpuUsed / cpuTotal

		// calculate memory use info
		conInfo.CpuShareCores = float64(app.Resources.Cpus)
		conInfo.MemoryUsed = value.Stats[1].Memory.Usage / (1024 * 1024)
		conInfo.MemoryTotal = app.Resources.Mem

		// calculate disk write and read rate B/s
		if value.Spec.HasNetwork && (len(value.Stats[1].Network.Interfaces) > 0) {
			var receivedBytes uint64
			var sentBytes uint64
			for _, networkStats := range value.Stats[1].Network.Interfaces {
				receivedBytes += uint64(networkStats.RxBytes)
				sentBytes += uint64(networkStats.TxBytes)
			}
			for _, networkStats := range value.Stats[0].Network.Interfaces {
				receivedBytes -= uint64(networkStats.RxBytes)
				sentBytes -= uint64(networkStats.TxBytes)
			}
			conInfo.NetworkReceviedByteRate = float64(receivedBytes) / deltatime.Seconds()
			conInfo.NetworkSentByteRate = float64(sentBytes) / deltatime.Seconds()
		}

		// calculate net send and receive rate B/s
		if value.Spec.HasDiskIo && (len(value.Stats[1].DiskIo.IoServiceBytes) > 0) {
			var readBytes uint64
			var writeBytes uint64
			for _, diskStats := range value.Stats[1].DiskIo.IoServiceBytes {
				readBytes += diskStats.Stats["Read"]
				writeBytes += diskStats.Stats["Write"]
			}
			for _, diskStats := range value.Stats[0].DiskIo.IoServiceBytes {
				readBytes -= diskStats.Stats["Read"]
				writeBytes -= diskStats.Stats["Write"]
			}
			conInfo.DiskIOReadBytesRate = float64(readBytes) / deltatime.Seconds()
			conInfo.DiskIOWriteBytesRate = float64(writeBytes) / deltatime.Seconds()
		}

		go WriteContainerInfoToCache(&conInfo)

		go db.WriteContainerInfoToInflux(&conInfo)
	}
	return nil
}

// write container info to cache for cluster monitor
func WriteContainerInfoToCache(conInfo *util.SlaveStateMar) {
	infoBytes, err := json.Marshal(conInfo)
	if err != nil {
		log.Error("[Slave state] marshal container info to cache got error: ", err)
		return
	}

	if err := cache.WriteStringToRedis(conInfo.App.SlaveId, string(infoBytes), config.ContainerMonitorTimeout); err != nil {
		log.Error("[Slave state] write container info to cache got error: ", err)
		return
	}

	return
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

// parse app sessions info
func ParseSessionData(message *string, sessionKey string) error {
	var sessionList []util.HaproxySession

	if err := json.Unmarshal([]byte(*message), &sessionList); err != nil {
		return err
	}

	var appReqInfos []util.AppRequestInfo
	for _, session := range sessionList {
		proxyName := strings.Trim(session.ProxyName, ":")
		splits := strings.Split(proxyName, "-")
		if len(splits) < 1 {
			continue
		}

		appReqInfo := util.AppRequestInfo{
			AppName: splits[0],
			ReqRate: session.ReqRate,
		}

		appReqInfos = append(appReqInfos, appReqInfo)
	}

	go WriteAppReqInfoToCache(sessionKey, &appReqInfos)
	return nil
}

// write app req info to cache
func WriteAppReqInfoToCache(key string, appReqInfo *[]util.AppRequestInfo) {
	infoBytes, err := json.Marshal(appReqInfo)
	if err != nil {
		log.Error("[Slave state] marshal appReqInfo got error: ", err)
		return
	}

	if err := cache.WriteStringToRedis(key, string(infoBytes), config.SessionInfoTimeout); err != nil {
		log.Error("[Slave state] Write appReqInfo to cache got error: ", err.Error())
		return
	}

	return
}

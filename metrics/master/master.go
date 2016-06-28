package master

import (
	"encoding/json"
	"strconv"

	"github.com/Dataman-Cloud/omega-metrics/cache"
	"github.com/Dataman-Cloud/omega-metrics/config"
	"github.com/Dataman-Cloud/omega-metrics/db"
	"github.com/Dataman-Cloud/omega-metrics/metrics"
	"github.com/Dataman-Cloud/omega-metrics/util"
	log "github.com/Sirupsen/logrus"
)

// handler master metric message
func MasterMetricHandler(messageBody *[]byte) {
	mqMessage := metrics.ParserMqMessage(messageBody)
	if mqMessage == nil {
		log.Error("[Master metric] Unmarshal master metrics got nil")
		return
	}

	masterMetric := ParseMasterMetrics(mqMessage)
	if masterMetric.ClusterId != "" && masterMetric.Leader == 1 {
		key := masterMetric.ClusterId + "_" + util.Master_metrics_routing
		value, err := json.Marshal(masterMetric)
		if err != nil {
			log.Error("[Master metric] Marshal master metric got error: ", err)
			return
		}

		if err := cache.WriteStringToRedis(key, string(value), config.DefaultTimeout); err != nil {
			log.Error("[Master metric] Write master metric to cache got error: ", err)
			return
		}
	}

	return
}

// parse master metric info and write to cache
func ParseMasterMetrics(rabbitMessage *util.RabbitMqMessage) util.MasterMetricsMar {
	var masMet util.MasterMetrics
	var masMetMar util.MasterMetricsMar
	clusterId := strconv.Itoa(rabbitMessage.ClusterId)
	err := json.Unmarshal([]byte(rabbitMessage.Message), &masMet)
	if err != nil {
		log.Error("[Master metric] unmarshal MasterMetrics error ", err)
		return masMetMar
	}

	return util.MasterMetricsMar{
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

// master state handler function
func MasterStateHandler(messageBody *[]byte) {
	mqMessage := metrics.ParserMqMessage(messageBody)
	if mqMessage == nil {
		return
	}

	masterState := ParseMasterState(mqMessage)
	clusterId := masterState.ClusterId
	if clusterId == "" || masterState.Leader != 1 {
		return
	}

	go WriteAppTaskInfoToCache(clusterId, masterState.AppAndTasks)
	CalculateAppReqRate(clusterId, masterState.Slaves)
}

// parse master state message
func ParseMasterState(rabbitMessage *util.RabbitMqMessage) util.MasterStateMar {
	var masSta util.MasterState
	var masStaMar util.MasterStateMar
	clusterId := strconv.Itoa(rabbitMessage.ClusterId)
	if err := json.Unmarshal([]byte(rabbitMessage.Message), &masSta); err != nil {
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
				var apps util.AppAndTasks
				apps.TaskId = task.Id
				apps.AppName = task.Name
				masStaMar.AppAndTasks = append(masStaMar.AppAndTasks, apps)
			}
		}
	}
	masStaMar.Leader = 1
	masStaMar.Slaves = masSta.Slaves
	return masStaMar
}

// write app task info to cache for cluster monitor
func WriteAppTaskInfoToCache(clusterId string, appTasks []util.AppAndTasks) {
	for _, task := range appTasks {
		label := clusterId + ":" + task.AppName
		err := cache.WriteSetToRedis(label, task.TaskId, config.ContainerMonitorTimeout)
		if err != nil {
			log.Error("[Master_state] writeSetToRedis has err: ", err)
		}
	}
}

// calculate app req rate in all slaves witn all app
func CalculateAppReqRate(clusterId string, slaves []util.MasterSlaveInfo) {
	appReqMap := make(map[string]*util.InfluxAppRequestInfo)

	// sum all slaves req rate
	for _, slave := range slaves {
		sessionKey := clusterId + ":" + slave.Id
		slaveReqInfo, err := cache.ReadFromRedis(sessionKey)
		if err != nil {
			log.Errorf("[Mastet state] Get %s from cache got error: %s ", sessionKey, err.Error())
			continue
		}

		var appSlaveReqs []util.AppRequestInfo
		if err := json.Unmarshal([]byte(slaveReqInfo), &appSlaveReqs); err != nil {
			log.Error("[Master state] Unmarshal salve reqs got errot: ", err)
			continue
		}

		// sum all app req rate in one of slave
		for _, appSlaveReq := range appSlaveReqs {
			appName := appSlaveReq.AppName
			appReq, ok := appReqMap[appName]
			if !ok {
				appReq = &util.InfluxAppRequestInfo{
					AppName:   appName,
					ReqRate:   appSlaveReq.ReqRate,
					ClusterId: clusterId,
				}
				appReqMap[appName] = appReq
			} else {
				appReq.ReqRate += appSlaveReq.ReqRate
			}
		}
	}

	go WriteAppReqInfoToDb(appReqMap)
}

func WriteAppReqInfoToDb(appReqMap map[string]*util.InfluxAppRequestInfo) {
	for _, appReqInfo := range appReqMap {
		db.WriteAppReqInfoToInflux(appReqInfo)
	}
}

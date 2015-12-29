package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Dataman-Cloud/omega-metrics/cache"
	"github.com/Dataman-Cloud/omega-metrics/config"
	"github.com/Dataman-Cloud/omega-metrics/util"
	log "github.com/cihub/seelog"
	redis "github.com/garyburd/redigo/redis"
	"github.com/gin-gonic/gin"
)

func startC() {
	log.Debug("start master metrics mq consumer")
	go util.MetricsSubscribe(util.Metrics_exchange, util.Master_metrics_routing, handler)
	go util.MetricsSubscribe(util.Metrics_exchange, util.Slave_state_routing, handler)
	go util.MetricsSubscribe(util.Metrics_exchange, util.Master_state_routing, handler)
	go util.MetricsSubscribe(util.Metrics_exchange, util.Slave_metrics_routing, func(routingKey string, messageBody []byte) {})
	go util.MetricsSubscribe(util.Metrics_exchange, util.Marathon_info_routing, func(routingKey string, messageBody []byte) {})
	go util.MetricsSubscribe(util.Metrics_exchange, util.Marathon_metrics_routing, func(routingKey string, messageBody []byte) {})
	go util.MetricsSubscribe(util.Metrics_exchange, util.Slave_monitor_routing, func(routingKey string, messageBody []byte) {})

}

//function use to handle cross-domain requests
func SetHeader(ctx *gin.Context) {
	ctx.Header("Access-Control-Allow-Origin", "*")
	ctx.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	ctx.Header("Access-Control-Allow-Headers", "Content-Type, Depth, User-Agent, X-File-Size, X-Requested-With, X-Requested-By, If-Modified-Since, X-File-Name, Cache-Control, X-XSRFToken, Authorization")
	if ctx.Request.Method == "OPTIONS" {
		ctx.String(204, "")
	}
	ctx.Next()
}

func handler(routingKey string, messageBody []byte) {
	mqMessage := util.ParserMqClusterMessage(messageBody)
	if mqMessage == nil {
		return
	}

	switch routingKey {
	case util.Master_metrics_routing:
		jsonstr := util.MasterMetricsJson(*mqMessage)
		if jsonstr.ClusterId != "" && jsonstr.Leader == 1 {
			label := jsonstr.ClusterId + "_" + routingKey
			value, _ := json.Marshal(jsonstr)
			err := cache.WriteStringToRedis(label, string(value), config.DefaultTimeout)
			if err != nil {
				log.Error("[Master_metrics] writeStringToRedis has err: ", err)
			}
		}
		log.Infof("received masterMetricsRouting message clusterId:%s leader:%d json:%+v", jsonstr.ClusterId, jsonstr.Leader, jsonstr)
	case util.Master_state_routing:
		jsonstr := util.MasterStateJson(*mqMessage)
		if jsonstr.ClusterId != "" && jsonstr.Leader == 1 {
			for _, task := range jsonstr.AppAndTasks {
				label := jsonstr.ClusterId + ":" + task.AppName
				err := cache.WriteSetToRedis(label, task.TaskId, config.ContainerMonitorTimeout)
				if err != nil {
					log.Error("[Master_state] writeSetToRedis has err: ", err)
				}
			}
		}
		log.Infof("received masterStateRouting message clusterId: %s, leader: %d, json: %+v", jsonstr.ClusterId, jsonstr.Leader, jsonstr)
	case util.Slave_state_routing:
		array := util.SlaveStateJson(*mqMessage)
		if len(array) != 0 {
			for _, v := range array {
				key := v.App.Task_id
				value, _ := json.Marshal(v)
				err := cache.WriteStringToRedis(key, string(value), config.ContainerMonitorTimeout)
				if err != nil {
					log.Error("[Slave_state] writeHashToRedis has err: ", err)
				}
			}
		}
		log.Infof("received slaveStateMessage array: %s", array)
	}
}

func masterMetrics(ctx *gin.Context) {
	conf := config.Pairs()
	conn := cache.Open()
	defer conn.Close()

	var httpstr util.AppListResponse
	var cm util.ClusterMetrics
	cluster_id := ctx.Param("cluster_id") + "_" + util.Master_metrics_routing
	log.Debug("cluster_id ", cluster_id)

	response := util.MonitorResponse{
		Code: 1,
		Data: nil,
		Err:  "",
	}

	rs, err := cache.ReadFromRedis(cluster_id)
	if err != nil {
		log.Errorf("[Master Metrics] read key %v FromRedis has err: %v", cluster_id, err)
		response.Err = "[Master Metrics] read from redis error " + err.Error()
		ctx.JSON(http.StatusOK, response)
		return
	}
	masMet, err := util.ReturnData(util.MonitorMasterMetrics, rs)
	if err != nil {
		log.Error("[Master Metrics] analysis error ", err)
		response.Err = "[Master Metrics] analysis error " + err.Error()
		ctx.JSON(http.StatusOK, response)
		return
	}

	token := util.Header(ctx, HeaderToken)
	client := &http.Client{}
	addr := fmt.Sprintf("%s:%d/api/v1/applications/0/appStatus", conf.Omega_app_host, conf.Omega_app_port)
	req, err := http.NewRequest("GET", addr, nil)
	if err != nil {
		log.Error("[Master Metrics] creat new http request error: ", err)
		response.Err = "[Master Metrics] creat new http request error: " + err.Error()
		ctx.JSON(http.StatusOK, response)
		return
	}
	req.Header.Add("Authorization", token)
	resp, err := client.Do(req)
	if err != nil {
		log.Error("http request error: ", err)
		response.Err = "[Master Metrics] http request error: " + err.Error()
		ctx.JSON(http.StatusOK, response)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("[Master Metrics] read response body error: ", err)
		response.Err = "[Master Metrics] read response body error: " + err.Error()
		ctx.JSON(http.StatusOK, response)
		return
	}
	err = json.Unmarshal([]byte(string(body)), &httpstr)
	if err != nil {
		log.Error("[Master Metrics] parse http response body error ", err)
		response.Err = "[Master Metrics] parse http response body error: " + err.Error()
		ctx.JSON(http.StatusOK, response)
		return
	}
	for _, v := range httpstr.Data {
		// 判断app所属集群
		if v.Cid != ctx.Param("cluster_id") {
			continue
		}
		appm, err := gatherApp(v)
		if err != nil {
			response.Err = err.Error()
			ctx.JSON(http.StatusOK, response)
			return
		}
		cm.AppMetrics = append(cm.AppMetrics, appm)
	}

	cm.MasMetrics = *masMet
	response.Code = 0
	response.Data = cm
	ctx.JSON(http.StatusOK, response)
}

func gatherApp(app util.StatusAndTask) (util.AppMetric, error) {
	conn := cache.Open()
	defer conn.Close()

	var result util.AppMetric
	key := app.Cid + ":" + app.Name
	smems, err := redis.Strings(conn.Do("SMEMBERS", key))
	if err != nil {
		log.Error("[gatherApp] redis error ", err)
		return result, err
	}
	var strs []string
	for _, smem := range smems {
		str, err := cache.ReadFromRedis(smem)
		if err != nil {
			log.Errorf("[App Metrics] Read key %v FromRedis error %v", smem, err)
			continue
		}
		if err == nil && str != "" {
			strs = append(strs, str)
		}
	}
	var cpuUsedSum float64
	var cpuShareSum float64
	var memUsedSum uint64
	var memTotalSum uint64
	for _, str := range strs {
		var task util.SlaveStateMar
		err := json.Unmarshal([]byte(str), &task)
		if err != nil {
			log.Error("[gatherApp] parse SlaveStateMar error ", err)
			return result, err
		}
		cpuUsedSum += task.CpuUsedCores
		cpuShareSum += task.CpuShareCores
		memUsedSum += task.MemoryUsed
		memTotalSum += task.MemoryTotal
	}
	result.AppName = app.Name
	result.AppCpuUsed = cpuUsedSum
	result.AppCpuShare = cpuShareSum
	result.AppMemUsed = memUsedSum
	result.AppMemShare = memTotalSum
	result.Status = app.Status
	return result, nil
}

func appMetrics(ctx *gin.Context) {
	conn := cache.Open()
	defer conn.Close()

	response := util.MonitorResponse{
		Code: 1,
		Data: nil,
		Err:  "",
	}

	key := ctx.Param("cluster_id") + ":" + ctx.Param("app")
	smems, err := redis.Strings(conn.Do("SMEMBERS", key))
	if err != nil {
		log.Error("[App Metrics] SMEMBERS error ", err)
		response.Err = "[App Metrics] SMEMBERS error " + err.Error()
		ctx.JSON(http.StatusOK, response)
		return
	}
	var strs []string
	for _, smem := range smems {
		str, err := cache.ReadFromRedis(smem)
		if err != nil {
			log.Errorf("[App Metrics] Read key %v FromRedis error %v ", smem, err)
			continue
		}
		if err == nil && str != "" {
			strs = append(strs, str)
		}
	}

	jsoninterface, err := util.ReturnMessage(util.MonitorAppMetrics, strs)
	if err != nil {
		log.Error("[App Metrics] analysis error ", err)
		response.Err = "[App Metrics] analysis error " + err.Error()
		ctx.JSON(http.StatusOK, response)
		return
	}
	response.Code = 0
	response.Data = *jsoninterface
	ctx.JSON(http.StatusOK, response)
}

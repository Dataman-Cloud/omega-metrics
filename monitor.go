package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/Dataman-Cloud/omega-metrics/cache"
	"github.com/Dataman-Cloud/omega-metrics/metrics/master"
	"github.com/Dataman-Cloud/omega-metrics/metrics/slave"
	"github.com/Dataman-Cloud/omega-metrics/util"
	log "github.com/cihub/seelog"
	redis "github.com/garyburd/redigo/redis"
	"github.com/gin-gonic/gin"
)

func startC() {
	log.Debug("start master metrics mq consumer")
	go util.MetricsSubscribe(util.Metrics_exchange, util.Master_state_routing, master.MasterStateHandler)
	go util.MetricsSubscribe(util.Metrics_exchange, util.Master_metrics_routing, master.MasterMetricHandler)
	go util.MetricsSubscribe(util.Metrics_exchange, util.Slave_state_routing, slave.SlaveStateHandler)
	go util.MetricsSubscribe(util.Metrics_exchange, util.Slave_metrics_routing, func(messageBody *[]byte) {})
	go util.MetricsSubscribe(util.Metrics_exchange, util.Marathon_info_routing, func(messageBody *[]byte) {})
	go util.MetricsSubscribe(util.Metrics_exchange, util.Marathon_metrics_routing, func(messageBody *[]byte) {})
	go util.MetricsSubscribe(util.Metrics_exchange, util.Slave_monitor_routing, func(messageBody *[]byte) {})

}

//function use to handle cross-domain requests
func OptionHandler(ctx *gin.Context) {
	ctx.Header("Access-Control-Allow-Origin", "*")
	ctx.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	ctx.Header("Access-Control-Allow-Headers", "Content-Type, Depth, User-Agent, X-File-Size, X-Requested-With, X-Requested-By, If-Modified-Since, X-File-Name, Cache-Control, X-XSRFToken, Authorization")
	if ctx.Request.Method == "OPTIONS" {
		ctx.String(204, "")
	}
	ctx.Next()
}

func gatherApp(app util.StatusAndTask) (util.AppMetric, error) {
	conn := cache.Open()
	defer conn.Close()

	var result util.AppMetric
	key := strconv.FormatInt(app.Cid, 10) + ":" + app.Alias
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
		memUsedSum += uint64(task.MemoryUsed)
		memTotalSum += uint64(task.MemoryTotal)
	}
	result.AppName = app.Name
	result.AppCpuUsed = cpuUsedSum
	result.AppCpuShare = cpuShareSum
	result.AppMemUsed = memUsedSum
	result.AppMemShare = memTotalSum
	result.Status = app.Status
	result.Instances = app.Instances
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

func Milliseconds(d time.Duration) float64 {
	min := d / 1e6
	nsec := d % 1e6
	return float64(min) + float64(nsec)*(1e-6)
}

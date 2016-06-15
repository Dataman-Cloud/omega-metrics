package controller

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/Dataman-Cloud/omega-metrics/cache"
	"github.com/Dataman-Cloud/omega-metrics/db"
	"github.com/Dataman-Cloud/omega-metrics/util"
	log "github.com/cihub/seelog"
	"github.com/gin-gonic/gin"
)

func AppMonitorHandler(c *gin.Context) {
	clusterId := c.Param("cluster_id")
	if clusterId == "" {
		ReturnError(c, InvalidParams, errors.New("cluster id is empty"))
		return
	}

	appName := c.Param("app")
	if appName == "" {
		ReturnError(c, InvalidParams, errors.New("appName is empty"))
		return
	}

	var startTime, endTime int64
	var err error
	startTimeStr := c.Query("starttime")
	if startTime, err = strconv.ParseInt(startTimeStr, 10, 64); err != nil {
		startTime = time.Now().Add(-1 * time.Hour).UnixNano()
	}

	endTimeStr := c.Query("endtime")
	if endTime, err = strconv.ParseInt(endTimeStr, 10, 64); err != nil {
		endTime = time.Now().UnixNano()
	}

	var results []map[string]interface{}
	results, err = db.QueryMetricsInfo(clusterId, appName, startTime, endTime)
	if err != nil {
		ReturnError(c, DbQueryError, err)
		return
	}

	ReturnOk(c, results)
	return
}

// cluster metrucs handler
func ClusterMetricsHandler(c *gin.Context) {
	clusterId := c.Param("cluster_id")
	if clusterId == "" {
		ReturnError(c, InvalidParams, errors.New("cluster id is empty"))
		return
	}

	masterMetrics, err := GetMasterMetricFromCache(clusterId)
	if err != nil {
		log.Error("[Master metric] Get master metric from cache got error: ", err)
		ReturnError(c, DbQueryError, err)
		return
	}

	token := GetToken(c)
	if token == "" {
		log.Error("[Master metric] Get token failed token is empty")
		ReturnError(c, InvalidParams, errors.New("token is empty"))
		return
	}

	appList, err := util.QueryApps(token, clusterId)
	if err != nil {
		log.Error("[Master metrics] Get app list from omega-app failed", err)
		ReturnError(c, DbQueryError, err)
		return
	}

	appStatus, err := util.QueryAppStatus(token)
	if err != nil {
		log.Error("[Master metrics] Get app status from omega-app failed ", err)
		ReturnError(c, DbQueryError, err)
		return
	}

	var clusterMetrics util.ClusterMetrics
	// check cluster app is created bu user
	for _, appConfig := range appList {
		appId := strconv.FormatInt(appConfig.Id, 10)
		appSt, ok := appStatus[appId]
		if !ok {
			continue
		}

		appMetrics, err := gatherApp(appSt)
		if err != nil {
			log.Errorf("[Cluster metrics] gatherApp %f metrics got error: %s", appId, err.Error())
			continue
		}

		clusterMetrics.AppMetrics = append(clusterMetrics.AppMetrics, appMetrics)
	}

	clusterMetrics.MasMetrics = masterMetrics
	ReturnOk(c, clusterMetrics)
	return
}

// get master metric form cache and parse data
func GetMasterMetricFromCache(clusterId string) (util.MasterMetricsMar, error) {
	var masterMetrics util.MasterMetricsMar
	key := clusterId + "_" + util.Master_metrics_routing
	value, err := cache.ReadFromRedis(key)
	if err != nil {
		return masterMetrics, err
	}

	if err := json.Unmarshal([]byte(value), &masterMetrics); err != nil {
		return masterMetrics, err
	}

	return masterMetrics, nil
}

func gatherApp(app util.AppStatus) (util.AppMetric, error) {
	var result util.AppMetric
	key := strconv.FormatInt(app.Cid, 10) + ":" + app.Alias
	smems, err := cache.ReadSetMembers(key)
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

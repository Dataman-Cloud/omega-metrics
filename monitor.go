package main

import (
	"encoding/json"
	"net/http"

	"github.com/Dataman-Cloud/omega-metrics/cache"
	"github.com/Dataman-Cloud/omega-metrics/util"
	log "github.com/cihub/seelog"
	redis "github.com/garyburd/redigo/redis"
	"github.com/gin-gonic/gin"
)

func startC() {
	log.Debug("start master metrics mq consumer")
	util.MetricsSubscribe(util.Metrics_exchange, util.Master_metrics_routing, handler)
	util.MetricsSubscribe(util.Metrics_exchange, util.Marathon_event_routing, handler)
	util.MetricsSubscribe(util.Metrics_exchange, util.Slave_state_routing, handler)
	util.MetricsSubscribe(util.Metrics_exchange, util.Slave_metrics_routing, func(routingKey string, messageBody []byte) {})
	util.MetricsSubscribe(util.Metrics_exchange, util.Marathon_apps_routing, func(routingKey string, messageBody []byte) {})
	util.MetricsSubscribe(util.Metrics_exchange, util.Marathon_metrics_routing, func(routingKey string, messageBody []byte) {})
	util.MetricsSubscribe(util.Metrics_exchange, util.Marathon_deployments_routing, func(routingKey string, messageBody []byte) {})

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
	var jsonstr, taskId, timestamp, clusterId, idOrApp, eventType, currentType string
	var leader int
	switch routingKey {
	case util.Master_metrics_routing:
		clusterId, leader, jsonstr = util.MasterMetricsJson(string(messageBody))
		log.Infof("received message clusterId:%s leader:%d json:%s", clusterId, leader, jsonstr)
	case util.Slave_state_routing:
		log.Debugf("**************** slavestatejson message: %s", string(messageBody))
		array := util.SlaveStateJson(string(messageBody))
		if len(array) != 0 {
			for _, v := range array {
				key := v.ClusterId + "-" + v.App.AppName
				field := v.App.AppId
				value, _ := json.Marshal(v)
				err := cache.WriteHashToRedis(key, field, string(value))
				if err != nil {
					log.Error("writeHashToRedis has err: ", err)
				}
			}
		}
		log.Infof("received slaveStateMessage array: %s", array)
	case util.Marathon_event_routing: // 应用级别的部署监控
		eventType, clusterId, idOrApp, timestamp, currentType, taskId = util.MarathonEventJson(string(messageBody))
		switch eventType {
		case util.Deployment_info:
			if idOrApp != "" && timestamp != "" {
				label := clusterId + "_" + idOrApp
				log.Info("[deployment_info] label: ", label)
				err := cache.WriteStringToRedis(label, currentType)
				if err != nil {
					log.Error("writeToRedis2 has err: ", err)
				}
			}
		case util.Deployment_success:
			if idOrApp != "" && timestamp != "" {
				app, err := cache.ReadFromRedis(idOrApp)
				if err != nil {
					log.Error("readFromRedis has err: ", err)
				}
				label := clusterId + "_" + app
				log.Debug("[deployment_success] label: ", label)
				event := util.MarathonEventMarshal(eventType, timestamp, idOrApp, currentType, taskId)
				log.Debug("[deployment_success] event: ", event)
				err = cache.WriteListToRedis(label, event)
				if err != nil {
					log.Error("marathon event writeToRedis has err: ", err)
				}
			}
		case util.Deployment_failed:
			if idOrApp != "" && timestamp != "" {
				app, err := cache.ReadFromRedis(idOrApp)
				if err != nil {
					log.Error("readFromRedis has err: ", err)
				}
				label := clusterId + "_" + app
				log.Debug("[deployment_failed] label: ", label)
				event := util.MarathonEventMarshal(eventType, timestamp, idOrApp, currentType, taskId)
				log.Debug("[deployment_failed] event: ", event)
				err = cache.WriteListToRedis(label, event)
				if err != nil {
					log.Error("marathon event writeToRedis has err: ", err)
				}
			}
		case util.Deployment_step_success:
			if idOrApp != "" && timestamp != "" && currentType != "" {
				label := clusterId + "_" + idOrApp
				log.Debug("[deployment_step_success] label: ", label)
				event := util.MarathonEventMarshal(eventType, timestamp, idOrApp, currentType, taskId)
				log.Debug("[deployment_step_success] event: ", event)
				err := cache.WriteListToRedis(label, event)
				if err != nil {
					log.Error("marathon event writeToRedis has err: ", err)
				}
			}
		case util.Deployment_step_failure:
			if idOrApp != "" && timestamp != "" && currentType != "" {
				label := clusterId + "_" + idOrApp
				log.Debug("[deployment_step_failure] label: ", label)
				event := util.MarathonEventMarshal(eventType, timestamp, idOrApp, currentType, taskId)
				log.Debug("[deployment_step_failure] event: ", event)
				err := cache.WriteListToRedis(label, event)
				if err != nil {
					log.Error("marathon event writeToRedis has err: ", err)
				}
			}
		case util.Status_update_event:
			if idOrApp != "" && timestamp != "" && currentType != "" {
				label := clusterId + "_" + idOrApp
				log.Debug("[status_update_event] label: ", label)
				event := util.MarathonEventMarshal(eventType, timestamp, idOrApp, currentType, taskId)
				log.Debug("[status_update_event] event: ", event)
				err := cache.WriteListToRedis(label, event)
				if err != nil {
					log.Error("[status_update_event] writeToRedis has err: ", err)
				}
			}
		case util.Destroy_app:
			if idOrApp != "" && clusterId != "" {
				label := clusterId + "_" + idOrApp
				log.Info("[destroy_app] label: ", label)
				err := cache.DeleteRedisByKey(label)
				if err != nil {
					log.Error("[destroy_app] deleteKeyFromRedis has err: ", err)
				}

			}
		}

	}

	if clusterId != "" && jsonstr != "" && leader == 1 {
		label := clusterId + "_" + routingKey
		err := cache.WriteListToRedis(label, jsonstr)
		if err != nil {
			log.Error("writeToRedis has err: ", err)
		}
	}
}

func masterMetrics(ctx *gin.Context) {
	conn := cache.Open()
	defer conn.Close()
	cluster_id := ctx.Param("cluster_id") + "_" + util.Master_metrics_routing
	log.Debug("cluster_id ", cluster_id)

	response := util.MonitorResponse{
		Code: 1,
		Data: nil,
		Err:  "",
	}

	strs, err := redis.Strings(conn.Do("LRANGE", cluster_id, 0, -1))
	if err != nil {
		log.Error("[Master Metrics] got error ", err)
		response.Err = "[Master Metrics] got error " + err.Error()
		ctx.JSON(http.StatusOK, response)
	}
	jsoninterface, err := util.ReturnMessage(util.MonitorMasterMetrics, strs)
	if err != nil {
		log.Error("[Master Metrics] analysis error ", err)
		response.Err = "[Master Metrics] analysis error " + err.Error()
		ctx.JSON(http.StatusOK, response)
	}
	response.Code = 0
	response.Data = *jsoninterface
	ctx.JSON(http.StatusOK, response)
}

func marathonEvent(ctx *gin.Context) {
	conn := cache.Open()
	defer conn.Close()

	response := util.MonitorResponse{
		Code: 1,
		Data: nil,
		Err:  "",
	}

	key := ctx.Param("cluster_id") + "_/" + ctx.Param("app")
	strs, err := redis.Strings(conn.Do("LRANGE", key, 0, -1))
	if err != nil {
		log.Error("[Marathon Event] got error ", err)
		response.Err = "[Marathon Event] got error " + err.Error()
		ctx.JSON(http.StatusOK, response)
	}
	jsoninterface, err := util.ReturnMessage(util.MonitorMarathonEvent, strs)
	if err != nil {
		log.Error("[Marathon Event] analysis error ", err)
		response.Err = "[Marathon Event] analysis error " + err.Error()
		ctx.JSON(http.StatusOK, response)
	}
	response.Code = 0
	response.Data = *jsoninterface
	ctx.JSON(http.StatusOK, response)
}

func appMetrics(ctx *gin.Context) {
	conn := cache.Open()
	defer conn.Close()

	response := util.MonitorResponse{
		Code: 1,
		Data: nil,
		Err:  "",
	}

	key := ctx.Param("cluster_id") + "-" + ctx.Param("app")
	strs, err := redis.Strings(conn.Do("HVALS", key))
	if err != nil {
		log.Error("[App Metrics] got error ", err)
		response.Err = "[App Metrics] got error " + err.Error()
		ctx.JSON(http.StatusOK, response)
	}
	jsoninterface, err := util.ReturnMessage(util.MonitorAppMetrics, strs)
	if err != nil {
		log.Error("[App Metrics] analysis error ", err)
		response.Err = "[App Metrics] analysis error " + err.Error()
		ctx.JSON(http.StatusOK, response)
	}
	response.Code = 0
	response.Data = *jsoninterface
	ctx.JSON(http.StatusOK, response)
}

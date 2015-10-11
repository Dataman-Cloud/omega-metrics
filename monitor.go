package main

import (
	"fmt"
	"github.com/Dataman-Cloud/omega-metrics/cache"
	"github.com/Dataman-Cloud/omega-metrics/config"
	"github.com/Dataman-Cloud/omega-metrics/util"
	log "github.com/Sirupsen/logrus"
	redis "github.com/garyburd/redigo/redis"
	"github.com/gin-gonic/gin"
	"net/http"
)

func init() {
	conn := cache.Open()
	defer conn.Close()
}

func startC() {
	log.Debug("start master metrics mq consumer")
	util.MetricsSubscribe(util.Metrics_exchange, util.Master_metrics_routing, handler)
	util.MetricsSubscribe(util.Metrics_exchange, util.Marathon_event_routing, handler)
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
	var json, clusterId, idOrApp, eventType string
	var leader int
	switch routingKey {
	case util.Master_metrics_routing:
		clusterId, leader, json = util.MasterMetricsJson(string(messageBody))
		log.Infof("received message clusterId:%s leader:%d json:%s", clusterId, leader, json)
	case util.Slave_metrics_routing:
		clusterId, json = util.SlaveMetricsJson(string(messageBody))
		log.Infof("received message id:%s json:%s", clusterId, json)
	case util.Marathon_event_routing: // 应用级别的部署监控
		eventType, clusterId, idOrApp, json = util.MarathonEventJson(string(messageBody))
		fmt.Println("**************** 1 ", util.Marathon_event_routing)
		switch eventType {
		case util.Deployment_info:
			fmt.Println("************ eventtype ", eventType)
			fmt.Println("************ 3 ", idOrApp+json)
			if idOrApp != "" && json != "" {
				label := clusterId + "_" + idOrApp
				log.Info("deployment_info label: ", label)
				err := writeToRedis2(label, json)
				if err != nil {
					log.Error("writeToRedis2 has err: ", err)
				}
			}
		case util.Deployment_success:
			fmt.Println("*********** eventtype ", eventType)
			if idOrApp != "" && json != "" {
				app, err := readFromRedis(idOrApp)
				if err != nil {
					log.Error("readFromRedis has err: ", err)
				}
				label := clusterId + "_" + app
				log.Info("deployment_success label: ", label)
				event := json + " " + app + " " + eventType
				log.Info("deployment_success event: ", event)
				err = writeToRedis(label, event)
				if err != nil {
					log.Error("marathon event writeToRedis has err: ", err)
				}
			}
		case util.Deployment_failed:
			fmt.Println("*********** eventType ", eventType)
			if idOrApp != "" && json != "" {
				app, err := readFromRedis(idOrApp)
				if err != nil {
					log.Error("readFromRedis has err: ", err)
				}
				label := clusterId + "_" + app
				log.Info("deployment_failed label: ", label)
				event := json + " " + app + " " + eventType
				log.Info("deployment_failed event: ", event)
				err = writeToRedis(label, event)
				if err != nil {
					log.Error("marathon event writeToRedis has err: ", err)
				}
			}
		case util.Deployment_step_success:
			fmt.Println("********** eventType ", eventType)
			if idOrApp != "" && json != "" {
				label := clusterId + "_" + idOrApp
				log.Info("deployment_step_success label: ", label)
				event := json + " " + idOrApp + " " + eventType
				log.Debug("deployment_step_success event: ", event)
				err := writeToRedis(label, event)
				if err != nil {
					log.Error("marathon event writeToRedis has err: ", err)
				}
			}
		case util.Deployment_step_failure:
			fmt.Println("********* eventType ", eventType)
			if idOrApp != "" && json != "" {
				label := clusterId + "_" + idOrApp
				log.Info("deployment_step_failure label: ", label)
				event := json + " " + idOrApp + " " + eventType
				log.Debug("deployment_step_failure event: ", event)
				err := writeToRedis(label, event)
				if err != nil {
					log.Error("marathon event writeToRedis has err: ", err)
				}
			}
		}
	}

	if clusterId != "" && json != "" && leader == 1 {
		label := clusterId + "_" + routingKey
		err := writeToRedis(label, json)
		if err != nil {
			log.Error("writeToRedis has err: ", err)
		}
	}
}

func readFromRedis(id string) (string, error) {
	conn := cache.Open()
	defer conn.Close()
	log.Debugf("redis Get key %s", id)
	value, err := redis.String(conn.Do("GET", id))
	return value, err
}

func writeToRedis2(id string, app string) error {
	conn := cache.Open()
	defer conn.Close()
	log.Debugf("redis Set marathon event id %s, app %s", id, app)
	_, err := conn.Do("SETEX", id, config.DefaultTimeout, app)
	return err
}

func writeToRedis(id string, json string) error {
	conn := cache.Open()
	defer conn.Close()
	conf := config.Pairs()
	log.Debugf("redis LPUSH id %s, json %s", id, json)
	conn.Send("LPUSH", id, json)
	log.Debugf("redis EXPIRE id %s, json %s", id, json)
	conn.Send("EXPIRE", id, config.DefaultTimeout)
	_, err := conn.Do("LTRIM", id, 0, conf.Cache.Llen)
	return err
}

func masterMetrics(ctx *gin.Context) {
	conn := cache.Open()
	defer conn.Close()
	cluster_id := ctx.Param("cluster_id") + "_" + util.Master_metrics_routing
	log.Debug("cluster_id ", cluster_id)

	response := util.MonitorResponse{
		Code: "1",
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
	response.Code = "0"
	response.Data = *jsoninterface
	ctx.JSON(http.StatusOK, response)
}

func marathonEvent(ctx *gin.Context) {
	conn := cache.Open()
	defer conn.Close()

	response := util.MonitorResponse{
		Code: "1",
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

	response.Data = strs
	response.Code = "0"
	ctx.JSON(http.StatusOK, response)
}

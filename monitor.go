package main

import (
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
}

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
	switch routingKey {
	case util.Master_metrics_routing:
		id, json := util.MasterMetricsJson(string(messageBody))
		if id != "" && json != "" {
			label := id + "_" + routingKey
			log.Debug(label)
			writeToRedis(label, json)
		}
	case util.Slave_metrics_routing:
		id, json := util.SlaveMetricsJson(string(messageBody))
		if id != "" && json != "" {
			label := id + "_" + routingKey
			writeToRedis(label, json)
		}
	}
}

func writeToRedis(id string, json string) {
	conn := cache.Open()
	defer conn.Close()
	log.Debug("write to redis")
	conn.Send("LPUSH", id, json)
	conn.Send("EXPIRE", id, config.DefaultTimeout)
	_, err := conn.Do("LTRIM", id, 0, 180)
	if err != nil {
		log.Errorf("LPUSH key:%s value:%s is wrong", id, json)
		log.Errorln("[writeToRedis] error is ", err)
	}
}

func masterMetrics(ctx *gin.Context) {
	conn := cache.Open()
	defer conn.Close()
	cluster_id := ctx.Param("cluster_id") + "_" + util.Master_metrics_routing
	log.Debug("cluster_id ", cluster_id)
	strs, err := redis.Strings(conn.Do("LRANGE", cluster_id, 0, -1))
	if err != nil {
		log.Error("[Master Metrics] got error ", err)
		jsoninterface := util.ReturnMessage("1", nil, "[MasterMetrics] got error")
		ctx.JSON(http.StatusOK, jsoninterface)
	}
	jsoninterface := util.ReturnMessage("0", strs, "")
	log.Infof("Got master metrics %+v", jsoninterface)
	ctx.JSON(http.StatusOK, jsoninterface)
}

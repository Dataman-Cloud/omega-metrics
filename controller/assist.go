package controller

import (
	"net/http"
	"time"

	"github.com/Dataman-Cloud/omega-metrics/cache"
	"github.com/Dataman-Cloud/omega-metrics/config"
	"github.com/Dataman-Cloud/omega-metrics/db"
	"github.com/Dataman-Cloud/omega-metrics/util"
	log "github.com/cihub/seelog"
	"github.com/gin-gonic/gin"
)

var monitor Monitor
var lastHealthCheckTime int64

type Monitor struct {
	OmegaMetrics HealthStatus `json:"omegaMetrics"`
	Redis        HealthStatus `json:"redis"`
	RabbitMQ     HealthStatus `json:"rabbitMQ"`
	InfluxDbTcp  HealthStatus `json:"influxDbTcp"`
	InfluxDbUdp  HealthStatus `json: "influxDbUdp"`
}

type HealthStatus struct {
	Status uint8   `json:"status"`
	Time   float64 `json:"time"`
}

func HealthCheck(ctx *gin.Context) {
	conf := config.Pairs()

	// health check has interval
	var duration int
	var start time.Time
	if conf.HealthCheckInterval != 0 {
		duration = conf.HealthCheckInterval
	} else {
		duration = config.DefaultHealthCheckInterval
	}
	if time.Now().Unix()-lastHealthCheckTime < int64(duration) {
		ctx.JSON(http.StatusOK, monitor)
		return
	}
	lastHealthCheckTime = time.Now().Unix()
	metricsHealth := true
	start = time.Now()

	// redis check
	err := cache.WriteStringToRedis("health", "check", 2)
	if err != nil {
		log.Error(err)
		monitor.Redis.Status = 1
		monitor.Redis.Time = Milliseconds(time.Since(start))
		metricsHealth = false
	} else {
		monitor.Redis.Status = 0
		monitor.Redis.Time = Milliseconds(time.Since(start))
	}

	// rabbitMQ check
	start = time.Now()
	err = util.Publish(util.ExchangeCluster, util.RoutingHealth, "health check")
	if err != nil {
		log.Error(err)
		monitor.RabbitMQ.Status = 1
		monitor.RabbitMQ.Time = Milliseconds(time.Since(start))
		metricsHealth = false
	} else {
		monitor.RabbitMQ.Status = 0
		monitor.RabbitMQ.Time = Milliseconds(time.Since(start))
	}

	// influxdb udp client check
	durationInfUdp, err := db.UdpClientHealthCheck()
	if err != nil {
		log.Error("health check InfluxDbUdp got error: ", err)
		monitor.InfluxDbUdp.Status = 1
	} else {
		monitor.InfluxDbUdp.Status = 0
	}
	monitor.InfluxDbUdp.Time = Milliseconds(durationInfUdp)

	// influxdb http client check
	durationInfHttp, err := db.HttpClientHealthCheck()
	if err != nil {
		log.Error("health check InfluxDbTcp got error: ", err)
		monitor.InfluxDbTcp.Status = 1
	} else {
		monitor.InfluxDbTcp.Status = 0
	}
	monitor.InfluxDbTcp.Time = Milliseconds(durationInfHttp)

	// OmegaMetrics check
	if metricsHealth {
		monitor.OmegaMetrics.Status = 0
		monitor.OmegaMetrics.Time = Milliseconds(time.Since(start))
	} else {
		monitor.OmegaMetrics.Status = 1
		monitor.OmegaMetrics.Time = Milliseconds(time.Since(start))
	}

	ctx.JSON(http.StatusOK, monitor)
}

package controller

import (
	"fmt"
	"github.com/Dataman-Cloud/health_checker"
	"github.com/Dataman-Cloud/omega-metrics/config"
	"github.com/gin-gonic/gin"
	"net/http"
)

func HealthCheck(ctx *gin.Context) {
	checker := health_checker.NewHealthChecker("omega-metrics")
	conf := config.Pairs()
	redisDsn := fmt.Sprintf("%s:%d",
		conf.Cache.Host, conf.Cache.Port)
	checker.AddCheckPoint("redis", redisDsn, nil, nil)

	influxdbDsn := fmt.Sprintf("%s:%s@http://%s:%d/%s?parseTime=true&loc=Local",
		conf.Db.User, conf.Db.Password, conf.Db.Host, conf.Db.Port, conf.Db.Database)
	checker.AddCheckPoint("influxdb", influxdbDsn, nil, nil)

	mqDsn := fmt.Sprintf("amqp://%s:%s@%s:%d/",
		conf.Mq.User, conf.Mq.Password, conf.Mq.Host, conf.Mq.Port)

	checker.AddCheckPoint("mq", mqDsn, nil, nil)

	ctx.JSON(http.StatusOK, checker.Check())
}

/** Omega-app service http api
**/
package main

import (
	"os"
	"os/signal"
	"runtime"

	"github.com/Dataman-Cloud/omega-metrics/cache"
	"github.com/Dataman-Cloud/omega-metrics/config"
	"github.com/Dataman-Cloud/omega-metrics/util"
	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
)

func init() {
	util.InitLog()
	util.InitMQ()
	cache.InitCache()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	go func() {
		log.Info("received exit signal", <-signals)
		destroy()
		os.Exit(0)
	}()

}

func destroy() {
	log.Info("destroying ...")
	cache.DestroyCache()
	util.DestroyMQ()
	util.DestroyLog()
}

func main() {
	initEnv()
	monitor()
	defer destroy()
}

func initEnv() {
	conf := config.Pairs()
	numCPU := conf.NumCPU
	runtime.GOMAXPROCS(numCPU)
	log.Info("Runing with ", numCPU, " CPUs")
}

func monitor() {
	startC()
	gin.SetMode(gin.ReleaseMode)
	log.Info("[monitor] is up")
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	monitorGroup := router.Group("/api/v1")
	{
		monitorGroup.GET("/metrics/cluster/:cluster_id", masterMetrics)
	}
	router.Run(":9005")

}

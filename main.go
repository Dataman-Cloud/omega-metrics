/** Omega-app service http api
**/
package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/Dataman-Cloud/omega-metrics/cache"
	"github.com/Dataman-Cloud/omega-metrics/config"
	"github.com/Dataman-Cloud/omega-metrics/controller"
	"github.com/Dataman-Cloud/omega-metrics/logger"
	"github.com/Dataman-Cloud/omega-metrics/util"
	log "github.com/cihub/seelog"
	"github.com/gin-gonic/gin"
)

func init() {
	logger.LoadLogConfig()
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
	log.Flush()
}

func main() {
	initEnv()
	initServer()
	defer destroy()
}

func initEnv() {
	conf := config.Pairs()
	numCPU := conf.NumCPU
	runtime.GOMAXPROCS(numCPU)
	log.Info("Runing with ", numCPU, " CPUs")
}

func initServer() {
	startC()
	gin.SetMode(gin.ReleaseMode)
	log.Info("[monitor] is up")
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	// options Handler
	// router.Use(controller.OptionHandler)
	router.GET("/", func(c *gin.Context) {
		c.String(200, "pass")
	})
	// healthcheck
	router.GET("/api/v3/health/metrics", controller.HealthCheck)

	monitorGroup := router.Group("/api/v3")
	{
		monitorGroup.GET("/clusters/:cluster_id/metrics", controller.ClusterMetricsHandler)
		monitorGroup.GET("/clusters/:cluster_id/apps/:app/metrics", controller.AppMetricsHandler)
		// Request the single application monitor data
		monitorGroup.GET("/clusters/:cluster_id/apps/:app/monitor", controller.AppMonitorHandler)
		monitorGroup.GET("/clusters/:cluster_id/apps/:app/session", controller.RequestRate)
		// get host app instance info by cluster id and host ip
		monitorGroup.GET("/clusters/:cluster_id/ip/:ip/instance", controller.HostInstanceHandler)
	}

	conf := config.Pairs()
	addr := fmt.Sprintf("%s:%d", conf.Host, conf.Port)
	server := &http.Server{
		Addr:           addr,
		Handler:        router,
		ReadTimeout:    20 * time.Second,
		WriteTimeout:   20 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	err := server.ListenAndServe()
	if err != nil {
		log.Error("Can't start monitor server: ", err)
		panic(-1)
	}
}

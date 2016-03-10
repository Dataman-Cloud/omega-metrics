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
	// healthcheck
	router.GET("/", func(c *gin.Context) {
		c.String(200, "pass")
	})
	router.GET("/api/v3/health/metrics", HealthCheck)

	monitorGroup := router.Group("/api/v3")
	{
		monitorGroup.GET("/clusters/:cluster_id/metrics", masterMetrics)
		monitorGroup.GET("/clusters/:cluster_id/apps/:app/metrics", appMetrics)
		monitorGroup.GET("/clusters/:cluster_id/apps/:app/monitor", appMonitor)
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

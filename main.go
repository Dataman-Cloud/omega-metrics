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
	redis "github.com/garyburd/redigo/redis"
	"github.com/gin-gonic/gin"
)

func init() {
	logger.LoadLogConfig()
	util.InitMQ()
	util.InitDB()
	cache.InitCache()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	go func() {
		log.Info("received exit signal", <-signals)
		destroy()
		os.Exit(0)
	}()
	go func() {
		ticker := time.NewTicker(20 * time.Second)
		for {
			select {
			case <-ticker.C:
				log.Debug("--------------------> begin to check auto scale")
				conn := cache.Open()
				defer conn.Close()
				token, _ := redis.String(conn.Do("GET", "AutoScaleToken"))
				log.Debug("token==========", token)
				if token != "" {
					log.Debug("into token")
					go AutoScale(token)
				}
			}
		}
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
	router.Use(gin.Logger(), gin.Recovery(), SetHeader)
	// healthcheck
	router.GET("/", func(c *gin.Context) {
		c.String(200, "pass")
	})

	monitorGroup := router.Group("/api/v1")
	{
		monitorGroup.GET("/metrics/cluster/:cluster_id", masterMetrics)
		monitorGroup.GET("/event/:cluster_id/:app", marathonEvent)
		monitorGroup.GET("/appmetrics/cluster/:cluster_id/app/:app", appMetrics)
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

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
	"github.com/howeyc/fsnotify"
)

var isAutoScaling bool

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

	conf := config.Pairs()
	isAutoScaling = conf.Scale.AutoScaling
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-ticker.C:
				if isAutoScaling {
					log.Info("autoScaling is true and begin to get token...")
					token, _ := cache.ReadFromRedis("AutoScaleToken")
					log.Debug("token==========", token)
					if token != "" {
						log.Debug("begin to check autoScaling...")
						go AutoScale(token)
					}
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
	go initNotify()
	monitor()
	defer destroy()
}

func initEnv() {
	conf := config.Pairs()
	numCPU := conf.NumCPU
	runtime.GOMAXPROCS(numCPU)
	log.Info("Runing with ", numCPU, " CPUs")
}
func initNotify() {
	log.Info("initNotify....................")
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Error("[main initNotify] NewWatcher error: ", err)
	}
	done := make(chan bool)
	go func() {
		for {
			select {
			case ev := <-watcher.Event:
				log.Info("event:", ev)
				if ev.IsModify() {
					config.Init()
					conf := config.Pairs()
					if conf.Scale.AutoScaling {
						log.Debug("auto scaling true")
						isAutoScaling = true
					} else {
						log.Debug("auto scaling false")
						isAutoScaling = false
					}
				}
			case err := <-watcher.Error:
				log.Error("error:", err)
			}
		}
	}()

	err = watcher.Watch("./conf")
	if err != nil {
		log.Error("[main initNotify---------------------------] watch error: ", err)
	}

	<-done
	watcher.Close()
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

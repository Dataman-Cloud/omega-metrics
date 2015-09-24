/** Omega-app service http api
**/
package main

import (
	//	"fmt"
	//	"net/http"
	"os"
	"os/signal"
	"runtime"
	//	"time"

	"github.com/Dataman-Cloud/omega-app/cache"
	"github.com/Dataman-Cloud/omega-app/config"
	"github.com/Dataman-Cloud/omega-app/util"
	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
)

func init() {
	util.InitLog()
	//	util.InitDB()
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
	//	util.DestroyDB()
	util.DestroyLog()
}

func main() {
	initEnv()
	//start()
	monitor()
	defer destroy()
}

func initEnv() {
	conf := config.Pairs()
	numCPU := conf.NumCPU
	runtime.GOMAXPROCS(numCPU)
	log.Info("Runing with ", numCPU, " CPUs")
}

/*func start() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery(), authenticate)

	router.GET("/", func(c *gin.Context) {
		c.String(200, "pass")
	})

	appGroup := router.Group("/api/v1/applications")
	{
		appGroup.POST("/deploy", deployApp)
	}
	conf := config.Pairs()
	addr := fmt.Sprintf("%s:%d", conf.Host, conf.Port)
	server := &http.Server{
		Addr:           addr,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal("can't start server: ", err)
		panic(-1)
	}
}*/

func monitor() {
	//startP()
	startC()
	gin.SetMode(gin.ReleaseMode)
	log.Info("[monitor] is up")
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	monitorGroup := router.Group("/api/v1")
	{
		monitorGroup.GET("/metrics/cluster/:cluster_id", masterMetrics)
		monitorGroup.GET("/masterState/:masterid", masterState)
		monitorGroup.GET("/slaveMetrics/:slaveid", slaveMetrics)
	}
	router.Run(":6666")

}

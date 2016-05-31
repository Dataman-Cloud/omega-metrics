package controller

import (
	"errors"
	"strconv"
	"time"

	"github.com/Dataman-Cloud/omega-metrics/db"
	"github.com/gin-gonic/gin"
)

func AppMonitorHandler(c *gin.Context) {
	clusterId := c.Param("cluster_id")
	if clusterId == "" {
		ReturnError(c, InvalidParams, errors.New("cluster id is empty"))
		return
	}

	appName := c.Param("app")
	if appName == "" {
		ReturnError(c, InvalidParams, errors.New("appName is empty"))
		return
	}

	var startTime, endTime int64
	var err error
	startTimeStr := c.Query("starttime")
	if startTime, err = strconv.ParseInt(startTimeStr, 10, 64); err != nil {
		startTime = time.Now().Add(-1 * time.Hour).UnixNano()
	}

	endTimeStr := c.Query("endtime")
	if endTime, err = strconv.ParseInt(endTimeStr, 10, 64); err != nil {
		endTime = time.Now().UnixNano()
	}

	var results []map[string]interface{}
	results, err = db.QueryMetricsInfo(clusterId, appName, startTime, endTime)
	if err != nil {
		ReturnError(c, DbQueryError, err)
		return
	}

	ReturnOk(c, results)
	return
}

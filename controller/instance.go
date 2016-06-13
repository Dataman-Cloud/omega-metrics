package controller

import (
	"encoding/json"
	"errors"

	"github.com/Dataman-Cloud/omega-metrics/cache"
	"github.com/Dataman-Cloud/omega-metrics/util"
	"github.com/gin-gonic/gin"
)

func HostInstanceHandler(c *gin.Context) {
	clusterId := c.Param("cluster_id")
	if clusterId == "" {
		ReturnError(c, InvalidParams, errors.New("cluster id is empty"))
		return
	}

	ip := c.Param("ip")
	if ip == "" {
		ReturnError(c, InvalidParams, errors.New("ip is empty"))
		return
	}

	instances, err := GetHostInstanceInfo(clusterId, ip)
	if err != nil {
		ReturnError(c, DbQueryError, err)
		return
	}

	ReturnOk(c, instances)
	return
}

// get host app instance info by clusterid and slave ip
func GetHostInstanceInfo(clusterId string, ip string) ([]util.HostInstance, error) {
	var instances []util.HostInstance
	key := clusterId + ":" + ip
	value, err := cache.ReadFromRedis(key)
	if err != nil {
		return instances, err
	}

	if err := json.Unmarshal([]byte(value), &instances); err != nil {
		return instances, err
	}

	return instances, nil
}

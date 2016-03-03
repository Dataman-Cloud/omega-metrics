package db

import (
	"fmt"
	//"sync"
	//"github.com/Dataman-Cloud/omega-metrics/config"
	log "github.com/cihub/seelog"
	"github.com/influxdata/influxdb/client/v2"
	"github.com/fatih/structs"
	"github.com/Dataman-Cloud/omega-metrics/util"
)

//var pool *redis.Pool
const (
  DB = "shurenyun"
	Addr = "http://influxdb:8086"
	username = "root"
	password = "root"
)

func WriteStringToInfluxdb(serie string, tags_value string, fields_value util.SlaveStateMar) error {

	conn, err := client.NewHTTPClient(client.HTTPConfig{
			Addr: Addr,
			Username: username,
			Password: password,
	})
	if err != nil {
		log.Error("Error creating Influxdb Client: ", err.Error())
	}
	defer conn.Close()

	// Create a new point batch
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
			Database:  DB,
			Precision: "s",
	})

	  type SlaveStateMar struct {
		  Timestamp     int64   `json:"Timestamp"`
		  ClusterID     string  `json:"ClusterID"`
		  SlaveID       string  `json:"SlaveID"`
			InstanceID    string  `json:"InstanceID"`
      AppName       string  `json:"AppName"`
			TaskID        string  `json:"TaskID"`
		  ContainerID   string  `json:"ContainerID"`
		  CpuUsedCores  float64 `json:"CpuUsedCores"`
		  CpuShareCores float64 `json:"CpuShareCores"`
		  MemoryTotal   uint64  `json:"MemoryTotal"`
		  MemoryUsed    uint64  `json:"MemoryUsed"`
	  }

		  InstanceID := fields_value.App.AppId
			AppName := fields_value.App.AppName

      fileds_new := &SlaveStateMar{
				Timestamp: fields_value.Timestamp,
				ClusterID: fields_value.ClusterId,
				SlaveID: fields_value.App.Slave_id,
				InstanceID: fields_value.App.AppId,
				AppName: fields_value.App.AppName,
				TaskID: fields_value.App.Task_id,
				CpuUsedCores: fields_value.CpuUsedCores,
				CpuShareCores: fields_value.CpuShareCores,
				MemoryTotal: fields_value.MemoryTotal,
				MemoryUsed: fields_value.MemoryUsed,
			}

		  fields := structs.Map(fileds_new)

		  fmt.Println("serie: %s", serie)
		  fmt.Println("fields: %s", fields)
			fmt.Println("InstanceID: %s", InstanceID)
			fmt.Println("AppName: %s", AppName)
		  //Create a point and add to batch
		  tags := map[string]string{"InstanceID": InstanceID, "AppName": AppName}

			pt, err := client.NewPoint(serie, tags, fields)
			if err != nil {
				log.Error("Error: ", err.Error())
			}

			bp.AddPoint(pt)


	// Write the batch
	conn.Write(bp)
	log.Infof("Write String to Influxdb %s, Serie %s", DB, serie)
	return nil
}

func WriteStringToInfluxdbSlaveState(serie string, tags_value string, fields_value util.SlaveStateMar) error {
	containerId := fields_value.ContainerId

	fmt.Println("serie: %s", serie)
	fmt.Println("containerId: %s", containerId)
	fmt.Println("fields_value: %s", fields_value)
	return nil
}

func WriteStringToInfluxdbMasterState(serie string, tags_value string, fields_value string) error {
	fmt.Println("serie: %s", serie)
	fmt.Println("tags_value: %s", tags_value)
	fmt.Println("fields_value: %s", fields_value)
	return nil
}

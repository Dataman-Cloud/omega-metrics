package db

import (
	"fmt"
	//"sync"
	//"github.com/Dataman-Cloud/omega-metrics/config"
	"encoding/json"
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

func WriteStringToInfluxdb(serie string, appname string, appid string, fields_value string) error {

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

	var slave_mar util.SlaveStateMar
        json.Unmarshal([]byte(fields_value), &slave_mar)

	fields := structs.Map(&slave_mar)

  fields["ContainerName"] = fields["ContainerId"]

	MemoryTotal, ok := fields["MemoryTotal"]
	if ok {
		fields["MemoryTotal"] = float64(MemoryTotal.(uint64))
	}

	MemoryUsed, ok := fields["MemoryUsed"]
	if ok {
		fields["MemoryUsed"] = float64(MemoryUsed.(uint64))
	}

	clusteid, _ := fields["ClusterId"].(string)

  delete(fields, "App")
	delete(fields, "ContainerId")
	delete(fields, "ClusterId")

	tags := map[string]string{"appname": appname, "instance": appid, "clusteid": clusteid}

	fmt.Println("fields after: %s", fields)


			pt, err := client.NewPoint(serie, tags, fields)
			if err != nil {
				log.Error("Error: ", err.Error())
			}

			bp.AddPoint(pt)

			fmt.Println("influxdb bp: %s", bp)

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


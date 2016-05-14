package db

import (
	"encoding/json"
	"fmt"
	"github.com/Dataman-Cloud/omega-metrics/config"
	"github.com/Dataman-Cloud/omega-metrics/util"
	log "github.com/cihub/seelog"
	"github.com/fatih/structs"
	"github.com/influxdata/influxdb/client/v2"
	"time"
)

func WriteContainerInfoToInflux(conInfo *util.SlaveStateMar) error {
	// Receive the variable serie, appname, appid and fiels_value. Write the fileds_value
	// into the name of serie of database conf.Db.Database, and set the tags with
	// appname, appid and clusterid.
	conf := config.Pairs()
	addr := fmt.Sprintf("%s:%d", conf.Db.Host, conf.Db.Port)
	database := fmt.Sprintf("%s", conf.Db.Database)
	conn, err := client.NewUDPClient(
		client.UDPConfig{
			Addr: addr,
		})
	if err != nil {
		log.Error("Error creating Influxdb Client: ", err.Error())
	} else {
		log.Infof("Connected the Influxdb Server %s, Database %s", addr, database)
	}
	defer conn.Close()

	// Create a new point batch
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  database,
		Precision: "s",
	})

	fields := structs.Map(&conInfo)

	fields["ContainerName"] = fields["ContainerId"]

	MemoryTotal, ok := fields["MemoryTotal"]
	if ok {
		fields["MemoryTotal"] = float64(MemoryTotal.(uint64))
	}

	MemoryUsed, ok := fields["MemoryUsed"]
	if ok {
		fields["MemoryUsed"] = float64(MemoryUsed.(uint64))
	}

	delete(fields, "App")
	delete(fields, "ContainerId")
	delete(fields, "ClusterId")

	appInfo := conInfo.App
	fields["SlaveId"] = appInfo.SlaveId
	tags := map[string]string{"appname": appInfo.AppName, "instance": appInfo.AppId, "clusterid": app.ClusterId}

	timestampInterface, ok := fields["Timestamp"]
	var timestamp time.Time
	if ok {
		timestamp = timestampInterface.(time.Time)
	} else {
		timestamp = time.Now()
	}

	pt, err := client.NewPoint(serie, tags, fields, timestamp)
	if err != nil {
		log.Error("Error: ", err.Error())
	}

	bp.AddPoint(pt)
	// Write the batch
	err = conn.Write(bp)
	if err != nil {
		log.Error("Error: ", err.Error())
	} else {
		log.Infof("Write String to Influxdb %s, Serie %s", database, serie)
	}
	return err
}

func InfluxdbClient_Query(command string) (client.Response, error) {
	// Receive the command and return the query respose.
	conf := config.Pairs()
	addr := fmt.Sprintf("http://%s:%d", conf.Db.Host, conf.Db.Port)
	username := fmt.Sprintf("%s", conf.Db.User)
	password := fmt.Sprintf("%s", conf.Db.Password)
	database := fmt.Sprintf("%s", conf.Db.Database)
	timeout := time.Second * 60
	conn, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     addr,
		Username: username,
		Password: password,
		Timeout:  timeout,
	})
	if err != nil {
		log.Error("Error creating Influxdb Client: ", err.Error())
	}
	defer conn.Close()

	q := client.Query{
		Command:  command,
		Database: database,
	}

	response, err := conn.Query(q)
	return *response, err
}

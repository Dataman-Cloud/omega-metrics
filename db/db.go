package db

import (
	"encoding/json"
	"fmt"
	"github.com/Dataman-Cloud/omega-metrics/config"
	"github.com/Dataman-Cloud/omega-metrics/util"
	log "github.com/cihub/seelog"
	"github.com/fatih/structs"
	"github.com/influxdata/influxdb/client/v2"
)

func WriteStringToInfluxdb(serie string, appname string, appid string, fields_value string) error {

	conf := config.Pairs()
	addr := fmt.Sprintf("http://%s:%d", conf.Db.Host, conf.Db.Port)
	username := fmt.Sprintf("%s", conf.Db.User)
	password := fmt.Sprintf("%s", conf.Db.Password)
	database := fmt.Sprintf("%s", conf.Db.Database)
	conn, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     addr,
		Username: username,
		Password: password,
	})
	if err != nil {
		log.Error("Error creating Influxdb Client: ", err.Error())
	}
	defer conn.Close()

	// Create a new point batch
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  database,
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

	clusterid, _ := fields["ClusterId"].(string)

	delete(fields, "App")
	delete(fields, "ContainerId")
	delete(fields, "ClusterId")

	tags := map[string]string{"appname": appname, "instance": appid, "clusterid": clusterid}

	pt, err := client.NewPoint(serie, tags, fields)
	if err != nil {
		log.Error("Error: ", err.Error())
	}

	bp.AddPoint(pt)

	fmt.Println("influxdb bp: %s", bp)

	// Write the batch
	conn.Write(bp)
	log.Infof("Write String to Influxdb %s, Serie %s", database, serie)
	return nil
}

func InfluxdbClient_Query(command string) (client.Response, error) {
	conf := config.Pairs()
  addr := fmt.Sprintf("http://%s:%d", conf.Db.Host, conf.Db.Port)
	username := fmt.Sprintf("%s", conf.Db.User)
	password := fmt.Sprintf("%s", conf.Db.Password)
	database := fmt.Sprintf("%s", conf.Db.Database)
	conn, err := client.NewHTTPClient(client.HTTPConfig{
			Addr: addr,
			Username: username,
			Password: password,
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
	fmt.Println(response.Results)
	return *response, err
}

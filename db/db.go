package db

import (
	"fmt"
	"reflect"

	"github.com/Dataman-Cloud/omega-metrics/config"
	"github.com/Dataman-Cloud/omega-metrics/util"
	log "github.com/cihub/seelog"
	"github.com/fatih/structs"
	"github.com/influxdata/influxdb/client/v2"
	"time"
)

var (
	InfluxAddr     string = "localhost:5008"
	InfulxDataBase string = "shurenyun"
)

func init() {
	conf := config.Pairs()
	InfluxAddr = fmt.Sprintf("%s:%d", conf.Db.Host, conf.Db.Port)
	InfulxDataBase = conf.Db.Database
}

// create new influxdb udp client
func CreateInfluxUDPClient() client.Client {
	conn, err := client.NewUDPClient(
		client.UDPConfig{
			Addr: InfluxAddr,
		})
	if err != nil {
		log.Error("Error creating Influxdb Client: ", err.Error())
	}

	return conn
}

// convert a struct instance to influxdb tags and fields. influx key use struct tag json
// use struct influx tag 'influx' to differentiate tag and field
func BuildInfluxData(instance interface{}) (tags map[string]string, fields map[string]interface{}) {
	tags = make(map[string]string)
	fields = make(map[string]interface{})

	val := reflect.ValueOf(instance).Elem()
	var str string

	for i := 0; i < val.NumField(); i++ {
		typefield := val.Type().Field(i)
		influxTag := typefield.Tag.Get("influx")
		if influxTag == "" {
			continue
		}

		jsonTag := typefield.Tag.Get("json")
		if jsonTag == "" {
			continue
		}

		valueFiled := val.Field(i).Interface()

		if influxTag == "tag" {
			switch value := valueFiled.(type) {
			case string:
				str = value
			}
			tags[jsonTag] = str
		} else if influxTag == "field" {
			fields[jsonTag] = valueFiled
		}
	}
	return
}

func WriteAppReqInfoToInflux(appReq *util.InfluxAppRequestInfo) error {
	conn := CreateInfluxUDPClient()
	if conn == nil {
		return fmt.Errorf("Create influx udp client failed connect is nil")
	}
	defer conn.Close()

	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  InfulxDataBase,
		Precision: "s",
	})
	tags, fields := BuildInfluxData(appReq)

	pt, err := client.NewPoint(config.AppRequestInfoSerie, tags, fields)
	if err != nil {
		log.Error("Create newPoint app req info for influx got error: ", err)
		return err
	}

	bp.AddPoint(pt)
	if err := conn.Write(bp); err != nil {
		log.Error("Write app req info to influx got error: ", err)
		return err
	}

	return nil
}

// write container monitor data to influxdb
func WriteContainerInfoToInflux(conInfo *util.SlaveStateMar) error {
	// Receive the variable serie, appname, appid and fiels_value. Write the fileds_value
	// into the name of serie of database conf.Db.Database, and set the tags with
	// appname, appid and clusterid.
	conn := CreateInfluxUDPClient()
	if conn == nil {
		return fmt.Errorf("Create influx udp client failed connect is nil")
	}
	defer conn.Close()

	// Create a new point batch
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  InfulxDataBase,
		Precision: "s",
	})

	fields := structs.Map(conInfo)

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
	tags := map[string]string{"appname": appInfo.AppName, "instance": appInfo.AppId, "clusterid": appInfo.ClusterId}

	timestampInterface, ok := fields["Timestamp"]
	var timestamp time.Time
	if ok {
		timestamp = timestampInterface.(time.Time)
	} else {
		timestamp = time.Now()
	}

	pt, err := client.NewPoint(config.ContainerMonitorSerie, tags, fields, timestamp)
	if err != nil {
		log.Error("Error: ", err.Error())
	}

	bp.AddPoint(pt)
	// Write the batch
	err = conn.Write(bp)
	if err != nil {
		log.Error("Error: ", err.Error())
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

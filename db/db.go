package db

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/Dataman-Cloud/omega-metrics/config"
	"github.com/Dataman-Cloud/omega-metrics/util"
	log "github.com/cihub/seelog"
	"github.com/fatih/structs"
	"github.com/influxdata/influxdb/client/v2"
	"github.com/influxdata/influxdb/models"
)

const (
	DefaultHttpTimeout = 15 * time.Second
)

var (
	InfluxAddr     string = "localhost:5008"
	InfulxDataBase string = "shurenyun"
	HttpInfluxAddr string = "http://localhost:5008"
	InfulxUserName string = "shurenyun"
	InfulxPassword string = "shurenyun"
)

func init() {
	conf := config.Pairs()
	InfluxAddr = fmt.Sprintf("%s:%d", conf.Db.Host, conf.Db.Port)
	HttpInfluxAddr = fmt.Sprintf("http://%s", InfluxAddr)
	InfulxDataBase = conf.Db.Database
	InfulxUserName = conf.Db.User
	InfulxPassword = conf.Db.Password
}

// create new influxdb udp client
func CreateInfluxUDPClient() (client.Client, error) {
	return client.NewUDPClient(
		client.UDPConfig{
			Addr: InfluxAddr,
		})
}

// create a new http influx client
func CreateInfluxHttpClient() (client.Client, error) {
	return client.NewHTTPClient(client.HTTPConfig{
		Addr:     HttpInfluxAddr,
		Username: InfulxUserName,
		Password: InfulxPassword,
		Timeout:  DefaultHttpTimeout,
	})
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

// Write app request info to influxdb
func WriteAppReqInfoToInflux(appReq *util.InfluxAppRequestInfo) error {
	conn, err := CreateInfluxUDPClient()
	if err != nil {
		return err
	}

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
	conn, err := CreateInfluxUDPClient()
	if err != nil {
		return err
	}

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
	conn, err := CreateInfluxHttpClient()
	if err != nil {
		log.Error("Error creating Influxdb Client: ", err.Error())
	}
	defer conn.Close()

	q := client.Query{
		Command:  command,
		Database: InfulxDataBase,
	}

	response, err := conn.Query(q)
	return *response, err
}

// Query from influx interface return influx row object
func Query(sql string) (results []map[string]interface{}, err error) {
	httpClient, err := CreateInfluxHttpClient()
	if err != nil {
		err = errors.New("Create influx http client got error: " + err.Error())
		return
	}

	query := client.NewQuery(sql, InfulxDataBase, "ns")
	response, err := httpClient.Query(query)
	if err != nil {
		err = errors.New("Query influxdb got error: " + err.Error())
		return
	}

	if response.Error() != nil {
		err = errors.New("Query influxdb response got error: " + response.Error().Error())
		return
	}

	if len(response.Results) < 1 {
		return
	}

	series := response.Results[0].Series
	if len(series) < 1 {
		return
	}

	results = ConvertSeriesToMap(series[0])
	return
}

// convert an influxdb row value to map
func ConvertSeriesToMap(row models.Row) (results []map[string]interface{}) {
	columns := row.Columns
	values := row.Values
	tags := row.Tags
	fieldNum := len(columns)
	if fieldNum < 1 {
		return
	}

	for index, value := range values {
		event := make(map[string]interface{})
		if len(value) != fieldNum {
			break
		}
		for i := 0; i < fieldNum; i++ {
			event[columns[i]] = value[i]
		}
		for k, v := range tags {
			event[k] = v
		}
		event["index"] = index + 1
		results = append(results, event)
	}

	return
}

// query requset info from influxdb by cluster id appname start time and end time
func QueryReqInfo(cid string, appname string, sTime int64, eTime int64) ([]map[string]interface{}, error) {
	formatSql := `select * from app_req_rate where clusterid = '%s' and appname = '%s' and time >= %d and time <= %d order by time`
	sql := fmt.Sprintf(formatSql, cid, appname, sTime, eTime)
	return Query(sql)
}

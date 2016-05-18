package util

import (
	"encoding/json"
	"errors"
	"reflect"

	log "github.com/cihub/seelog"
)

const (
	Deployment_success      = "deployment_success"
	Deployment_failed       = "deployment_failed"
	Deployment_info         = "deployment_info"
	Deployment_step_success = "deployment_step_success"
	Deployment_step_failure = "deployment_step_failure"
	Status_update_event     = "status_update_event"
	Destroy_app             = "destroy_app"
)

var parserTypeMappings map[string]reflect.Type

func init() {
	recognizedTypes := []interface{}{
		MasterMetricsMar{},
		SlaveStateMar{},
	}

	parserTypeMappings = make(map[string]reflect.Type)
	for _, recognizedType := range recognizedTypes {
		parserTypeMappings[reflect.TypeOf(recognizedType).Name()] = reflect.TypeOf(recognizedType)
	}
}

func NewOfType(typ string) (interface{}, bool) {
	rtype, ok := parserTypeMappings[typ]
	if !ok {
		return nil, false
	}

	return reflect.New(rtype).Interface(), true
}

func ReturnMessage(typ string, strs []string) (*[]interface{}, error) {
	var monitorDatas []interface{}
	for _, str := range strs {
		monitorType, ok := NewOfType(typ)
		if !ok {
			return nil, errors.New(typ + " is not support type")
		}
		err := json.Unmarshal([]byte(str), &monitorType)
		if err != nil {
			log.Error("[ReturnMessage] unmarshal monitorType error ", err)
			return nil, err
		}
		monitorDatas = append(monitorDatas, monitorType)
	}
	return &monitorDatas, nil
}

func ReturnData(typ, str string) (*interface{}, error) {
	monitorType, ok := NewOfType(typ)
	if !ok {
		return nil, errors.New(typ + " is not support type")
	}
	err := json.Unmarshal([]byte(str), &monitorType)
	if err != nil {
		log.Error("[ReturnData] unmarshal monitorType error ", err)
		return nil, err
	}
	return &monitorType, nil
}

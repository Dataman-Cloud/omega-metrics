package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Dataman-Cloud/omega-metrics/cache"
	"github.com/Dataman-Cloud/omega-metrics/config"
	"github.com/Dataman-Cloud/omega-metrics/util"
	log "github.com/cihub/seelog"
	"io/ioutil"
	"net/http"
)

func AutoScale(token string) error {
	log.Debug("into AutoScale")
	conf := config.Pairs()
	applications, err := util.GetAllApps()
	if err != nil {
		log.Error(err)
	}
	for _, app := range applications {
		appMonitor, err := gatherApp(app)
		if err != nil {
			log.Error(err)
		}
		log.Debug("appCpuUsed: ", appMonitor.AppCpuUsed)
		log.Debug("appCpuShare: ", appMonitor.AppCpuShare)
		cpuUsedPercent := appMonitor.AppCpuUsed / appMonitor.AppCpuShare
		memUsedPercent := float64(appMonitor.AppMemUsed) / float64(appMonitor.AppMemShare)
		log.Debug("cpuUsedPercent:  ", cpuUsedPercent)
		log.Debug("menUsedPercent:  ", memUsedPercent)
		log.Debug("MaxMemPercent:  ", conf.MaxMemPercent)
		log.Debug("MinCpuPercent:  ", conf.MinCpuPercent)
		if appMonitor.AppCpuShare != 0 && appMonitor.AppMemShare != 0 {
			if cpuUsedPercent > conf.MaxCpuPercent || memUsedPercent > conf.MaxMemPercent {
				// 调用扩容接口
				err := AppRest(token, app.Instances+1, fmt.Sprintf("%d", app.Id))
				if err != nil {
					log.Error(err)
				}
			}
			if app.Instances > 1 {
				if cpuUsedPercent < conf.MinCpuPercent && memUsedPercent < conf.MinMemPercent {
					// 调用扩接口
					err := AppRest(token, app.Instances-1, fmt.Sprintf("%d", app.Id))
					if err != nil {
						log.Error(err)
					}
				}
			}
		}
	}
	return nil
}

type UpdateContainerForm struct {
	AppId              string `json:"appId"`
	UpdateContainerNum int    `json:"updateContainerNum"`
}

type Resp struct {
	Code  int64
	Data  string
	Error error
}

// 调用app接口
func AppRest(token string, containerNum int, appId string) error {
	conf := config.Pairs()
	client := &http.Client{}
	addr := fmt.Sprintf("%s:%d/api/v1/applications/update-container-num", conf.Omega_app_host, conf.Omega_app_port)
	conn := cache.Open()
	defer conn.Close()

	from := UpdateContainerForm{
		AppId:              appId,
		UpdateContainerNum: containerNum,
	}
	b, err := json.Marshal(from)
	if err != nil {
		log.Error(err)
		return err
	}
	body := bytes.NewBuffer([]byte(b))
	req, err := http.NewRequest("POST", addr, body)
	if err != nil {
		log.Error(err)
		return err
	}
	req.Header.Set("Content-Type", "applicaton/json")
	req.Header.Set("Authorization", token)
	resp, err := client.Do(req)
	if err != nil {
		log.Error(err)
		return err
	}
	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return err
	}
	log.Debug(string(respData))

	var respInfo Resp
	err = json.Unmarshal(respData, &respInfo)
	if err != nil {
		log.Error(err)
	}

	if respInfo.Code != 0 {
		return respInfo.Error
	}
	return nil
}

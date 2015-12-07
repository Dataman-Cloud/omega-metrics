package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Dataman-Cloud/omega-metrics/cache"
	"github.com/Dataman-Cloud/omega-metrics/config"
	"github.com/Dataman-Cloud/omega-metrics/util"
	log "github.com/cihub/seelog"
	redis "github.com/garyburd/redigo/redis"
	"io/ioutil"
	"net/http"
)

var checkUpTimes = make(map[int64]int)
var checkDownTimes = make(map[int64]int)
var overMaxTimes = make(map[int64]int)
var overMinTimes = make(map[int64]int)

func AutoScale(token string) error {
	conf := config.Pairs()
	conn := cache.Open()
	defer conn.Close()

	uid, err := redis.String(conn.Do("HGET", "s:"+token, "user_id"))
	if err != nil {
		log.Error(err)
		return err
	}
	log.Debug("uid: ", uid)

	applications, err := util.GetAllApps(uid)
	if err != nil {
		log.Error(err)
		return err
	}

	for _, app := range applications {
		appMonitor, err := gatherApp(app)
		if err != nil {
			log.Error(err)
			return err
		}

		log.Debug("appCpuUsed: ", appMonitor.AppCpuUsed)
		log.Debug("appCpuShare: ", appMonitor.AppCpuShare)
		cpuUsedPercent := appMonitor.AppCpuUsed / appMonitor.AppCpuShare
		memUsedPercent := float64(appMonitor.AppMemUsed) / float64(appMonitor.AppMemShare)
		log.Debug("cpuUsedPercent:  ", cpuUsedPercent)
		log.Debug("menUsedPercent:  ", memUsedPercent)
		log.Debug("MaxMemPercent:  ", conf.Scale.MaxMemPercent)
		log.Debug("MinCpuPercent:  ", conf.Scale.MinCpuPercent)
		log.Debug("MaxInstances: ", conf.Scale.MaxInstances)

		if appMonitor.AppCpuShare != 0 && appMonitor.AppMemShare != 0 {
			if app.Instances <= conf.Scale.MaxInstances/2 {
				checkUpTimes[app.Id]++
				log.Debug("checkUptimes ===", app.Id, checkUpTimes[app.Id])
				if cpuUsedPercent > conf.Scale.MaxCpuPercent || memUsedPercent > conf.Scale.MaxMemPercent {
					overMaxTimes[app.Id]++
					log.Debug("overMaxTimes ===", app.Id, overMaxTimes[app.Id])
					// 调用扩容接口
					if checkUpTimes[app.Id] == overMaxTimes[app.Id] && overMaxTimes[app.Id] > conf.Scale.WaitCheckTimes {
						ClearMap(overMaxTimes)
						ClearMap(checkUpTimes)
						log.Debug("调用扩容接口")
						err := AppRest(token, app.Instances*2, fmt.Sprintf("%d", app.Id))
						if err != nil {
							log.Error("扩容失败：", err)
						}
					}
				} else {
					overMaxTimes[app.Id] = 0
					checkUpTimes[app.Id] = 0
					log.Debug("set ", app.Id, "to 0")
				}
			}

			if app.Instances >= conf.Scale.MinInstances {
				checkDownTimes[app.Id]++
				if cpuUsedPercent < conf.Scale.MinCpuPercent && memUsedPercent < conf.Scale.MinMemPercent {
					overMinTimes[app.Id]++
					// 调用扩接口
					if checkDownTimes[app.Id] == overMinTimes[app.Id] && overMinTimes[app.Id] > conf.Scale.WaitCheckTimes {
						ClearMap(overMinTimes)
						ClearMap(checkDownTimes)
						err := AppRest(token, app.Instances/2, fmt.Sprintf("%d", app.Id))
						if err != nil {
							log.Error(err)
						}
					}
				} else {
					ClearMap(overMinTimes)
					ClearMap(checkDownTimes)
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

func ClearMap(m map[int64]int) {
	for k, _ := range m {
		delete(m, k)
	}
}

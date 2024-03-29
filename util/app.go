package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Dataman-Cloud/omega-metrics/config"
	"io/ioutil"
	"net/http"
	"time"
)

var (
	AppStatusUrl  string
	AppServerHost string
	AppServerPort int

	HeaderToken        = "Authorization"
	DefaultHttpTimeout = 15 * time.Second
)

func InitApp() {
	conf := config.Pairs()
	AppServerHost = conf.Omega_app_host
	AppServerPort = conf.Omega_app_port
	AppStatusUrl = fmt.Sprintf("%s:%d/api/v3/apps/status", AppServerHost, AppServerPort)
}

// query all apps by user id with token
func QueryApps(token, clusterId string) ([]AppConfig, error) {
	appListUrl := fmt.Sprintf("%s:%d/api/v3/clusters/%s/apps", AppServerHost, AppServerPort, clusterId)
	response, err := doHttpRequest(appListUrl, token)
	if err != nil {
		return nil, err
	}

	var appListResp ClusterAppListResp
	if err := json.Unmarshal(response, &appListResp); err != nil {
		return nil, err
	}

	if appListResp.Code != 0 {
		return nil, fmt.Errorf("[App list] Get app list failed code is %d", appListResp.Code)
	}

	return appListResp.Data.App, nil
}

// query all app status under one user all clusters
func QueryAppStatus(token string, clusterId string) (map[string]AppStatus, error) {
	url := fmt.Sprintf("%s?cid=%s", AppStatusUrl, clusterId)
	response, err := doHttpRequest(url, token)
	if err != nil {
		return nil, err
	}

	var appStatusResp AppStatusResp
	if err := json.Unmarshal(response, &appStatusResp); err != nil {
		return nil, err
	}

	if appStatusResp.Code != 0 {
		return nil, fmt.Errorf("[App status] Get app status failed code: %d", appStatusResp.Code)
	}
	return appStatusResp.Data, nil
}

// do a http request bu url addr return response body
func doHttpRequest(url, token string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add(HeaderToken, token)
	client := http.Client{Timeout: DefaultHttpTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp == nil {
		return nil, errors.New("[App list] App list response is nil")
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

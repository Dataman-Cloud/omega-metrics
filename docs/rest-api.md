#omega-metrics REST API

## API List
  - [GET http://localhost:9005/](#healthCheck)  :healthCheck  检查服务是否正常运行
  - [GET http://localhost:9005/api/v1/metrics/cluster/{clusterID}](#get master mertrics) 获取ID 为`clusterID`的集群的集群资源消耗信息
  - [GET http://localhost:9005/api/v1/event/{clusterID}/{appName}](#get marathon event) 获取`clusterID`的集群上`appName`应用的部署信息

#### GET `http://localhost:9005/`
检查服务是否正常运行 (healthCheck)   </br>
***Response***
```
pass
```

#### GET `http://localhost:9005/api/v1/metrics/cluster/{clusterID}`
根据集群ID 获取集群监控的信息

e.g :
```shell
curl -X GET -H {Authorization:token} http://localhost:9005/api/v1/metrics/cluster/140
```

***Response***
```json

{
    "code":0,
    "data":{
        "masMetrics":{
            "cpuPercent":20,
            "cpuShare":0.4,
            "cpuTotal":2,
            "memTotal":2928,
            "memUsed":64,
            "diskTotal":95545,
            "diskUsed":0,
            "timestamp":1446626558,
            "leader":1,
            "clusterId":"1"
        },
        "appMetrics":[
            {
                "appName":"hope",
                "appCpuShare":0.19921875,
                "appCpuUsed":0,
                "appMemShare":64,
                "appMemUsed":6,
                "instances":2
            },
            {
                "appName":"2048",
                "appCpuShare":0.099609375,
                "appCpuUsed":0,
                "appMemShare":32,
                "appMemUsed":3,
                "instances":2
            }
        ]
    },
    "error":""
}
```

|Key          |Remarks                           |
|-------------|----------------------------------|
|cpuPercent   |CPU占用百分比                     |
|cpuShare     |CPU占用量                         |
|cpuTotal     |cpu总量                           |
|memTotal     |内存总量                          |
|memUsed      |内存使用量                        |
|diskTotal    |磁盘总量                          |
|diskUsed     |磁盘使用量                        |
|timestamp    |时间戳  响应生成时间              |
|clusterId    |集群Id                            |
|appName      |应用名称                          |
|appCpuShare  |应用申请CPU占用量                 |
|appCpuUsed   |应用实际使用CPU量                 |
|appMemShare  |应用申请内存使用量                |
|appMemUsed   |应用实际内存使用量                |
|instances    |应用的实例个数                    |
|error        |响应失败原因 若code为0 则 error为"" |
|code         |响应状态 0 代表成功 非0 为失败       |

说明:
* masMetrics数据取自mesos-master metrics 磁盘 内存 和 CPU 数据均和直接查看主机数据不同
* appMetrics数据取自cadvisor

#### GET `http://localhost:9005/api/v1/event/{clusterID}/{appName}`
根据集群ID和appName, 获取事件的监控信息

e.g :
```shell
curl  http://localhost:9005/api/v1/event/140/testapp
```

***Response***
```json
{
    "code":"0",
    "data":[
        "2015-10-12T02:08:59.525Z /dataman-nginx-test2 ScaleApplication deployment_step_success",
        "2015-10-12T02:08:39.514Z /dataman-nginx-test2 StartApplication deployment_step_success",
        "2015-10-12T02:07:59.677Z /dataman-nginx-test2 StopApplication deployment_step_success"
    ],
    "error":""
}
```

#### GET `http://localhost:9005/api/v1/appmetrics/clusterId/{clusterId}/app/{app}`
根据集群ID和appName，获取app的应用数据监控

e.g :
```shell
curl  http://localhost:9005/api/v1/appmetrics/clusterId/140/app/test
```

***Response***
```json
{

    "code": 0,
    "data": [
        {
            "clusterId": "1",
            "app": {
                "appName": "test222",
                "appId": "10.3.10.83:31903"
            },
            "containerId": "mesos-20151030-033030-1393165066-5050-1-S0.2f8b7126-1391-4d72-8477-8c92a2a83677",
            "cpuUsedCores": 0.00025954916658571005,
            "cpuShareCores": 0.19921875,
            "memoryTotal": 128,
            "memoryUsed": 13
        },
        {
            "clusterId": "1",
            "app": {
                "appName": "test222",
                "appId": "10.3.10.83:31710"
            },
            "containerId": "mesos-20151030-033030-1393165066-5050-1-S0.c37109b5-4058-4da3-a3fd-6785a3fa32a6",
            "cpuUsedCores": 0.0002526872589691358,
            "cpuShareCores": 0.19921875,
            "memoryTotal": 128,
            "memoryUsed": 2
        }
    ],
    "error": ""

}
```


|Key          |Remarks                           |
|-------------|----------------------------------|
|clusterId    |实例所属的集群ID                  |
|appName      |实例的应用名称                    |
|appId        |实例ID                            |
|containerId  |实例所在容器的ID                  |
|cpuUsedCores |实例实际使用CPU量                 |
|cpuShareCores|实例申请的CPU量                   |
|memoryTotal  |实例申请的内存量                  |
|memoryUsed   |实例实际使用的内存量              |

说明:
* 数据取自cadvisor

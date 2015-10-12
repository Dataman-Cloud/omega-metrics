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
curl  http://localhost:9005/api/v1/metrics/cluster/140
```

***Response***
```json
{
    "code":"0",
    "data":[
        {
            "cpuPercent":0,
            "memTotal":2928,
            "memUsed":0,
            "diskTotal":4975,
            "diskUsed":0,
            "timestamp":1443579381
        },
        {
            "cpuPercent":0,
            "memTotal":2928,
            "memUsed":0,
            "diskTotal":4975,
            "diskUsed":0,
            "timestamp":1443579361
        },
        {
            "cpuPercent":0,
            "memTotal":2928,
            "memUsed":0,
            "diskTotal":4975,
            "diskUsed":0,
            "timestamp":1443579341
        }
    ],
    "error":""
}
```

|Key          |Remarks                           |
|-------------|----------------------------------|
|cpuPercent   |CPU 使用百分比                     |
|memTotal     |内存总量                           |
|memUsed      |内存使用量                         |
|diskTotal    |磁盘总量                           |
|diskUsed     |磁盘使用量                         |
|timestamp    |时间戳  响应生成时间                |
|error        |响应失败原因 若code为0 则 error为"" |
|code         |响应状态 0 代表成功 非0 为失败       |

说明:
* 数据取自mesos-master metrics 磁盘 内存 和 CPU 数据均和直接查看主机数据不同

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

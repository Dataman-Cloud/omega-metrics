#omega-app REST API

* [Summary](#summary)

* [Application](#application)
  - [GET /api/v1/applications/list](#get-apiv1applicationslist): List Applications by user token(请求应用列表)
  - [GET /api/v1/applications/{appId}](#get-apiv1applicationsappid): Get the application infomation with id `appId` (请求 Id 为 appId 应用的详细信息)
  - [POST /api/v1/applications/deploy](#post-apiv1applicationsdeploy): Deloy An Application (创建应用)
  - [GET /api/v1/applications/{appId}/actions](#get-apiv1applicationsappidactions): List the actions of the application with id `appId`.(请求 Id 为 appId 的应用事件)
  - [GET /api/v1/applications/{appId}/instances](#get-apiv1applicationsappidinstances): List the instances of the application with id `appId`.(请求Id 为 appId 应用实例列表)
  - [POST /api/v1/applications/{appId}/update-container-num](#post-apiv1applicationsappidupdate-container-num): Update the Container Number of the application with id `appId`(修改 Id 为 appId 应用配置中的容器个数)
  - [GET /api/v1/applications/{appId}/config](#get-apiv1applicationsappidconfig):  Get the config of the application with id `appId`(请求应用配置信息)
  - [POST /api/v1/applications/{appId}/rollback](#post-apiv1applicationsappidrollback): Rollback the application with id `appId`(应用实例回滚)
  - [GET /api/v1/applications/{appId}/delete](#get-apiv1applicationsappiddelete): Delete the application with id `appId` .(删除 Id 为 appId 的应用)
  - [GET /api/v1/applications/{appId}/stop](#get-apiv1applicationsappidstop): Stop the application with id `appId` .(停止 Id 为 appId 的应用)

###summary

Response Format:

```json
{
    "code": 0,
    "data":  "None",
    "errors": "None",
}
```
- **code**: required, 0 = success handled, while >0 = error occured, 1 = invalid, 2 = permission denied, ..
- **data**: a dictionary or list of successful data or None, required
- **errors**: a dictionary or string of error[s], form errors for example, required


###Application
#### GET `/api/v1/applications/list`
List application by user token (请求应用列表)
**Request**

|key          |remarks   |
|-------------|----------|
|authorization|user token|

e.g: 
```shell
curl -X GET --header "authorization:usertoken" 123.59.58.58:8080/api/v1/applications/list | python -m json.tool
```
**Response:**
```json
[
    {
        "appId": "3",
        "appName": "Test",
        "appStatus": "\u90e8\u7f72\u4e2d",
        "clusterId": "1",
        "update": "2015-08-12T14:15:20.536+08:00"
    },
    {
        "appId": "1",
        "appName": "Test",
        "appStatus": "\u90e8\u7f72\u4e2d",
        "clusterId": "1",
        "update": "2015-08-11T20:01:05.727+08:00"
    }
]

```

|key        |remarks                     |
|-----------|----------------------------|
|appId      |应用ID,查看App详情时需要传入|
|appName    |应用名称                    |
|appStatus  |应用状态                    |
|update     |最后更新时间                |
|clusterId  |集群uuid                    |

#### GET `/api/v1/applications/{appId}`

Get the application infomation with the id `appId`(请求 Id 为 appId 应用的详细信息)

e.g:
```shell
curl -X GET --header "authorization:usertoken" localhost:8080/api/v1/applications/1 | python -m json.tool
```

**Response**
```json
{
    "code": 0,
    "data": {
        "appId": "1",
        "clusterId": "1",
        "latestUpdated": "2015-08-18T13:18:31.434Z",
        "latestVersion": "1439875110959",
        "name": "hello world",
        "status": "\u90e8\u7f72\u4e2d"
    },
    "error": ""
}

```

|Key          |Remarks                           |
|-------------|----------------------------------|
|appId        |                                  |
|clusterId    |                                  |
|latestUpdated|最新版本号                        |
|latestVersion|最近更新时间                      |
|name         |应用名称                          |
|status       |应用状态，需与前端提前定义好状态码|

**STATUS**

|Status  |Code|
|--------|----|
|部署失败|1   |
|部署中  |2   |
|运行中  |3   |
|已停止  |4   |
|删除    |5   |

#### POST `/api/v1/applications/deploy`
Deloy an application (创建应用)

e.g:
```shell
curl -X POST http://123.59.58.58:8080/api/v1/applications/deploy \
        -H Authorization:usertoken \
        -H Content-Type:application/json -d '{
           "appName": "app1",
           "clusterId": "2",
           "containerNum": "2",
           "containerPortsInfo": {"inner":[8080,8000],"outer":[8001,8000]},
           "containerSize": "2",
           "envs": {"java": "1.9", "c++": "1.6"},
           "imageURI": "www.baidu.com",
           "imageversion": "1.0.1"
        }'

```

|key          |remarks       |
|-------------|--------------|
|appName    |应用名称        |
|clusterId      |集群id      |
|containerNum     |容器个数|
|containerPortsInfo     |容器端口(可选项 .用户不填就没有)| 
|containerSize     |Cpu 内存大小(例如:"1" -->1 CPU 512MB 内存)|
|envs     |环境变量(可选项,Key:Value.用户不填就没有)| 
|imageURI     |   镜像地址      |
|imageversion     |镜像版本  |

|containerSize  | show |
|--------|------------|
| 1 | 1 CPU  512MB 内存|
| 2 | 1 CPU  1G 内存  |
| 3 | 1 CPU  2G 内存  |
| 4 | 2 CPU  2G 内存  |

#### GET `/api/v1/applications/{appId}/actions`
List the actions of the application with id `appId` (请求 Id 为 appId 的应用事件)
e.g:
```shell
curl -X GET --header "authorization:usertoken" 123.59.58.58:8080/api/v1/applications/1/actions | python -m json.tool
```

**Response**
```json
{
    "code": 0,
    "data":
        [{
            "action": "\u521b\u5efa\u5e94\u7528Test,\u7248\u672c[1439294464943]",
            "actionType": "Aplication",
            "id": 2,
            "time": "2015-08-11T20:01:05.868+08:00"
        }],
    "errors": ""
}
```

|key       |remarks     |
|----------|------------|
|id        |事件ID      |
|time      |事件发生时间|
|action    |具体事件    |
|actionType|事件类型    |

#### GET `/api/v1/applications/{appId}/instances`
List the instances of the application with id `appId`.(请求Id 为 appId 应用实例列表)

e.g:
```shell
curl 123.59.58.58:8080/api/v1/applications/1/instances | python -m json.tool
```

**Response**
```json
{
    "code": 0,
    "data":
        [{
            "configId": "1",
            "configInstance": "1439294464943",
            "lastUpdate": "2015-08-11T20:01:04.944+08:00"
        }],
    "errors": ""
}
```
|key          |remarks       |
|-------------|--------------|
|configId     |配置文件ID    |
|configInstance|配置文件实例号|
|lastUpdate   |更新时间      |

#### POST `/api/v1/applications/{appId}/update-container-num`

Update the Container Number of the application with id `appId`(修改 Id 为 appId 应用配置中的容器个数)

```shell
curl -X POST http://123.59.58.58:8080/api/v1/applications/1/update-container-num \          
        -H Authorization:usertoken \                                            
        -H Content-Type:application/json -d '{                                  
            "updateContainerNum": 1
        }' 
```

#### GET `/api/v1/applications/{appId}/config`

get the config of the application with id `appId` (请求应用配置信息)

e.g: 
```shell
curl -X GET --header "authorization:usertoken" localhost:8080/api/v1/applications/1/config | python -m json.tool
```
**Response**
```json
{
    "code": 0,
    "data":
        {
          "containerNum" : "2",
           "containerPortsInfo" : {"inner":[8888,8001], "outer":[8000,8080]},
           "containerSize" : "1",
           "envs" : {"java": "1.9", "c++": "1.6"},
           "imageURI" : "www.baidu.com",
           "imageversion" : "1.0.1"
        },
    "errors": ""
}
```
|key          |remarks       |
|-------------|--------------|
|containerNum     |容器个数|
|containerPortsInfo     |容器端口(可选项,用户不填就没有)| 
|containerSize     |Cpu 内存大小(例如:"1" -->1 CPU 512MB 内存)|
|envs     |环境变量(可选项,用户不填就没有)| 
|imageURI     |镜像地址|
|imageversion     |镜像版本|

#### POST `/api/v1/applications/{appId}/rollback`

rollback the applications with id `appId` (应用实例回滚)

```shell
curl -X POST http://123.59.58.58:8080/api/v1/applications/1/rollback \          
        -H Authorization:usertoken \                                            
        -H Content-Type:application/json -d '{                                  
            "clusterId": "1234",                                                
            "appId": 1,
            "appName": "test",                                                  
            "appConfigId": 1                                                    
        }' 
```
#### GET `/api/v1/applications/{appId}/delete`
Delete the application with id `appId`.(删除 Id 为 appId 的应用)

e.g:
```shell
curl 123.59.58.58:8080/api/v1/applications/1/delete | python -m json.tool
```

**Response**
```json
{
    "code": 0,
    "data": {
        "deletState": 0
    },
    "errors": ""
}
```
|deletState   |show |
|-------------|-----|
| 0  | 删除成功     |
| -1 | 删除失败     |

#### GET `/api/v1/applications/{appId}/stop`
Stop the application with id `appId`.(停止 Id 为 appId 的应用)

e.g:
```shell
curl 123.59.58.58:8080/api/v1/applications/1/stop | python -m json.tool
```


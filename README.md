# OS Remote Executor Api

## Help
```text
usage: remote_executor-amd64 [<flags>] <command> [<args> ...]


Flags:
  -h, --[no-]help         Show context-sensitive help (also try --help-long and --help-man).
      --addr=":2376"      host:port for execution.
      --[no-]debug        Enable debug messages
      --root="$TEMP/remote_executor-amd64"  
                          Working root directory
      --key_expire=48h    Set the database key expire time. Example: "key_expire=1h"
      --exec_timeout=24h  Set the exec command expire time. Example: "exec_timeout=30m"
      --timeout=30s       Timeout for calling endpoints on the engine
      --max-requests=0    Maximum number of concurrent requests. 0 to disable.
      --pool_size=30      Set the size of the execution work pool.
      --[no-]version      Show application version.

Commands:
help [<command>...]
    Show help.

run
    Run server
```

## Router
```text
[GIN-debug] GET    /debug/pprof/             --> github.com/gin-gonic/gin.WrapF.func1 (5 handlers)
[GIN-debug] GET    /debug/pprof/cmdline      --> github.com/gin-gonic/gin.WrapF.func1 (5 handlers)
[GIN-debug] GET    /debug/pprof/profile      --> github.com/gin-gonic/gin.WrapF.func1 (5 handlers)
[GIN-debug] POST   /debug/pprof/symbol       --> github.com/gin-gonic/gin.WrapF.func1 (5 handlers)
[GIN-debug] GET    /debug/pprof/symbol       --> github.com/gin-gonic/gin.WrapF.func1 (5 handlers)
[GIN-debug] GET    /debug/pprof/trace        --> github.com/gin-gonic/gin.WrapF.func1 (5 handlers)
[GIN-debug] GET    /debug/pprof/allocs       --> github.com/gin-gonic/gin.WrapH.func1 (5 handlers)
[GIN-debug] GET    /debug/pprof/block        --> github.com/gin-gonic/gin.WrapH.func1 (5 handlers)
[GIN-debug] GET    /debug/pprof/goroutine    --> github.com/gin-gonic/gin.WrapH.func1 (5 handlers)
[GIN-debug] GET    /debug/pprof/heap         --> github.com/gin-gonic/gin.WrapH.func1 (5 handlers)
[GIN-debug] GET    /debug/pprof/mutex        --> github.com/gin-gonic/gin.WrapH.func1 (5 handlers)
[GIN-debug] GET    /debug/pprof/threadcreate --> github.com/gin-gonic/gin.WrapH.func1 (5 handlers)
[GIN-debug] GET    /swagger/*any             --> github.com/swaggo/gin-swagger.CustomWrapHandler.func1 (5 handlers)
[GIN-debug] GET    /version                  --> github.com/xmapst/osreapi/internal/handlers.version (5 handlers)
[GIN-debug] GET    /healthyz                 --> github.com/xmapst/osreapi/internal/handlers.healthyz (5 handlers)
[GIN-debug] GET    /metrics                  --> github.com/xmapst/osreapi/internal/handlers.metrics (5 handlers)
[GIN-debug] GET    /heartbeat                --> github.com/xmapst/osreapi/internal/handlers.heartbeat (5 handlers)
[GIN-debug] HEAD   /heartbeat                --> github.com/xmapst/osreapi/internal/handlers.heartbeat (5 handlers)
[GIN-debug] GET    /api/v1/task              --> github.com/xmapst/osreapi/internal/handlers/api/v1/task.List (6 handlers)
[GIN-debug] POST   /api/v1/task              --> github.com/xmapst/osreapi/internal/handlers/api/v1/task.Post (6 handlers)
[GIN-debug] GET    /api/v1/task/:task        --> github.com/xmapst/osreapi/internal/handlers/api/v1/task.Detail (6 handlers)
[GIN-debug] PUT    /api/v1/task/:task        --> github.com/xmapst/osreapi/internal/handlers/api/v1/task.Stop (6 handlers)
[GIN-debug] PUT    /api/v1/task/:task/:step  --> github.com/xmapst/osreapi/internal/handlers/api/v1/task.StopStep (6 handlers)
[GIN-debug] GET    /api/v1/task/:task/:step/console --> github.com/xmapst/osreapi/internal/handlers/api/v1/task.StepDetail (6 handlers)
[GIN-debug] GET    /api/v1/pool              --> github.com/xmapst/osreapi/internal/handlers/api/v1/pool.Detail (6 handlers)
[GIN-debug] POST   /api/v1/pool              --> github.com/xmapst/osreapi/internal/handlers/api/v1/pool.Post (6 handlers)
[GIN-debug] GET    /api/v1/state             --> github.com/xmapst/osreapi/internal/handlers/api/v1/status.Detail (6 handlers)
```

## 特性

- [x] 支持Windows/Linux/Mac
- [x] 动态调整工人数量
- [x] 基于有向无环(DAG)编排执行
- [x] 支持任务/步骤的强制终止
- [x] Workspace隔离

## 服用方式
以windows服务形式部署运行
### 用管理模式打开powershell执行
```powershell
New-Service -Name remote_executor -BinaryPathName "C:\remote_executor.exe run --addr=:2376" -DisplayName  "Remote Executor " -StartupType Automatic
sc.exe failure remote_executor reset= 0 actions= restart/0/restart/0/restart/0
sc.exe start remote_executor
```

## 本地编译
```shell
git clone https://github.com/xmapst/osreapi.git
cd osreapi
make
```

## 测试
### 获取当前所有任务
```shell
# 默认按开始时间排序
curl -XGET http://localhost:2376/api/v1/task
# 按完成时间排序
curl -XGET http://localhost:2376/api/v1/task?sort=et
# 按过期时间排序
curl -XGET http://localhost:2376/api/v1/task?sort=rt
```
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "running": 1,
    "tasks": [
      {
        "id": "ad2fc4c8-05fc-4aab-9f28-5ceea65d0982",
        "url": "http://localhost:2376/api/v1/task/ad2fc4c8-05fc-4aab-9f28-5ceea65d0982",
        "state": 0,
        "code": 510,
        "count": 4,
        "msg": "execution failed: [Step: 0, Name: step0; Step: 2, Name: step2]",
        "times": {
          "st": "2022-11-17T17:17:58+08:00",
          "et": "2022-11-17T17:17:58+08:00",
          "rt": "47h53m30.619752642s"
        }
      },
      {
        "id": "ed95e570-ce58-4f97-a5ae-abd7694b8dbc",
        "url": "http://localhost:2376/api/v1/task/ed95e570-ce58-4f97-a5ae-abd7694b8dbc",
        "state": 0,
        "code": 0,
        "count": 4,
        "msg": "all steps executed successfully",
        "times": {
          "st": "2022-11-17T17:20:29+08:00",
          "et": "2022-11-17T17:21:30+08:00",
          "rt": "47h57m2.619748142s"
        }
      },
      {
        "id": "aece7b89-3081-4294-ac78-0a9b0987a493",
        "url": "http://localhost:2376/api/v1/task/aece7b89-3081-4294-ac78-0a9b0987a493",
        "state": 1,
        "code": 0,
        "count": 4,
        "msg": "currently executing: [Step: 0, Name: step0; Step: 2, Name: step2]",
        "times": {
          "st": "2022-11-17T17:24:25+08:00",
          "rt": "47h59m57.619748942s"
        }
      }
    ],
    "total": 3
  }
}
```
### 执行脚本或命令
顺序执行:
```shell
curl -XPOST http://localhost:2376/api/v1/task -d '[
  {
    "command_type":"cmd",
    "command_content":"curl -I https://%envhost%",
    "env_vars":[
      "envhost=www.q1.com",
      "env2=value2"
    ]
  },
  {
    "command_type":"powershell",
    "command_content":"sleep 30",
    "env_vars":[
      "env1=value1",
      "env2=value2"
    ]
  },
  {
    "command_type":"cmd",
    "command_content":"curl -I https://%envhost%",
    "env_vars":[
      "envhost=baidu.com"
    ]
  }
]'
```
并行执行
```shell
curl -XPOST http://localhost:2376/api/v1/task?ansync=true -d '[
  {
    "command_type":"cmd",
    "command_content":"curl -I https://%envhost%",
    "env_vars":[
      "envhost=www.q1.com",
      "env2=value2"
    ]
  },
  {
    "command_type":"powershell",
    "command_content":"sleep 30",
    "env_vars":[
      "env1=value1",
      "env2=value2"
    ]
  },
  {
    "command_type":"cmd",
    "command_content":"curl -I https://%envhost%",
    "env_vars":[
      "envhost=baidu.com"
    ]
  }
]'
```

自定义编排执行
```shell
curl -XPOST http://localhost:2376/api/v1/task?ansync=true -d '[
  {
    "name": "step0",
    "command_type": "cmd",
    "command_content": "ping baidu.com",
    "env_vars": [
      "env1=a",
      "env2=b",
      "env3=c"
    ]
  },
  {
    "name": "step1",
    "command_type": "cmd",
    "command_content": "curl https://www.baidu.com",
    "env_vars": [
      "env1=a",
      "env2=b",
      "env3=c"
    ],
    "depets_on": [
      "step0"
    ]
  },
  {
    "name": "step2",
    "command_type": "cmd",
    "command_content": "set",
    "env_vars": [
      "env1=a",
      "env2=b",
      "env3=c"
    ],
    "depets_on": [
      "step0"
    ]
  }
]'
```

返回内容:  
成功:
```json
{
  "code": 0,
  "message": "success",
  "data":{
    "count": 4,
    "id": "7f478334-1f44-4580-8aa9-9772b84e4bf6",
    "url": "http://localhost:2376/api/v1/task/7f478334-1f44-4580-8aa9-9772b84e4bf6",
    "timestamp": 1668649373430555500
  }
}
```
参数不全:
```json
{
  "code": 400,
  "message": "[0]: Key: 'PostStruct.CommandType' Error:Field validation for 'CommandType' failed on the 'required' tag"
}
```

### 查询结果  
请求:  
```shell
curl http://localhost:2376/api/v1/task/a55c705f-d279-48cd-b1cb-803a345f8cd9
```

返回内容:  
成功:  
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "step": 0,
      "url": "http://localhost:2376/api/v1/task/a55c705f-d279-48cd-b1cb-803a345f8cd9/0/console",
      "name": "5a8fbc27-8b94-487a-adc4-174d76a16275-0",
      "state": 0,
      "code": 0,
      "msg": "execution succeed",
      "times": {
        "st": "2022-11-18T16:20:53+08:00",
        "et": "2022-11-18T16:20:56+08:00",
        "rt": "47h59m20.842028157s"
      }
    },
    {
      "step": 1,
      "url": "http://localhost:2376/api/v1/task/a55c705f-d279-48cd-b1cb-803a345f8cd9/1/console",
      "name": "5a8fbc27-8b94-487a-adc4-174d76a16275-1",
      "state": 0,
      "code": 0,
      "msg": "execution succeed",
      "depets_on": [
        "5a8fbc27-8b94-487a-adc4-174d76a16275-0"
      ],
      "times": {
        "st": "2022-11-18T16:20:56+08:00",
        "et": "2022-11-18T16:20:56+08:00",
        "rt": "47h59m20.842010857s"
      }
    },
    {
      "step": 2,
      "url": "http://localhost:2376/api/v1/task/a55c705f-d279-48cd-b1cb-803a345f8cd9/2/console",
      "name": "5a8fbc27-8b94-487a-adc4-174d76a16275-2",
      "state": 0,
      "code": 0,
      "msg": "execution succeed",
      "depets_on": [
        "5a8fbc27-8b94-487a-adc4-174d76a16275-1"
      ],
      "times": {
        "st": "2022-11-18T16:20:56+08:00",
        "et": "2022-11-18T16:21:26+08:00",
        "rt": "47h59m50.842009757s"
      }
    },
    {
      "step": 3,
      "url": "http://localhost:2376/api/v1/task/a55c705f-d279-48cd-b1cb-803a345f8cd9/3/console",
      "name": "5a8fbc27-8b94-487a-adc4-174d76a16275-3",
      "state": 0,
      "code": 0,
      "msg": "execution succeed",
      "depets_on": [
        "5a8fbc27-8b94-487a-adc4-174d76a16275-2"
      ],
      "times": {
        "st": "2022-11-18T16:21:26+08:00",
        "et": "2022-11-18T16:21:26+08:00",
        "rt": "47h59m50.842002157s"
      }
    }
  ]
}
```

执行中:
```json
{
  "code": 1001,
  "message": "in progress: [Step: 0, Name: a446f074-d37f-48f9-9fdb-f49211847b9e-0]",
  "data": [
    {
      "step": 0,
      "url": "http://localhost:2376/api/v1/task/a446f074-d37f-48f9-9fdb-f49211847b9e/0/console",
      "name": "a446f074-d37f-48f9-9fdb-f49211847b9e-0",
      "state": 1,
      "code": 0,
      "msg": "The step is running",
      "times": {
        "st": "2022-11-18T16:15:56+08:00",
        "rt": "47h59m57.433539558s"
      }
    },
    {
      "step": 1,
      "url": "http://localhost:2376/api/v1/task/a446f074-d37f-48f9-9fdb-f49211847b9e/1/console",
      "name": "a446f074-d37f-48f9-9fdb-f49211847b9e-1",
      "state": 2,
      "code": 0,
      "msg": "The current step only proceeds if the previous step succeeds.",
      "depets_on": [
        "a446f074-d37f-48f9-9fdb-f49211847b9e-0"
      ],
      "times": {
        "rt": "47h59m57.433535058s"
      }
    },
    {
      "step": 2,
      "url": "http://localhost:2376/api/v1/task/a446f074-d37f-48f9-9fdb-f49211847b9e/2/console",
      "name": "a446f074-d37f-48f9-9fdb-f49211847b9e-2",
      "state": 2,
      "code": 0,
      "msg": "The current step only proceeds if the previous step succeeds.",
      "depets_on": [
        "a446f074-d37f-48f9-9fdb-f49211847b9e-1"
      ],
      "times": {
        "rt": "47h59m57.433533458s"
      }
    },
    {
      "step": 3,
      "url": "http://localhost:2376/api/v1/task/a446f074-d37f-48f9-9fdb-f49211847b9e/3/console",
      "name": "a446f074-d37f-48f9-9fdb-f49211847b9e-3",
      "state": 2,
      "code": 0,
      "msg": "The current step only proceeds if the previous step succeeds.",
      "depets_on": [
        "a446f074-d37f-48f9-9fdb-f49211847b9e-2"
      ],
      "times": {
        "rt": "47h59m57.433531858s"
      }
    }
  ]
}
```
执行失败:
```json
{
  "code": 1002,
  "message": "execution failed: [Step: 2, Name: 98ab2628-f94e-4a18-933c-d8cebcc3343c-2, Exit Code: 2]",
  "data": [
    {
      "step": 0,
      "url": "http://localhost:2376/api/v1/task/98ab2628-f94e-4a18-933c-d8cebcc3343c/0/console",
      "name": "98ab2628-f94e-4a18-933c-d8cebcc3343c-0",
      "state": 0,
      "code": 0,
      "msg": "execution succeed",
      "times": {
        "st": "2022-11-18T16:14:23+08:00",
        "et": "2022-11-18T16:14:26+08:00",
        "rt": "47h59m57.363535524s"
      }
    },
    {
      "step": 1,
      "url": "http://localhost:2376/api/v1/task/98ab2628-f94e-4a18-933c-d8cebcc3343c/1/console",
      "name": "98ab2628-f94e-4a18-933c-d8cebcc3343c-1",
      "state": 0,
      "code": 0,
      "msg": "execution succeed",
      "depets_on": [
        "98ab2628-f94e-4a18-933c-d8cebcc3343c-0"
      ],
      "times": {
        "st": "2022-11-18T16:14:26+08:00",
        "et": "2022-11-18T16:14:26+08:00",
        "rt": "47h59m57.363532724s"
      }
    },
    {
      "step": 2,
      "url": "http://localhost:2376/api/v1/task/98ab2628-f94e-4a18-933c-d8cebcc3343c/2/console",
      "name": "98ab2628-f94e-4a18-933c-d8cebcc3343c-2",
      "state": 0,
      "code": 2,
      "msg": "Step: 2, Name: 98ab2628-f94e-4a18-933c-d8cebcc3343c-2, Exit Code: 2",
      "depets_on": [
        "98ab2628-f94e-4a18-933c-d8cebcc3343c-1"
      ],
      "times": {
        "st": "2022-11-18T16:14:26+08:00",
        "et": "2022-11-18T16:14:26+08:00",
        "rt": "47h59m57.363530924s"
      }
    },
    {
      "step": 3,
      "url": "http://localhost:2376/api/v1/task/98ab2628-f94e-4a18-933c-d8cebcc3343c/3/console",
      "name": "98ab2628-f94e-4a18-933c-d8cebcc3343c-3",
      "state": 2,
      "code": 0,
      "msg": "The current step only proceeds if the previous step succeeds.",
      "depets_on": [
        "98ab2628-f94e-4a18-933c-d8cebcc3343c-2"
      ],
      "times": {
        "rt": "47h59m54.363529124s"
      }
    }
  ]
}
```
步骤控制台输出:
```shell
curl http://localhost:2376/api/v1/task/98ab2628-f94e-4a18-933c-d8cebcc3343c/3/console
```

```json
{
  "code": 0,
  "message": "succeed",
  "data": [
    "正在 Ping baidu.com [110.242.68.66] 具有 32 字节的数据:",
    "来自 110.242.68.66 的回复: 字节=32 时间=45ms TTL=43",
    "来自 110.242.68.66 的回复: 字节=32 时间=45ms TTL=43",
    "来自 110.242.68.66 的回复: 字节=32 时间=45ms TTL=43",
    "来自 110.242.68.66 的回复: 字节=32 时间=45ms TTL=43",
    "110.242.68.66 的 Ping 统计信息:",
    "数据包: 已发送 = 4，已接收 = 4，丢失 = 0 (0% 丢失)，",
    "往返行程的估计时间(以毫秒为单位):",
    "最短 = 45ms，最长 = 45ms，平均 = 45ms"
  ]
}
```
不存在:
```json
{
  "code": 1003,
  "message": "task does not exist, no data"
}
```

[注释]  
+ code:  
  - 0: success
  - 1001: in progress
  - 1002: execution failed
  - 1003: data not found/destroyed
+ state:
  - 0: stop
  - 1: running
  - 2: peting
  - 255: system error
# OS Remote Executor Api

## Help
```text
usage: remote_executor.exe [<flags>] <command> [<args> ...]

Flags:
  -h, --help              Show context-sensitive help (also try --help-long and
                          --help-man).
      --addr=":2376"      host:port for execution.
      --debug             Enable debug messages
      --key_expire=48h    Set the database key expire time. Example:
                          "key_expire=1h"
      --exec_timeout=24h  Set the exec command expire time. Example:
                          "exec_timeout=30m"
      --timeout=30s       Timeout for calling endpoints on the engine
      --max-requests=0    Maximum number of concurrent requests. 0 to disable.
      --pool_size=30      Set the size of the execution work pool.
      --version           Show application version.

Commands:
  help [<command>...]
    Show help.

  run
    Run server
```

## Router
```text
[GIN-debug] GET    /debug/pprof/             --> github.com/gin-gonic/gin.WrapF.func1 (4 handlers)
[GIN-debug] GET    /debug/pprof/cmdline      --> github.com/gin-gonic/gin.WrapF.func1 (4 handlers)
[GIN-debug] GET    /debug/pprof/profile      --> github.com/gin-gonic/gin.WrapF.func1 (4 handlers)
[GIN-debug] POST   /debug/pprof/symbol       --> github.com/gin-gonic/gin.WrapF.func1 (4 handlers)
[GIN-debug] GET    /debug/pprof/symbol       --> github.com/gin-gonic/gin.WrapF.func1 (4 handlers)
[GIN-debug] GET    /debug/pprof/trace        --> github.com/gin-gonic/gin.WrapF.func1 (4 handlers)
[GIN-debug] GET    /debug/pprof/allocs       --> github.com/gin-gonic/gin.WrapH.func1 (4 handlers)
[GIN-debug] GET    /debug/pprof/block        --> github.com/gin-gonic/gin.WrapH.func1 (4 handlers)
[GIN-debug] GET    /debug/pprof/goroutine    --> github.com/gin-gonic/gin.WrapH.func1 (4 handlers)
[GIN-debug] GET    /debug/pprof/heap         --> github.com/gin-gonic/gin.WrapH.func1 (4 handlers)
[GIN-debug] GET    /debug/pprof/mutex        --> github.com/gin-gonic/gin.WrapH.func1 (4 handlers)
[GIN-debug] GET    /debug/pprof/threadcreate --> github.com/gin-gonic/gin.WrapH.func1 (4 handlers)
[GIN-debug] GET    /swagger/*any             --> github.com/swaggo/gin-swagger.CustomWrapHandler.func1 (4 handlers)
[GIN-debug] GET    /version                  --> github.com/xmapst/osreapi/handlers.Version (4 handlers)
[GIN-debug] GET    /healthyz                 --> github.com/xmapst/osreapi/handlers.Router.func2 (4 handlers)
[GIN-debug] GET    /                         --> github.com/xmapst/osreapi/handlers.List (4 handlers)
[GIN-debug] GET    /:id                      --> github.com/xmapst/osreapi/handlers.Get (4 handlers)
[GIN-debug] POST   /                         --> github.com/xmapst/osreapi/handlers.Post (4 handlers)
```

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
curl -XGET http://localhost:2376
# 按完成时间排序
curl -XGET http://localhost:2376/?sort=end
# 按过期时间排序
curl -XGET http://localhost:2376/?sort=ttl
```
```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "running": 1,
    "tasks": [
      {
        "id": "ad2fc4c8-05fc-4aab-9f28-5ceea65d0982",
        "url": "http://localhost:2376/ad2fc4c8-05fc-4aab-9f28-5ceea65d0982",
        "state": "已结束",
        "code": 510,
        "count": 4,
        "message": "执行失败: [步骤: 0, 名称: step0; 步骤: 2, 名称: step2]",
        "times": {
          "begin": "2022-11-17T17:17:58+08:00",
          "end": "2022-11-17T17:17:58+08:00",
          "ttl": "47h53m30.619752642s"
        }
      },
      {
        "id": "ed95e570-ce58-4f97-a5ae-abd7694b8dbc",
        "url": "http://localhost:2376/ed95e570-ce58-4f97-a5ae-abd7694b8dbc",
        "state": "已结束",
        "code": 0,
        "count": 4,
        "message": "所有步骤执行成功",
        "times": {
          "begin": "2022-11-17T17:20:29+08:00",
          "end": "2022-11-17T17:21:30+08:00",
          "ttl": "47h57m2.619748142s"
        }
      },
      {
        "id": "aece7b89-3081-4294-ac78-0a9b0987a493",
        "url": "http://localhost:2376/aece7b89-3081-4294-ac78-0a9b0987a493",
        "state": "执行中",
        "code": 0,
        "count": 4,
        "message": "当前正在执行: [步骤: 0, 名称: step0; 步骤: 2, 名称: step2]",
        "times": {
          "begin": "2022-11-17T17:24:25+08:00",
          "ttl": "47h59m57.619748942s"
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
curl -XPOST http://localhost:2376 -d '[
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
curl -XPOST http://localhost:2376/?ansync=true -d '[
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
curl -XPOST http://localhost:2376/?ansync=true -d '[
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
    "depends_on": [
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
    "depends_on": [
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
  "message": "成功",
  "data":{
    "count": 4,
    "id": "7f478334-1f44-4580-8aa9-9772b84e4bf6",
    "url": "http://localhost:2376/7f478334-1f44-4580-8aa9-9772b84e4bf6",
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
curl http://localhost:2376/a55c705f-d279-48cd-b1cb-803a345f8cd9
```

返回内容:  
成功:  
```json
{
  "code": 0,
  "message": "成功",
  "data": [
    {
      "step": 0,
      "url": "http://localhost:2376/a55c705f-d279-48cd-b1cb-803a345f8cd9/0",
      "name": "5a8fbc27-8b94-487a-adc4-174d76a16275-0",
      "state": "已结束",
      "code": 0,
      "message": "执行成功",
      "times": {
        "begin": "2022-11-18T16:20:53+08:00",
        "end": "2022-11-18T16:20:56+08:00",
        "ttl": "47h59m20.842028157s"
      }
    },
    {
      "step": 1,
      "url": "http://localhost:2376/a55c705f-d279-48cd-b1cb-803a345f8cd9/1",
      "name": "5a8fbc27-8b94-487a-adc4-174d76a16275-1",
      "state": "已结束",
      "code": 0,
      "message": "执行成功",
      "depends_on": [
        "5a8fbc27-8b94-487a-adc4-174d76a16275-0"
      ],
      "times": {
        "begin": "2022-11-18T16:20:56+08:00",
        "end": "2022-11-18T16:20:56+08:00",
        "ttl": "47h59m20.842010857s"
      }
    },
    {
      "step": 2,
      "url": "http://localhost:2376/a55c705f-d279-48cd-b1cb-803a345f8cd9/2",
      "name": "5a8fbc27-8b94-487a-adc4-174d76a16275-2",
      "state": "已结束",
      "code": 0,
      "message": "执行成功",
      "depends_on": [
        "5a8fbc27-8b94-487a-adc4-174d76a16275-1"
      ],
      "times": {
        "begin": "2022-11-18T16:20:56+08:00",
        "end": "2022-11-18T16:21:26+08:00",
        "ttl": "47h59m50.842009757s"
      }
    },
    {
      "step": 3,
      "url": "http://localhost:2376/a55c705f-d279-48cd-b1cb-803a345f8cd9/3",
      "name": "5a8fbc27-8b94-487a-adc4-174d76a16275-3",
      "state": "已结束",
      "code": 0,
      "message": "执行成功",
      "depends_on": [
        "5a8fbc27-8b94-487a-adc4-174d76a16275-2"
      ],
      "times": {
        "begin": "2022-11-18T16:21:26+08:00",
        "end": "2022-11-18T16:21:26+08:00",
        "ttl": "47h59m50.842002157s"
      }
    }
  ]
}
```

执行中:
```json
{
  "code": 1001,
  "message": "执行中: [步骤: 0, 名称: a446f074-d37f-48f9-9fdb-f49211847b9e-0]",
  "data": [
    {
      "step": 0,
      "url": "http://localhost:2376/a446f074-d37f-48f9-9fdb-f49211847b9e/0",
      "name": "a446f074-d37f-48f9-9fdb-f49211847b9e-0",
      "state": "执行中",
      "code": 0,
      "message": "步骤: 0, 名称: a446f074-d37f-48f9-9fdb-f49211847b9e-0",
      "times": {
        "begin": "2022-11-18T16:15:56+08:00",
        "ttl": "47h59m57.433539558s"
      }
    },
    {
      "step": 1,
      "url": "http://localhost:2376/a446f074-d37f-48f9-9fdb-f49211847b9e/1",
      "name": "a446f074-d37f-48f9-9fdb-f49211847b9e-1",
      "state": "等待执行",
      "code": 0,
      "message": "如上一依赖步骤执行失败则一直保持待执行, 只有上一依赖步骤成功才会执行",
      "depends_on": [
        "a446f074-d37f-48f9-9fdb-f49211847b9e-0"
      ],
      "times": {
        "ttl": "47h59m57.433535058s"
      }
    },
    {
      "step": 2,
      "url": "http://localhost:2376/a446f074-d37f-48f9-9fdb-f49211847b9e/2",
      "name": "a446f074-d37f-48f9-9fdb-f49211847b9e-2",
      "state": "等待执行",
      "code": 0,
      "message": "如上一依赖步骤执行失败则一直保持待执行, 只有上一依赖步骤成功才会执行",
      "depends_on": [
        "a446f074-d37f-48f9-9fdb-f49211847b9e-1"
      ],
      "times": {
        "ttl": "47h59m57.433533458s"
      }
    },
    {
      "step": 3,
      "url": "http://localhost:2376/a446f074-d37f-48f9-9fdb-f49211847b9e/3",
      "name": "a446f074-d37f-48f9-9fdb-f49211847b9e-3",
      "state": "等待执行",
      "code": 0,
      "message": "如上一依赖步骤执行失败则一直保持待执行, 只有上一依赖步骤成功才会执行",
      "depends_on": [
        "a446f074-d37f-48f9-9fdb-f49211847b9e-2"
      ],
      "times": {
        "ttl": "47h59m57.433531858s"
      }
    }
  ]
}
```
执行失败:
```json
{
  "code": 1002,
  "message": "执行失败: [步骤: 2, 名称: 98ab2628-f94e-4a18-933c-d8cebcc3343c-2, 退出码: 2]",
  "data": [
    {
      "step": 0,
      "url": "http://localhost:2376/98ab2628-f94e-4a18-933c-d8cebcc3343c/0",
      "name": "98ab2628-f94e-4a18-933c-d8cebcc3343c-0",
      "state": "已结束",
      "code": 0,
      "message": "执行成功",
      "times": {
        "begin": "2022-11-18T16:14:23+08:00",
        "end": "2022-11-18T16:14:26+08:00",
        "ttl": "47h59m57.363535524s"
      }
    },
    {
      "step": 1,
      "url": "http://localhost:2376/98ab2628-f94e-4a18-933c-d8cebcc3343c/1",
      "name": "98ab2628-f94e-4a18-933c-d8cebcc3343c-1",
      "state": "已结束",
      "code": 0,
      "message": "执行成功",
      "depends_on": [
        "98ab2628-f94e-4a18-933c-d8cebcc3343c-0"
      ],
      "times": {
        "begin": "2022-11-18T16:14:26+08:00",
        "end": "2022-11-18T16:14:26+08:00",
        "ttl": "47h59m57.363532724s"
      }
    },
    {
      "step": 2,
      "url": "http://localhost:2376/98ab2628-f94e-4a18-933c-d8cebcc3343c/2",
      "name": "98ab2628-f94e-4a18-933c-d8cebcc3343c-2",
      "state": "已结束",
      "code": 2,
      "message": "步骤: 2, 名称: 98ab2628-f94e-4a18-933c-d8cebcc3343c-2, 退出码: 2",
      "depends_on": [
        "98ab2628-f94e-4a18-933c-d8cebcc3343c-1"
      ],
      "times": {
        "begin": "2022-11-18T16:14:26+08:00",
        "end": "2022-11-18T16:14:26+08:00",
        "ttl": "47h59m57.363530924s"
      }
    },
    {
      "step": 3,
      "url": "http://localhost:2376/98ab2628-f94e-4a18-933c-d8cebcc3343c/3",
      "name": "98ab2628-f94e-4a18-933c-d8cebcc3343c-3",
      "state": "等待执行",
      "code": 0,
      "message": "如上一依赖步骤执行失败则一直保持待执行, 只有上一依赖步骤成功才会执行",
      "depends_on": [
        "98ab2628-f94e-4a18-933c-d8cebcc3343c-2"
      ],
      "times": {
        "ttl": "47h59m54.363529124s"
      }
    }
  ]
}
```
步骤输出:
```json
{
  "code": 0,
  "message": "成功",
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
  "message": "任务不存在, 沒有数据"
}
```

[注释]  
+ 返回code有以下:  
   - 0: 成功
   - 1001: 执行中
   - 1002: 执行失败
   - 1003: 未找到数据/已销毁

<<<<<<< HEAD
# OS Remote Executor Api

## Help
```text
usage: remote_executor-amd64 [<flags>] <command> [<args> ...]


Flags:
  -h, --[no-]help         Show context-sensitive help (also try --help-long and --help-man).
      --addr=":2376"      host:port for execution.
      --[no-]normal       Normal wait for all task execution to complete
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
[GIN-debug] POST   /api/v2/task              --> github.com/xmapst/osreapi/internal/handlers/api/v2/task.Post (6 handlers)
[GIN-debug] GET    /                         --> github.com/xmapst/osreapi/internal/handlers/api/v1/task.List (5 handlers)
[GIN-debug] POST   /                         --> github.com/xmapst/osreapi/internal/handlers/api/v1/task.Post (5 handlers)
[GIN-debug] GET    /:task                    --> github.com/xmapst/osreapi/internal/handlers/api/v1/task.Detail (5 handlers)
[GIN-debug] GET    /:task/:step/console      --> github.com/xmapst/osreapi/internal/handlers/api/v1/task.StepDetail (5 handlers)
```
=======
# Operating system remote execution interface

[![Go](https://github.com/xmapst/osreapi/actions/workflows/go.yml/badge.svg)](https://github.com/xmapst/osreapi/actions/workflows/go.yml)

一个无任何第三方依赖的跨平台自定义编排执行步骤的`API`, 基于`DAG`实现了依赖步骤依次顺序执行、非依赖步骤并发执行的调度功能.

提供API远程操作方式，批量执行Shell、Powershell、Python等命令，轻松完成运行自动化运维脚本等常见管理任务，轮询进程、安装或卸载软件、更新应用程序以及安装补丁。
>>>>>>> githubB

## 特性

- [x] 支持Windows/Linux/Mac
- [x] 动态调整工人数量
- [x] 基于有向无环(DAG)编排执行
- [x] 支持任务或步骤的强制终止
<<<<<<< HEAD
- [x] 支持任务或步骤单独的超时
- [x] 任务级的Workspace隔离
- [ ] 任务或步骤执行前/后发送通知/事件
- [ ] 任务或步骤插件实现

## 服用方式
以windows服务形式部署运行
### 用管理模式打开powershell执行
```powershell
New-Service -Name remote_executor -BinaryPathName "C:\remote_executor.exe run --addr=:2376" -DisplayName  "Remote Executor " -StartupType Automatic
sc.exe failure remote_executor reset= 0 actions= restart/0/restart/0/restart/0
sc.exe start remote_executor
=======
- [x] 支持任务或步骤的挂起与恢复
- [x] 支持任务或步骤单独的超时
- [x] 任务级的Workspace隔离
- [x] 任务Workspace的浏览与文件上传及下载
- [x] 自更新
- [x] WebShell
- [ ] 延时任务
- [ ] 任务或步骤执行前/后发送事件
- [ ] 任务或步骤插件实现

## Help
```text
Usage:
  linux-remote_executor-amd64 [command]

Available Commands:
  client      a self-sufficient executor
  help        Help about any command
  server      start server

Flags:
      --help      Print usage
  -v, --version   Print version information and quit

Use "linux-remote_executor-amd64 [command] --help" for more information about a command.
```

## 服用方式
### Windows
用管理模式打开powershell执行
```powershell
New-Service -Name osreapi -BinaryPathName "C:\osreapi\windows-remote_executor-amd64.exe server" -DisplayName  "Remote Executor " -StartupType Automatic
sc.exe failure osreapi reset= 0 actions= restart/0/restart/0/restart/0
sc.exe start osreapi
```

### Linux
```shell
echo > /etc/systemd/system/osreapi.service <<EOF
[Unit]
Description=This is a OS Remote Executor Api
Documentation=https://github.com/xmapst/osreapi.git
After=network.target nss-lookup.target

[Service]
NoNewPrivileges=true
ExecStart=/usr/local/bin/linux-remote_executor-amd64 server
Restart=on-failure
RestartSec=10s
LimitNOFILE=infinity

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable --now osreapi.service
>>>>>>> githubB
```

## 本地编译
```shell
git clone https://github.com/xmapst/osreapi.git
cd osreapi
make
```

<<<<<<< HEAD
## Swagger
![swagger](https://raw.githubusercontent.com/xmapst/osreapi/main/img/swagger.jpg)[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fxmapst%2Fosreapi.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fxmapst%2Fosreapi?ref=badge_shield)

=======
## 示例

### 创建任务

```shell
# url参数支持
name: 自定义任务名称
timeout: 任务超时时间
env: 任务全局环境变量注入
async: 并发执行或自定义编排

# 默认按顺序执行
curl -X POST -H "Content-Type:application/json" -d '[
  {
    "type": "bash", # 支持[python2,python3,bash,sh,cmd,powershell]
    "content": "env", # 脚本内容
    "env": { # 环境变量注入
      "TEST_SITE": "www.google.com"
    }
  },
  {
    "type": "bash", # 支持[python2,python3,bash,sh,cmd,powershell]
    "content": "curl ${TEST_SITE}", # 脚本内容
    "env": { # 环境变量注入
      "TEST_SITE": "www.baidu.com"
    }
  }
]' 'http://localhost:2376/api/v1/task' 

# 并发执行
curl -X POST -H "Content-Type:application/json" -d '[
  {
    "type": "bash", # 支持[python2,python3,bash,sh,cmd,powershell]
    "content": "env", # 脚本内容
    "env": { # 环境变量注入
      "TEST_SITE": "www.google.com"
    }
  },
  {
    "type": "bash", # 支持[python2,python3,bash,sh,cmd,powershell]
    "content": "curl ${TEST_SITE}", # 脚本内容
    "env": { # 环境变量注入
      "TEST_SITE": "www.baidu.com"
    }
  }
]' 'http://localhost:2376/api/v1/task?async=true'

# 自定义编排执行
curl -X POST -H "Content-Type:application/json" -d '[
  {
    "name": "step0",
    "type": "bash", # 支持[python2,python3,bash,sh,cmd,powershell]
    "content": "env", # 脚本内容
    "env": { # 环境变量注入
      "TEST_SITE": "www.google.com"
    }
  },
  {
    "name": "step1",
    "type": "bash", # 支持[python2,python3,bash,sh,cmd,powershell]
    "content": "curl ${TEST_SITE}", # 脚本内容
    "env": { # 环境变量注入
      "TEST_SITE": "www.baidu.com"
    },
    "depends": [
      "step1"
    ]
  }
]' 'http://localhost:2376/api/v1/task?async=true'
```

### 获取任务列表

```shell
# 按开始执行时间排序
curl -X GET -H "Content-Type:application/json" 'http://localhost:2376/api/v1/task'
```

### 获取任务详情

```shell
curl -X GET -H "Content-Type:application/json" http://localhost:2376/api/v1/task/{任务名称}
```

### 获取任务工作目录

```shell
curl -X GET -H "Content-Type:application/json" http://localhost:2376/api/v1/task/{任务名称}/workspace
```

### 任务控制

```shell
# 强杀任务
curl -X PUT -H "Content-Type:application/json" http://localhost:2376/api/v1/task/{任务名称}?action=kill

# 暂停任务执行[只有待运行的任务才能暂停]
curl -X PUT -H "Content-Type:application/json" http://localhost:2376/api/v1/task/{任务名称}?action=pause

# 暂停任务执行(暂停5分钟)[只有待运行的任务才能暂停]
curl -X PUT -H "Content-Type:application/json" http://localhost:2376/api/v1/task/{任务名称}?action=pause&duration=5m

# 继续执行任务
curl -X PUT -H "Content-Type:application/json" http://localhost:2376/api/v1/task/{任务名称}?action=resume
```

### 获取步骤控制台输出

```shell
curl -X GET -H "Content-Type:application/json" http://localhost:2376/api/v1/task/{任务名称}/step/{步骤名称}
```

### 步骤控制

```shell
# 强杀步骤
curl -X PUT -H "Content-Type:application/json" http://localhost:2376/api/v1/task/{任务名称}/step/{步骤名称}?action=kill

# 暂停步骤执行[只有待运行的步骤才能暂停]
curl -X PUT -H "Content-Type:application/json" http://localhost:2376/api/v1/task/{任务名称}/step/{步骤名称}?action=pause

# 暂停步骤执行(暂停5分钟)[只有待运行的步骤才能暂停]
curl -X PUT -H "Content-Type:application/json" http://localhost:2376/api/v1/task/{任务名称}/step/{步骤名称}?action=pause&duration=5m

# 继续执行步骤
curl -X PUT -H "Content-Type:application/json" http://localhost:2376/api/v1/task/{任务名称}/step/{步骤名称}?action=resume
```
>>>>>>> githubB

[注释]  
+ code:  
  - 0: success
<<<<<<< HEAD
  - 500: internal error
  - 1000: parameter error
  - 1001: in progress
  - 1002: execution failed
  - 1003: data not found or destroyed
  - 1004: pending
+ state:
  - 0: stop
  - 1: running
  - 2: pending
  - -255: killed
  - -256: timeout
  - -999: system error

## License
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fxmapst%2Fosreapi.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fxmapst%2Fosreapi?ref=badge_large)
=======
  - 1001: running
  - 1002: failed
  - 1003: not found
  - 1004: pending
  - 1005: paused
>>>>>>> githubB

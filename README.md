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

## 特性

- [x] 支持Windows/Linux/Mac
- [x] 动态调整工人数量
- [x] 基于有向无环(DAG)编排执行
- [x] 支持任务或步骤的强制终止
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
```

## 本地编译
```shell
git clone https://github.com/xmapst/osreapi.git
cd osreapi
make
```

## Swagger
![swagger](https://raw.githubusercontent.com/xmapst/osreapi/main/img/swagger.jpg)

[注释]  
+ code:  
  - 0: success
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
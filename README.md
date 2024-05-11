# Operating system remote execution interface

[![Go](https://github.com/xmapst/osreapi/actions/workflows/go.yml/badge.svg)](https://github.com/xmapst/osreapi/actions/workflows/go.yml)

一个无任何第三方依赖的跨平台自定义编排执行步骤的`API`, 基于`DAG`实现了依赖步骤依次顺序执行、非依赖步骤并发执行的调度功能.

提供API远程操作方式，批量执行Shell、Powershell、Python等命令，轻松完成运行自动化运维脚本等常见管理任务，轮询进程、安装或卸载软件、更新应用程序以及安装补丁。

## 特性

- [x] 支持Windows/Linux/Mac
- [x] 动态调整工人数量
- [x] 基于有向无环(DAG)编排执行
- [x] 支持任务或步骤的强制终止
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
```

## 本地编译
```shell
git clone https://github.com/xmapst/osreapi.git
cd osreapi
make
```

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

[注释]  
+ code:  
  - 0: success
  - 1001: running
  - 1002: failed
  - 1003: not found
  - 1004: pending
  - 1005: paused

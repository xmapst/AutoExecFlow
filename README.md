# AutoExecFlow

[![Go](https://github.com/xmapst/AutoExecFlow/actions/workflows/go.yml/badge.svg)](https://github.com/xmapst/AutoExecFlow/actions/workflows/go.yml)

An `API` for cross-platform custom orchestration of execution steps without any third-party dependencies.
Based on `DAG` , it implements the scheduling function of sequential execution of dependent steps and concurrent execution of non-dependent steps.

It provides `API` remote operation mode, batch execution of `Shell` , `Powershell` , `Python` and other commands,
and easily completes common management tasks such as running automated operation and maintenance scripts, polling processes, installing or uninstalling software, updating applications, and installing patches.

## Operating system remote execution interface

![](images/dag.png)

## Feature

- [x] support `Windows` / `Linux` / `Mac`
- [x] Dynamically adjust the number of workers
- [x] Orchestrating execution based on directed acyclic graph ( `DAG` )
- [x] Supports forced termination of tasks or steps
- [x] Supports suspension and resumption of tasks or steps
- [x] Support timeout for tasks or steps
- [x] Task-level Workspace isolation
- [x] Browse, upload, and download tasks in Workspace
- [x] Self-update, use parameter `--self_url`
- [x] WebShell
- [ ] Support delayed Task
- [ ] Send events before/after a task or step is executed
- [ ] Task or step plugin implementation

## Help
```text
Usage:
  AutoExecFlow_linux_amd64_v1 [command]

Available Commands:
  client      a self-sufficient executor
  help        Help about any command
  server      start server

Flags:
      --help      Print usage
  -v, --version   Print version information and quit

Use "AutoExecFlow_linux_amd64_v1 [command] --help" for more information about a command.
```

## How to use
### Windows
Open PowerShell in management mode to add services
```powershell
New-Service -Name AutoExecFlow -BinaryPathName "C:\AutoExecFlow\bin\AutoExecFlow_windows_amd64_v1.exe server" -DisplayName  "AutoExecFlow " -StartupType Automatic
sc.exe failure AutoExecFlow reset= 0 actions= restart/0/restart/0/restart/0
sc.exe start AutoExecFlow
```

### Linux
```shell
echo > /etc/systemd/system/AutoExecFlow.service <<EOF
[Unit]
Description=Operating system remote execution interface
Documentation=https://github.com/busybox-org/AutoExecFlow.git
After=network.target nss-lookup.target

[Service]
NoNewPrivileges=true
ExecStart=/usr/local/AutoExecFlow/bin/AutoExecFlow_linux_amd64_v1 server
Restart=on-failure
RestartSec=10s
LimitNOFILE=infinity

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable --now AutoExecFlow.service
```

## Local compilation (Linux)

+ Depends on the Docker environment

```shell
git clone https://github.com/xmapst/AutoExecFlow.git
cd AutoExecFlow
make
```

## Request Example

### Create a task

```shell
# URL parameter support
name: "Customize task name"
timeout: "Task timeout"
env: "Task global environment variable injection"
async: "Concurrent execution or custom orchestration"

# example:
# http://localhost:2376/api/v1/task?name=test&timeout=10m&env=TEST_SITE=www.google.com&async=true

# By default, the execution is in order.
curl -X POST -H "Content-Type:application/json" -d '[
  {
    "type": "bash", # support[python2,python3,bash,sh,cmd,powershell]
    "content": "env", # Script content
    "env": { # Environment variable injection
      "TEST_SITE": "www.google.com"
    }
  },
  {
    "type": "bash", # support[python2,python3,bash,sh,cmd,powershell]
    "content": "curl ${TEST_SITE}", # Script content
    "env": { # Environment variable injection
      "TEST_SITE": "www.baidu.com"
    }
  }
]' 'http://localhost:2376/api/v1/task' 

# Concurrent Execution
curl -X POST -H "Content-Type:application/json" -d '[
  {
    "type": "bash", # support[python2,python3,bash,sh,cmd,powershell]
    "content": "env", # Script content
    "env": { # Environment variable injection
      "TEST_SITE": "www.google.com"
    }
  },
  {
    "type": "bash", # support[python2,python3,bash,sh,cmd,powershell]
    "content": "curl ${TEST_SITE}", # Script content
    "env": { # Environment variable injection
      "TEST_SITE": "www.baidu.com"
    }
  }
]' 'http://localhost:2376/api/v1/task?async=true'

# Customized orchestration execution
curl -X POST -H "Content-Type:application/json" -d '[
  {
    "name": "step0",
    "type": "bash", # support[python2,python3,bash,sh,cmd,powershell]
    "content": "env", # Script content
    "env": { # Environment variable injection
      "TEST_SITE": "www.google.com"
    }
  },
  {
    "name": "step1",
    "type": "bash", # support[python2,python3,bash,sh,cmd,powershell]
    "content": "curl ${TEST_SITE}", # Script content
    "env": { # Environment variable injection
      "TEST_SITE": "www.baidu.com"
    },
    "depends": [
      "step1"
    ]
  }
]' 'http://localhost:2376/api/v1/task?async=true'
```

### Get the task list

```shell
curl -X GET -H "Content-Type:application/json" 'http://localhost:2376/api/v1/task'
```

### Get task details

```shell
curl -X GET -H "Content-Type:application/json" http://localhost:2376/api/v1/task/{task name}
```

### Get the task working directory

```shell
curl -X GET -H "Content-Type:application/json" http://localhost:2376/api/v1/task/{task name}/workspace
```

### Task Control

```shell
# Task to force kill
curl -X PUT -H "Content-Type:application/json" http://localhost:2376/api/v1/task/{task name}?action=kill

# Pause task execution [Only pending tasks can be paused]
curl -X PUT -H "Content-Type:application/json" http://localhost:2376/api/v1/task/{task name}?action=pause

# Pause task execution (pause for 5 minutes) [Only tasks to be run can be paused]
curl -X PUT -H "Content-Type:application/json" http://localhost:2376/api/v1/task/{task name}?action=pause&duration=5m

# Continue the task
curl -X PUT -H "Content-Type:application/json" http://localhost:2376/api/v1/task/{task name}?action=resume
```

### Get step console output

```shell
curl -X GET -H "Content-Type:application/json" http://localhost:2376/api/v1/task/{task name}/step/{step name}
```

### Step Control

```shell
# Steps to force kill
curl -X PUT -H "Content-Type:application/json" http://localhost:2376/api/v1/task/{task name}/step/{step name}?action=kill

# Pause step execution [Only pending steps can be paused]
curl -X PUT -H "Content-Type:application/json" http://localhost:2376/api/v1/task/{task name}/step/{step name}?action=pause

# Pause step execution (pause for 5 minutes) [Only steps to be run can be paused]
curl -X PUT -H "Content-Type:application/json" http://localhost:2376/api/v1/task/{task name}/step/{step name}?action=pause&duration=5m

# Continue to step
curl -X PUT -H "Content-Type:application/json" http://localhost:2376/api/v1/task/{task name}/step/{step name}?action=resume
```

[Notes]  
+ code:  
  - 0: success
  - 1001: running
  - 1002: failed
  - 1003: not found
  - 1004: pending
  - 1005: paused

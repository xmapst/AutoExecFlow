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

![](images/dag_exec.png)

```text
name: 测试
desc: 这是一段任务描述
async: true
timeout: 2m
env:
  - name: GLOBAL_NAME
    value: "全局变量"
step:
  - name: shell
    desc: 执行shell脚本
    timeout: 2m
    env:
      - name: Test
        value: "test_env"
    type: sh
    content: |-
      ping -c 4 1.1.1.1
  - name: python
    desc: 执行python脚本
    timeout: 2m
    env:
      - name: Test
        value: "test_env"
    depends:
      - shell
    type: py3
    content: |-
      import subprocess
      command = ["ping", "-c", "4", "1.1.1.1"]
      try:
          result = subprocess.run(command, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True, check=True)
          print("Ping 命令的输出：")
          print(result.stdout)
      except subprocess.CalledProcessError as e:
          print("执行 ping 命令时发生错误：")
          print(e.stderr)
  - name: lua
    desc: 执行lua脚本
    timeout: 2m
    env:
      - name: Test
        value: "test_env"
    depends:
      - shell
    type: lua
    content: |-
      local cmd = require("cmd")
      function EvalCall(params)
        print(params)
        local command = "ping -c 4 1.1.1.1"
        local result, err = cmd.exec(command)
        if err then
          print("Error executing command:", err)
          return
        end
        if not(result.status == 0) then
          print("Ping failed with status:", result.status)
          return
        end
        
        print("Ping 命令的输出：")
        print(result.stdout)
      end
  - name: star
    desc: 执行starlark脚本
    env:
      - name: Test
        value: "test_env"
    depends:
      - lua
    type: star
    content: |-
      def EvalCall(params):
        print(params)
        cmd = "ping -c 4 1.1.1.1"
        exit_code, stdout, stderr = exec_command(cmd)
        if exit_code != 0:
          print("Ping 命令执行失败 (退出码: %d)" % exit_code)
          if stderr:
            print("错误输出: %s" % stderr)
          return
        print("Ping 命令的输出：")
        print(stdout)
  - name: yaegi
    desc: 执行yaegi脚本
    env:
      - name: Test
        value: "test_env"
    depends:
      - python
    type: yaegi
    content: |-
      import (
        "fmt"
        "os/exec"
      )
      func EvalCall(params map[string]interface{}) {
        fmt.Println(params)
        cmd := exec.Command("ping", "-c", "4", "1.1.1.1")
        output, err := cmd.CombinedOutput()
        if err != nil {
          fmt.Println("执行 ping 命令时发生错误：", err)
          return
        }
        fmt.Println("Ping 命令的输出：")
        fmt.Println(string(output))
      }
  - name: 聚合测试
    desc: 等待所有脚本执行完成
    env:
      - name: Test
        value: "test_env"
    depends:
      - yaegi
      - star
    type: sh
    content: |-
      echo "done done"
  - name: 测试lua-http
    desc: 测试lua执行http获取内容
    env:
      - name: Test
        value: "test_env"
    depends:
      - shell
    type: lua
    content: |-
      local http = require("http")
      local client = http.client()
      function EvalCall(params)
        local request = http.request("GET", "https://www.baidu.com")
        local result, err = client:do_request(request)
        if err then
          error(err)
          return
        end
        print(result)
      end
  - name: 多分支执行
    desc: 测试多分支执行
    env:
      - name: Test
        value: "test_env"
    type: star
    content: |-
      load('http.star', 'http')
      def EvalCall(params):
        result = http.get("https://www.baidu.com")
        if result.status_code != 200:
          log.error(result.status_code)
          return
        print(result.body())
  - name: 多分支执行2
    desc: 测试多分支执行
    env:
      - name: Test
        value: "test_env"
    depends:
      - 多分支执行
    type: yaegi
    content: |-
      import (
        "fmt"
        "io"
        "log"
        "net/http"
      )
      func EvalCall(params map[string]interface{}) {
        resp, err := http.Get("https://www.baidu.com")
        if err != nil {
          log.Fatalf("HTTP 请求失败: %v", err)
          return
        }
        defer resp.Body.Close()
        if resp.StatusCode != http.StatusOK {
          log.Printf("HTTP 请求失败，状态码: %d", resp.StatusCode)
          return
        }
        // 读取响应体
        body, err := io.ReadAll(resp.Body)
        if err != nil {
        	log.Fatalf("读取响应体失败: %v", err)
        	return
        }
        
        // 打印响应内容
        fmt.Println("HTTP 响应内容:")
        fmt.Println(string(body))
      }
```

### Create a task

```shell
# By default, the execution is in order.
curl -X POST -H "Content-Type:application/json" -d '"name": "test",
"timeout": "10m",
"env": [
  {
    "name": "TEST_SITE",
    "value" : "www.google.com"
  }
],
"step": [
  {
    "type": "bash", # support[python2,python3,bash,sh,cmd,powershell]
    "content": "env", # Script content
    "env": [ # Environment variable injection
      {
        "name": "TEST_SITE",
        "value" : "www.google.com"
      }
    ]
  },
  {
    "type": "bash", # support[python2,python3,bash,sh,cmd,powershell]
    "content": "curl ${TEST_SITE}", # Script content
    "env": [ # Environment variable injection
      {
        "name": "TEST_SITE",
        "value" : "www.baidu.com"
      }
    ]
  }
]' 'http://localhost:2376/api/v1/task' 

# Concurrent Execution
curl -X POST -H "Content-Type:application/json" -d '"name": "test",
"timeout": "10m",
"env": [
  {
    "name": "TEST_SITE",
    "value" : "www.google.com"
  }
],
"async: true,
"step": [
  {
    "type": "bash", # support[python2,python3,bash,sh,cmd,powershell]
    "content": "env", # Script content
    "env": [ # Environment variable injection
      {
        "name": "TEST_SITE",
        "value" : "www.google.com"
      }
    ]
  },
  {
    "type": "bash", # support[python2,python3,bash,sh,cmd,powershell]
    "content": "curl ${TEST_SITE}", # Script content
    "env": [ # Environment variable injection
      {
        "name": "TEST_SITE",
        "value" : "www.baidu.com"
      }
    ]
  }
]' 'http://localhost:2376/api/v1/task'

# Customized orchestration execution
curl -X POST -H "Content-Type:application/json" -d '"name": "test",
"timeout": "10m",
"env": [
  {
    "name": "TEST_SITE",
    "value" : "www.google.com"
  }
],
"async: true,
"step": [
  {
    "name": "step0",
    "type": "bash", # support[python2,python3,bash,sh,cmd,powershell]
    "content": "env", # Script content
    "env": [ # Environment variable injection
      {
        "name": "TEST_SITE",
        "value" : "www.google.com"
      }
    ]
  },
  {
    "name": "step1",
    "type": "bash", # support[python2,python3,bash,sh,cmd,powershell]
    "content": "curl ${TEST_SITE}", # Script content
    "env": [ # Environment variable injection
      {
        "name": "TEST_SITE",
        "value" : "www.baidu.com"
      }
    ],
    "depends": [
      "step1"
    ]
  }
]' 'http://localhost:2376/api/v1/task'
```

### Get the task list

```shell
curl -X GET -H "Content-Type:application/json" 'http://localhost:2376/api/v1/task'
```

### Get task details

```shell
curl -X GET -H "Content-Type:application/json" http://localhost:2376/api/v1/task/{task name}
```

### Get task step list

```shell
curl -X GET -H "Content-Type:application/json" http://localhost:2376/api/v1/task/{task name}/step
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

## Script language support
+ [bash/sh/ps1/bat/python2/python3](internal/worker/runner/exec/README.md)
+ [lua](internal/worker/runner/lua/README.md)
+ [starlark](internal/worker/runner/starlark/README.md)
+ [yaegi](internal/worker/runner/yaegi/README.md)

## Swagger API documentation
[Swagger API documentation](https://github.com/xmapst/AutoExecFlow/blob/main/docs/swagger.yaml)

![](images/swagger.png)
]()

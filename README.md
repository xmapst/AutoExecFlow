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
    "running": 0,
    "tasks": [
      {
        "id": "a55c705f-d279-48cd-b1cb-803a345f8cd9",
        "state": "已结束",
        "code": 0,
        "count": 3,
        "message": "所有步骤执行成功",
        "times": {
          "begin": "2022-11-16T21:59:49+08:00",
          "end": "2022-11-16T22:00:21+08:00",
          "ttl": "47h59m46.135436s"
        }
      },
      {
        "id": "fa2a19d3-08ea-4243-a680-d60bc5144935",
        "state": "已结束",
        "code": 0,
        "count": 3,
        "message": "所有步骤执行成功",
        "times": {
          "begin": "2022-11-16T21:59:58+08:00",
          "end": "2022-11-16T22:00:28+08:00",
          "ttl": "47h59m53.135415s"
        }
      }
    ],
    "total": 2
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
      "name": "a55c705f-d279-48cd-b1cb-803a345f8cd9-0",
      "state": "已结束",
      "code": 0,
      "message": " % Total % Received % Xferd Average Speed Time Time Time Current Dload Upload Total Spent Left Speed 0 0 0 0 0 0 0 0 --:--:-- --:--:-- --:--:-- 0 0 0 0 0 0 0 0 0 --:--:-- 0:00:01 --:--:-- 0 0 25397 0 0 0 0 0 0 --:--:-- 0:00:01 --:--:-- 0 HTTP/2 200 server: marco/2.18 date: Wed, 16 Nov 2022 13:59:50 GMT content-type: text/html content-length: 25397 vary: Accept-Encoding x-source: C/304 cache-control: max-age=600 etag: \"42a8f17ee1c7d81:0\" last-modified: Wed, 14 Sep 2022 02:27:11 GMT accept-ranges: bytes x-powered-by: ASP.NET x-request-id: 47d7fce68d12361ef87300a853827a2a; b17cedd08dea0b74b2bcdc2a1b400453 age: 397 via: T.3.N, V.mix-zj-sad2-012, T.164.H, M.cmn-gd-szx-169 ",
      "times": {
        "begin": "2022-11-16T21:59:49+08:00",
        "end": "2022-11-16T21:59:50+08:00",
        "ttl": "47h58m49.436157s"
      }
    },
    {
      "step": 1,
      "name": "a55c705f-d279-48cd-b1cb-803a345f8cd9-1",
      "state": "已结束",
      "code": 0,
      "message": "",
      "times": {
        "begin": "2022-11-16T21:59:50+08:00",
        "end": "2022-11-16T22:00:20+08:00",
        "ttl": "47h59m19.436153s"
      }
    },
    {
      "step": 2,
      "name": "a55c705f-d279-48cd-b1cb-803a345f8cd9-2",
      "state": "已结束",
      "code": 0,
      "message": " % Total % Received % Xferd Average Speed Time Time Time Current Dload Upload Total Spent Left Speed 0 0 0 0 0 0 0 0 --:--:-- --:--:-- --:--:-- 0 0 277 0 0 0 0 0 0 --:--:-- --:--:-- --:--:-- 0 HTTP/1.1 200 OK Accept-Ranges: bytes Cache-Control: private, no-cache, no-store, proxy-revalidate, no-transform Connection: keep-alive Content-Length: 277 Content-Type: text/html Date: Wed, 16 Nov 2022 14:00:21 GMT Etag: \"575e1f6f-115\" Last-Modified: Mon, 13 Jun 2016 02:50:23 GMT Pragma: no-cache Server: bfe/1.0.8.18 ",
      "times": {
        "begin": "2022-11-16T22:00:20+08:00",
        "end": "2022-11-16T22:00:21+08:00",
        "ttl": "47h59m20.436134s"
      }
    }
  ]
}
```

执行中:
```json
{
  "code": 1001,
  "message": "步骤: 1, 名称: eb766bf4-88e4-4449-ad12-bafdbf75c78e-1,步骤: 3, 名称: eb766bf4-88e4-4449-ad12-bafdbf75c78e-3, 执行中",
  "data": [
    {
      "step": 0,
      "name": "eb766bf4-88e4-4449-ad12-bafdbf75c78e-0",
      "state": "已结束",
      "code": 0,
      "message": " C:\\Windows\\system32>curl -I http://www.q1.com % Total % Received % Xferd Average Speed Time Time Time Current Dload Upload Total Spent Left Speed 0 0 0 0 0 0 0 0 --:--:-- --:--:-- --:--:-- 0 0 25397 0 0 0 0 0 0 --:--:-- --:--:-- --:--:-- 0 HTTP/1.1 200 OK Server: marco/2.18 Date: Thu, 17 Nov 2022 01:42:20 GMT Content-Type: text/html Content-Length: 25397 Connection: keep-alive Vary: Accept-Encoding X-Source: C/200 Cache-Control: max-age=600 ETag: \"42a8f17ee1c7d81:0\" Last-Modified: Wed, 14 Sep 2022 02:27:11 GMT Accept-Ranges: bytes X-Powered-By: ASP.NET X-Request-Id: 6cfd420f5bc829dc11bd04645ce29a57; e494a33dbc770066f42498c27878a744 Age: 261 Via: T.164.N, V.mix-hz-fdi-165, T.15.H, M.ctn-fj-foc-015 ",
      "times": {
        "begin": "2022-11-17T09:42:19+08:00",
        "end": "2022-11-17T09:42:20+08:00",
        "ttl": "47h59m51.3247184s"
      }
    },
    {
      "step": 1,
      "name": "eb766bf4-88e4-4449-ad12-bafdbf75c78e-1",
      "state": "执行中",
      "code": 0,
      "message": "",
      "times": {
        "begin": "2022-11-17T09:42:19+08:00",
        "ttl": "47h59m50.3247184s"
      }
    },
    {
      "step": 2,
      "name": "eb766bf4-88e4-4449-ad12-bafdbf75c78e-2",
      "state": "已结束",
      "code": 0,
      "message": " C:\\Windows\\system32>curl -I http://baidu.com % Total % Received % Xferd Average Speed Time Time Time Current Dload Upload Total Spent Left Speed 0 0 0 0 0 0 0 0 --:--:-- --:--:-- --:--:-- 0 0 81 0 0 0 0 0 0 --:--:-- --:--:-- --:--:-- 0 HTTP/1.1 200 OK Date: Thu, 17 Nov 2022 01:42:20 GMT Server: Apache Last-Modified: Tue, 12 Jan 2010 13:48:00 GMT ETag: \"51-47cf7e6ee8400\" Accept-Ranges: bytes Content-Length: 81 Cache-Control: max-age=86400 Expires: Fri, 18 Nov 2022 01:42:20 GMT Connection: Keep-Alive Content-Type: text/html ",
      "times": {
        "begin": "2022-11-17T09:42:19+08:00",
        "end": "2022-11-17T09:42:20+08:00",
        "ttl": "47h59m51.3245453s"
      }
    },
    {
      "step": 3,
      "name": "eb766bf4-88e4-4449-ad12-bafdbf75c78e-3",
      "state": "执行中",
      "code": 0,
      "message": "",
      "times": {
        "begin": "2022-11-17T09:42:19+08:00",
        "ttl": "47h59m50.3245453s"
      }
    }
  ]
}
```
执行失败:
```json
{
  "code": 1002,
  "message": "步骤: 2, 名称: 35225177-e524-439c-8416-50d4455657fd-2, 退出码: 52, 执行失败",
  "data": [
    {
      "step": 0,
      "name": "35225177-e524-439c-8416-50d4455657fd-0",
      "state": "已结束",
      "code": 0,
      "message": " C:\\Windows\\system32>curl -I http://www.q1.com % Total % Received % Xferd Average Speed Time Time Time Current Dload Upload Total Spent Left Speed 0 0 0 0 0 0 0 0 --:--:-- --:--:-- --:--:-- 0 0 25397 0 0 0 0 0 0 --:--:-- --:--:-- --:--:-- 0 HTTP/1.1 200 OK Server: marco/2.18 Date: Thu, 17 Nov 2022 01:42:52 GMT Content-Type: text/html Content-Length: 25397 Connection: keep-alive Vary: Accept-Encoding X-Source: C/200 Cache-Control: max-age=600 ETag: \"42a8f17ee1c7d81:0\" Last-Modified: Wed, 14 Sep 2022 02:27:11 GMT Accept-Ranges: bytes X-Powered-By: ASP.NET X-Request-Id: 6cfd420f5bc829dc11bd04645ce29a57; 8d3b503bd4676a143f49e9c358bac337 Age: 293 Via: T.164.N, V.mix-hz-fdi-165, T.15.H, M.ctn-fj-foc-005 ",
      "times": {
        "begin": "2022-11-17T09:42:52+08:00",
        "end": "2022-11-17T09:42:52+08:00",
        "ttl": "47h57m52.6698884s"
      }
    },
    {
      "step": 1,
      "name": "35225177-e524-439c-8416-50d4455657fd-1",
      "state": "已结束",
      "code": 0,
      "message": "",
      "times": {
        "begin": "2022-11-17T09:42:52+08:00",
        "end": "2022-11-17T09:43:23+08:00",
        "ttl": "47h58m23.6698884s"
      }
    },
    {
      "step": 2,
      "name": "35225177-e524-439c-8416-50d4455657fd-2",
      "state": "已结束",
      "code": 52,
      "message": " C:\\Windows\\system32>curl -I http://baidu.com %!T(MISSING)otal %!R(MISSING)eceived %!X(MISSING)ferd Average Speed Time Time Time Current Dload Upload Total Spent Left Speed 0 0 0 0 0 0 0 0 --:--:-- --:--:-- --:--:-- 0 0 0 0 0 0 0 0 0 --:--:-- --:--:-- --:--:-- 0 curl: (52) Empty reply from server exit status 52",
      "times": {
        "begin": "2022-11-17T09:42:52+08:00",
        "end": "2022-11-17T09:42:52+08:00",
        "ttl": "47h57m52.6698884s"
      }
    },
    {
      "step": 3,
      "name": "35225177-e524-439c-8416-50d4455657fd-3",
      "state": "已结束",
      "code": 0,
      "message": "",
      "times": {
        "begin": "2022-11-17T09:42:52+08:00",
        "end": "2022-11-17T09:43:23+08:00",
        "ttl": "47h58m23.6698884s"
      }
    }
  ]
}
```
不存在:
```json
{
  "code": 1003,
  "message": "id不存在, 沒有数据"
}
```

[注释]  
+ 返回code有以下:  
   - 0: 成功
   - 1001: 执行中
   - 1002: 执行失败
   - 1003: 未找到数据/已销毁

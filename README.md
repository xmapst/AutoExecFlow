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
      "name": "step0",
      "state": "已结束",
      "code": 0,
      "message": "PING baidu.com (110.242.68.66) 56(84) bytes of data. 64 bytes from 110.242.68.66 (110.242.68.66): icmp_seq=1 ttl=42 time=58.0 ms 64 bytes from 110.242.68.66 (110.242.68.66): icmp_seq=2 ttl=42 time=39.4 ms 64 bytes from 110.242.68.66 (110.242.68.66): icmp_seq=3 ttl=42 time=39.5 ms 64 bytes from 110.242.68.66 (110.242.68.66): icmp_seq=4 ttl=42 time=54.0 ms --- baidu.com ping statistics --- 4 packets transmitted, 4 received, 0% packet loss, time 3104ms rtt min/avg/max/mdev = 39.439/47.747/58.032/8.401 ms ",
      "times": {
        "begin": "2022-11-17T17:20:29+08:00",
        "end": "2022-11-17T17:20:33+08:00",
        "ttl": "47h58m40.860920739s"
      }
    },
    {
      "step": 1,
      "name": "step1",
      "state": "已结束",
      "code": 0,
      "message": " % Total % Received % Xferd Average Speed Time Time Time Current Dload Upload Total Spent Left Speed 0 0 0 0 0 0 0 0 --:--:-- --:--:-- --:--:-- 0 100 2443 100 2443 0 0 14716 0 --:--:-- --:--:-- --:--:-- 14806 <!DOCTYPE html> <!--STATUS OK--><html> <head><meta http-equiv=content-type content=text/html;charset=utf-8><meta http-equiv=X-UA-Compatible content=IE=Edge><meta content=always name=referrer><link rel=stylesheet type=text/css href=https://ss1.bdstatic.com/5eN1bjq8AAUYm2zgoY3K/r/www/cache/bdorz/baidu.min.css><title>百度一下，你就知道</title></head> <body link=#0000cc> <div id=wrapper> <div id=head> <div class=head_wrapper> <div class=s_form> <div class=s_form_wrapper> <div id=lg> <img hidefocus=true src=//www.baidu.com/img/bd_logo1.png width=270 height=129> </div> <form id=form name=f action=//www.baidu.com/s class=fm> <input type=hidden name=bdorz_come value=1> <input type=hidden name=ie value=utf-8> <input type=hidden name=f value=8> <input type=hidden name=rsv_bp value=1> <input type=hidden name=rsv_idx value=1> <input type=hidden name=tn value=baidu><span class=\"bg s_ipt_wr\"><input id=kw name=wd class=s_ipt value maxlength=255 autocomplete=off autofocus=autofocus></span><span class=\"bg s_btn_wr\"><input type=submit id=su value=百度一下 class=\"bg s_btn\" autofocus></span> </form> </div> </div> <div id=u1> <a href=http://news.baidu.com name=tj_trnews class=mnav>新闻</a> <a href=https://www.hao123.com name=tj_trhao123 class=mnav>hao123</a> <a href=http://map.baidu.com name=tj_trmap class=mnav>地图</a> <a href=http://v.baidu.com name=tj_trvideo class=mnav>视频</a> <a href=http://tieba.baidu.com name=tj_trtieba class=mnav>贴吧</a> <noscript> <a href=http://www.baidu.com/bdorz/login.gif?login&amp;tpl=mn&amp;u=http%3A%2F%2Fwww.baidu.com%2f%3fbdorz_come%3d1 name=tj_login class=lb>登录</a> </noscript> <script>document.write('<a href=\"http://www.baidu.com/bdorz/login.gif?login&tpl=mn&u='+ encodeURIComponent(window.location.href+ (window.location.search === \"\" ? \"?\" : \"&\")+ \"bdorz_come=1\")+ '\" name=\"tj_login\" class=\"lb\">登录</a>'); </script> <a href=//www.baidu.com/more/ name=tj_briicon class=bri style=\"display: block;\">更多产品</a> </div> </div> </div> <div id=ftCon> <div id=ftConw> <p id=lh> <a href=http://home.baidu.com>关于百度</a> <a href=http://ir.baidu.com>About Baidu</a> </p> <p id=cp>&copy;2017&nbsp;Baidu&nbsp;<a href=http://www.baidu.com/duty/>使用百度前必读</a>&nbsp; <a href=http://jianyi.baidu.com/ class=cp-feedback>意见反馈</a>&nbsp;京ICP证030173号&nbsp; <img src=//www.baidu.com/img/gs.gif> </p> </div> </div> </div> </body> </html> ",
      "depends_on": [
        "step0"
      ],
      "times": {
        "begin": "2022-11-17T17:20:33+08:00",
        "end": "2022-11-17T17:20:33+08:00",
        "ttl": "47h58m40.860905239s"
      }
    },
    {
      "step": 2,
      "name": "step2",
      "state": "已结束",
      "code": 0,
      "message": "",
      "times": {
        "begin": "2022-11-17T17:20:29+08:00",
        "end": "2022-11-17T17:21:29+08:00",
        "ttl": "47h59m36.860904239s"
      }
    },
    {
      "step": 3,
      "name": "step3",
      "state": "已结束",
      "code": 0,
      "message": " % Total % Received % Xferd Average Speed Time Time Time Current Dload Upload Total Spent Left Speed 0 0 0 0 0 0 0 0 --:--:-- --:--:-- --:--:-- 0 100 2443 100 2443 0 0 14716 0 --:--:-- --:--:-- --:--:-- 14806 <!DOCTYPE html> <!--STATUS OK--><html> <head><meta http-equiv=content-type content=text/html;charset=utf-8><meta http-equiv=X-UA-Compatible content=IE=Edge><meta content=always name=referrer><link rel=stylesheet type=text/css href=https://ss1.bdstatic.com/5eN1bjq8AAUYm2zgoY3K/r/www/cache/bdorz/baidu.min.css><title>百度一下，你就知道</title></head> <body link=#0000cc> <div id=wrapper> <div id=head> <div class=head_wrapper> <div class=s_form> <div class=s_form_wrapper> <div id=lg> <img hidefocus=true src=//www.baidu.com/img/bd_logo1.png width=270 height=129> </div> <form id=form name=f action=//www.baidu.com/s class=fm> <input type=hidden name=bdorz_come value=1> <input type=hidden name=ie value=utf-8> <input type=hidden name=f value=8> <input type=hidden name=rsv_bp value=1> <input type=hidden name=rsv_idx value=1> <input type=hidden name=tn value=baidu><span class=\"bg s_ipt_wr\"><input id=kw name=wd class=s_ipt value maxlength=255 autocomplete=off autofocus=autofocus></span><span class=\"bg s_btn_wr\"><input type=submit id=su value=百度一下 class=\"bg s_btn\" autofocus></span> </form> </div> </div> <div id=u1> <a href=http://news.baidu.com name=tj_trnews class=mnav>新闻</a> <a href=https://www.hao123.com name=tj_trhao123 class=mnav>hao123</a> <a href=http://map.baidu.com name=tj_trmap class=mnav>地图</a> <a href=http://v.baidu.com name=tj_trvideo class=mnav>视频</a> <a href=http://tieba.baidu.com name=tj_trtieba class=mnav>贴吧</a> <noscript> <a href=http://www.baidu.com/bdorz/login.gif?login&amp;tpl=mn&amp;u=http%3A%2F%2Fwww.baidu.com%2f%3fbdorz_come%3d1 name=tj_login class=lb>登录</a> </noscript> <script>document.write('<a href=\"http://www.baidu.com/bdorz/login.gif?login&tpl=mn&u='+ encodeURIComponent(window.location.href+ (window.location.search === \"\" ? \"?\" : \"&\")+ \"bdorz_come=1\")+ '\" name=\"tj_login\" class=\"lb\">登录</a>'); </script> <a href=//www.baidu.com/more/ name=tj_briicon class=bri style=\"display: block;\">更多产品</a> </div> </div> </div> <div id=ftCon> <div id=ftConw> <p id=lh> <a href=http://home.baidu.com>关于百度</a> <a href=http://ir.baidu.com>About Baidu</a> </p> <p id=cp>&copy;2017&nbsp;Baidu&nbsp;<a href=http://www.baidu.com/duty/>使用百度前必读</a>&nbsp; <a href=http://jianyi.baidu.com/ class=cp-feedback>意见反馈</a>&nbsp;京ICP证030173号&nbsp; <img src=//www.baidu.com/img/gs.gif> </p> </div> </div> </div> </body> </html> ",
      "depends_on": [
        "step2"
      ],
      "times": {
        "begin": "2022-11-17T17:21:29+08:00",
        "end": "2022-11-17T17:21:30+08:00",
        "ttl": "47h59m37.860857739s"
      }
    }
  ]
}
```

执行中:
```json
{
  "code": 1001,
  "message": "执行中: [步骤: 2, 名称: step2]",
  "data": [
    {
      "step": 0,
      "name": "step0",
      "state": "已结束",
      "code": 0,
      "message": "PING baidu.com (110.242.68.66) 56(84) bytes of data. 64 bytes from 110.242.68.66 (110.242.68.66): icmp_seq=1 ttl=42 time=58.0 ms 64 bytes from 110.242.68.66 (110.242.68.66): icmp_seq=2 ttl=42 time=39.4 ms 64 bytes from 110.242.68.66 (110.242.68.66): icmp_seq=3 ttl=42 time=39.5 ms 64 bytes from 110.242.68.66 (110.242.68.66): icmp_seq=4 ttl=42 time=54.0 ms --- baidu.com ping statistics --- 4 packets transmitted, 4 received, 0% packet loss, time 3104ms rtt min/avg/max/mdev = 39.439/47.747/58.032/8.401 ms ",
      "times": {
        "begin": "2022-11-17T17:20:29+08:00",
        "end": "2022-11-17T17:20:33+08:00",
        "ttl": "47h59m33.939310868s"
      }
    },
    {
      "step": 1,
      "name": "step1",
      "state": "已结束",
      "code": 0,
      "message": " % Total % Received % Xferd Average Speed Time Time Time Current Dload Upload Total Spent Left Speed 0 0 0 0 0 0 0 0 --:--:-- --:--:-- --:--:-- 0 100 2443 100 2443 0 0 14716 0 --:--:-- --:--:-- --:--:-- 14806 <!DOCTYPE html> <!--STATUS OK--><html> <head><meta http-equiv=content-type content=text/html;charset=utf-8><meta http-equiv=X-UA-Compatible content=IE=Edge><meta content=always name=referrer><link rel=stylesheet type=text/css href=https://ss1.bdstatic.com/5eN1bjq8AAUYm2zgoY3K/r/www/cache/bdorz/baidu.min.css><title>百度一下，你就知道</title></head> <body link=#0000cc> <div id=wrapper> <div id=head> <div class=head_wrapper> <div class=s_form> <div class=s_form_wrapper> <div id=lg> <img hidefocus=true src=//www.baidu.com/img/bd_logo1.png width=270 height=129> </div> <form id=form name=f action=//www.baidu.com/s class=fm> <input type=hidden name=bdorz_come value=1> <input type=hidden name=ie value=utf-8> <input type=hidden name=f value=8> <input type=hidden name=rsv_bp value=1> <input type=hidden name=rsv_idx value=1> <input type=hidden name=tn value=baidu><span class=\"bg s_ipt_wr\"><input id=kw name=wd class=s_ipt value maxlength=255 autocomplete=off autofocus=autofocus></span><span class=\"bg s_btn_wr\"><input type=submit id=su value=百度一下 class=\"bg s_btn\" autofocus></span> </form> </div> </div> <div id=u1> <a href=http://news.baidu.com name=tj_trnews class=mnav>新闻</a> <a href=https://www.hao123.com name=tj_trhao123 class=mnav>hao123</a> <a href=http://map.baidu.com name=tj_trmap class=mnav>地图</a> <a href=http://v.baidu.com name=tj_trvideo class=mnav>视频</a> <a href=http://tieba.baidu.com name=tj_trtieba class=mnav>贴吧</a> <noscript> <a href=http://www.baidu.com/bdorz/login.gif?login&amp;tpl=mn&amp;u=http%3A%2F%2Fwww.baidu.com%2f%3fbdorz_come%3d1 name=tj_login class=lb>登录</a> </noscript> <script>document.write('<a href=\"http://www.baidu.com/bdorz/login.gif?login&tpl=mn&u='+ encodeURIComponent(window.location.href+ (window.location.search === \"\" ? \"?\" : \"&\")+ \"bdorz_come=1\")+ '\" name=\"tj_login\" class=\"lb\">登录</a>'); </script> <a href=//www.baidu.com/more/ name=tj_briicon class=bri style=\"display: block;\">更多产品</a> </div> </div> </div> <div id=ftCon> <div id=ftConw> <p id=lh> <a href=http://home.baidu.com>关于百度</a> <a href=http://ir.baidu.com>About Baidu</a> </p> <p id=cp>&copy;2017&nbsp;Baidu&nbsp;<a href=http://www.baidu.com/duty/>使用百度前必读</a>&nbsp; <a href=http://jianyi.baidu.com/ class=cp-feedback>意见反馈</a>&nbsp;京ICP证030173号&nbsp; <img src=//www.baidu.com/img/gs.gif> </p> </div> </div> </div> </body> </html> ",
      "depends_on": [
        "step0"
      ],
      "times": {
        "begin": "2022-11-17T17:20:33+08:00",
        "end": "2022-11-17T17:20:33+08:00",
        "ttl": "47h59m33.939252768s"
      }
    },
    {
      "step": 2,
      "name": "step2",
      "state": "执行中",
      "code": 0,
      "message": "",
      "times": {
        "begin": "2022-11-17T17:20:29+08:00",
        "ttl": "47h59m29.939247268s"
      }
    },
    {
      "step": 3,
      "name": "step3",
      "state": "等待执行",
      "code": 0,
      "message": "如上一依赖步骤执行失败则一直保持待执行, 只有上一依赖步骤成功才会执行",
      "depends_on": [
        "step2"
      ],
      "times": {
        "ttl": "47h59m29.939244168s"
      }
    }
  ]
}
```
执行失败:
```json
{
  "code": 1002,
  "message": "执行失败: [步骤: 0, 名称: step0, 退出码: 255; 步骤: 2, 名称: step2, 退出码: 255]",
  "data": [
    {
      "step": 0,
      "name": "step0",
      "state": "已结束",
      "code": 255,
      "message": "wrong script type",
      "times": {
        "begin": "2022-11-17T17:17:58+08:00",
        "end": "2022-11-17T17:17:58+08:00",
        "ttl": "47h59m47.964165525s"
      }
    },
    {
      "step": 1,
      "name": "step1",
      "state": "等待执行",
      "code": 0,
      "message": "如上一依赖步骤执行失败则一直保持待执行, 只有上一依赖步骤成功才会执行",
      "depends_on": [
        "step0"
      ],
      "times": {
        "ttl": "47h59m47.964161225s"
      }
    },
    {
      "step": 2,
      "name": "step2",
      "state": "已结束",
      "code": 255,
      "message": "wrong script type",
      "times": {
        "begin": "2022-11-17T17:17:58+08:00",
        "end": "2022-11-17T17:17:58+08:00",
        "ttl": "47h59m47.964158825s"
      }
    },
    {
      "step": 3,
      "name": "step3",
      "state": "等待执行",
      "code": 0,
      "message": "如上一依赖步骤执行失败则一直保持待执行, 只有上一依赖步骤成功才会执行",
      "depends_on": [
        "step2"
      ],
      "times": {
        "ttl": "47h59m47.964156425s"
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

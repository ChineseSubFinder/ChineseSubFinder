# websocket 接口设计

## ws 连接地址

暂定入口统一，后续如果新增功能，具体的内容根据 json 结构体的字段来划分

`ws://localhost:19035/ws`

## 通信的类型

* auth
* 

## 接口认证方式

当 client 链接到 server 的 ws 时，第一次交互，需要发送以下的 json 结构，用于权限的校验，否则将无法通过 ws 的注册，直接踢下线

```json
{
	"type": "auth",
	"token": "xxxxxxx"	// 由 web 登录成功后得到的 token
}
```

成功：

```json
{
    "type" ："common_reply"
	"message": "auth ok"
}
```

失败：

```json
{
    "type" ："common_reply"
	"message": "auth error"
}
```

## 获取正在运行扫描的日志

在 Client 认证通过后，每次与 Server 通信的时候都需要附带上之前获取到的 token 信息。

第一次获取日志是 Client 发起，Server 会一次性把当前最新的日志发送过去，然后后续就会针对这个 Client ，Server 主动推送实时日志到 Client 端

第一次主动获取，由 Client 发起：

```json
{
	"type": "get_running_log",
}
```

针对这个第一次的主动日志获取操作，Server 将回复

```json
{
	"type" ："running_log",
    "log": {
    		"index": 0,	// 这个字段无需关注
    		"log_lines":[
                {"level": "INFO", "date_time": "2022-02-11 08:51:16", "content": "ChineseSubFinder Version: unknow"},
                {"level": "INFO", "date_time": "2022-02-11 08:51:16", "content": "Need do Setup"}
            ]
	}
}
```

如果 Client 收到本次日志，则需要回复确认收到，以便服务器确认继续向下偏移日志内容发送

```json
{
	"type": "common_reply",
	"message": "ok"
}
```

后续的日志，理论上 Server 会**主动**每隔 5s 进行一次批量的日志汇总发给 Client，Client 仅需要拼接起来就好了，顺序和偏移问题由服务器解决。

> 拼接 log_lines 的条目即可

> Server 后续自动发送的日志逻辑，必须由 Client 先发起第一次的日志获取才会触发

```json
{
    "type" ："running_log",
	"log": {
			"index": 0,// 这个字段无需关注
			"log_lines":[
                {"level": "INFO", "date_time": "2022-02-11 08:52:16", "content": "123"},
                {"level": "INFO", "date_time": "2022-02-11 08:52:16", "content": "456"}
            ]
	}
}
```

哪怕是服务器主动发送的日志上来，也需要 Client 回复收到

```json
{
	"type": "common_reply",
	"message": "ok"
}
```


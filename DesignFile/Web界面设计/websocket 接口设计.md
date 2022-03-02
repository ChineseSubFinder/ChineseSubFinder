# websocket 接口设计

## ws 连接地址

暂定入口统一，后续如果新增功能，具体的内容根据 json 结构体的字段来划分

`ws://localhost:19035/ws`

## 通用的传输结构

```json
{
	"type": "通信的类型",
	"data": "string"
}
```

其中"type"见下面的通信的类型，具体的类型也是一个 class -> json 的string，也就是说其实通用的传输结构中的"data" 中还由一个嵌套的 json 字符串。

## 通信的类型

* auth

  ```json
  {
  	"token": "xxxxxxx"	// 由 web 登录成功后得到的 token
  }
  ```

* common_reply

  ```json
  {
  	"message": "auth ok"	// 针对不同的通信类型可能回复的 message 内容有所不同，但是至少针对这个通信类型是固定的，一般是用户反馈成功或者失败，无具体的数据
  }
  ```

* running_log

  ```json
  {
  	"index": 0,	// 这个字段无需关注
      "log_lines":[
          {"level": "INFO", "date_time": "2022-02-11 08:51:16", "content": "123"},
          {"level": "INFO", "date_time": "2022-02-11 08:51:16", "content": "456"}
      ]
  }
  ```

* sub_download_jobs_status

  ```json
  {
      "status": "running", // "waiting"，不是运行中，就是等待中
      "started_time": "2022-03-01 15:04:05",	// 任务开始的时间
      "working_unit_index": 10,	//正在处理到第几部电影或者连续剧
      "unit_count": 1000,			//一共有多少部电影或者连续剧
      "working_unit_name": "电影名称，或者连续剧的名称",
  
      "working_video_index": 10,	//正在处理到第几个视频
      "video_count": 1500,		//一共有几个视频
      "working_video_name": "电影名称，或者是连续剧中某一季、某一集的名称",	//正在处理到第几个视频
  }
  ```

下面提及对应的通信类型，其实都会转换成 json 的字符串，放入 data 字段传输。

## 接口认证方式

当 client 链接到 server 的 ws 时，第一次交互，需要发送以下的 json 结构，用于权限的校验，否则将无法通过 ws 的注册，直接踢下线。

```json
// 对应 "type": "auth"
{
	"token": "xxxxxxx"	// 由 web 登录成功后得到的 token
}
```

成功：

```json
// 对应 "type" ："common_reply"
{
	"message": "auth ok"
}
```

失败：发送完毕后，会马上执行清理这个 ws 示例的操作，就是提掉线

```json
// 对应 "type" ："common_reply"
{
	"message": "auth error"
}
```

## 获取正在运行扫描的日志

认证通过后，由 Server 主动发送给 Client

> Client 拿到信息后，需要根据拿到的 log_lines 顺序来拼接

```json
{
    "index": 0,	// 这个字段无需关注
    "log_lines":[
        {"level": "INFO", "date_time": "2022-02-11 08:51:16", "content": "123"},
        {"level": "INFO", "date_time": "2022-02-11 08:51:16", "content": "456"}
    ]
}
```

如果 Client 收到本次日志，则需要回复确认收到，以便服务器确认继续向下偏移日志内容发送

Client 回复收到成功：

```json
// 	对应 "type": "common_reply",
{
	"message": "running log recv ok"
}
```

Client 回复收到失败：

```json
//	对应 "type": "common_reply",
{
	"message": "running log recv error"
}
```

后续的日志，理论上 Server 会**主动**每隔 5s 进行一次批量的日志汇总发给 Client，Client 仅需要拼接起来就好了，顺序和偏移问题由服务器解决。

> 拼接 log_lines 的条目即可

## 字幕扫描任务状态

```json
{
    "status": "running", // "waiting"，不是运行中，就是等待中
    "started_time": "2022-03-01 15:04:05",	// 任务开始的时间
    "working_unit_index": 10,	//正在处理到第几部电影或者连续剧
    "unit_count": 1000,			//一共有多少部电影或者连续剧
    "working_unit_name": "电影名称，或者连续剧的名称",

    "working_video_index": 10,	//正在处理到第几个视频
    "video_count": 1500,		//一共有几个视频
    "working_video_name": "电影名称，或者是连续剧中某一季、某一集的名称",	//正在处理到第几个视频
}
```

根据 status 不同的值，started_time 会有不同的解释：

* status == running，那么 started_time 就是任务开始的时间，应该是小于等于当前的系统时间，也就可以推算字幕扫描开始运行多久了。
* status == waitting，那么 started_time 就是任务将要开始的时间，那么这个时间应该大于等于当前系统的时间，也就是个倒计时，还有多久才开始执行扫描

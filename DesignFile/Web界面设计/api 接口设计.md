## 界面流程设计

目前有两个主要的流程

![界面流程设计](pics/界面流程设计.png)

## Settings JSON 数据结构

```json
{
	"user_info": {
		"username": "abcd",
		"password": "123456"
	},
	"common_settings": {
		"scan_interval": "12h",
		"threads": 12,
		"run_scan_at_start_up": true,
		"movie_paths": ["aaa", "bbb"],
		"series_paths": ["ccc", "ddd"]
	},
	"advanced_settings": {
		"proxy_settings": {
			"use_http_proxy": true,
			"http_proxy_address": "123"
		},
		"debug_mode": true,
		"save_full_season_tmp_subtitles": true,
		"sub_type_priority": 1,
		"sub_name_formatter": 1,
		"save_multi_sub": true,
		"custom_video_exts": ["aaa", "bbb"],
		"fix_time_line": true,
		"topic": 0
	},
	"emby_settings": {
		"enable": true,
		"address_url": "123456",
		"api_key": "api123",
		"max_request_video_number": 1000,
		"skip_watched": true,
		"movie_directory_mapping": {
			"aa": "123",
			"bb": "456"
		},
		"series_directory_mapping": {
			"aab": "123",
			"bbc": "456"
		}
	},
	"developer_settings": {
		"bark_server_url": "bark"
	},
	"timeline_fixer_settings": null
}
```



## 接口认证方式

接口认证通过HTTP头`Authorization: Bearer <token>`传递

## API 列表

`content-type`均为`application/json`

### 获取系统的状态

`Get /system-status`

获取系统是否已经做过初始化，如果做过初始化就可以直接开始登录流程

请求参数：无

返回 HTTP 码200：

```js
{
  version: '0.0.1', // 版本号
  is_setup: false, // 系统是否已经初始化完成，true或false
}
```



### 应用初始化安装

`POST /setup`

无需权限认证，只在首次安装时有效，用于用户第一次安装程序时的引导页面。提交的时候务必要全部字段信息都填写。

> 注意，这里需要填写账号和密码的信息

请求参数：

```json
{
	"settings": {
        "user_info": {
            "username": "abcd",
            "password": "123456"
        },
        "common_settings": {
            "scan_interval": "12h",
            "threads": 12,
            "run_scan_at_start_up": true,
            "movie_paths": ["aaa", "bbb"],
            "series_paths": ["ccc", "ddd"]
        },
        "advanced_settings": {
            "proxy_settings": {
                "use_http_proxy": true,
                "http_proxy_address": "123"
            },
            "debug_mode": true,
            "save_full_season_tmp_subtitles": true,
            "sub_type_priority": 1,
            "sub_name_formatter": 1,
            "save_multi_sub": true,
            "custom_video_exts": ["aaa", "bbb"],
            "fix_time_line": true,
            "topic": 0
        },
        "emby_settings": {
            "enable": true,
            "address_url": "123456",
            "api_key": "api123",
            "max_request_video_number": 1000,
            "skip_watched": true,
            "movie_directory_mapping": {
                "aa": "123",
                "bb": "456"
            },
            "series_directory_mapping": {
                "aab": "123",
                "bbc": "456"
            }
        },
        "developer_settings": {
            "bark_server_url": "bark"
        },
        "timeline_fixer_settings": null
    }
}
```



### 用户登录

`POST /login`

请求参数：
```js
{
  "username": 'user',
  "password": 'pass',
}
```

返回 HTTP 码204：

* You need do Setup
* Username or Password Error

返回 HTTP 码200：
```js
{
  "access_token": 'xxxxxx',
  "settings": 完整的 settings 信息，密码被替换,
  "message": "xxx",
}
```



### 用户注销

`POST /logout`

将清空 AccessToken，需要验证 AccessToken 才会执行

请求参数：无

返回 HTTP 401：AccessToken 不正确

返回 HTTP 200："ok, need ReLogin"



### 修改密码

`POST /change-pwd`

修改用户的密码，需要验证 AccessToken 才会执行

请求参数：

```json
{
  "org_pwd": "xxx",
  "new_pwd": "xxx",
}
```

返回 HTTP 401：AccessToken 不正确

返回 HTTP 码204：

* You need do Setup
* Org Password Error

返回 HTTP 200："ok, need ReLogin"，然后会清空 AccessToken，需要重新登录



### 设置界面 -- 获取设置的信息

`GET /settings`

需要权限认证，这里获取到的 settings 信息与“应用初始化安装”填写的 settings 数据结构一致。

> 这里虽然也会拿到 password 信息，但是是 \*\*\*\*\*\* 6个 \* 号

请求参数：无

返回 HTTP 401：AccessToken 不正确

返回 HTTP 200：

> 注意，这里获取的是直接的 settings json 信息，没有 settings 这个 key

```json
{
        "user_info": {
            "username": "abcd",
            "password": "******"
        },
        "common_settings": {
            "scan_interval": "12h",
            "threads": 12,
            "run_scan_at_start_up": true,
            "movie_paths": ["aaa", "bbb"],
            "series_paths": ["ccc", "ddd"]
        },
        "advanced_settings": {
            "proxy_settings": {
                "use_http_proxy": true,
                "http_proxy_address": "123"
            },
            "debug_mode": true,
            "save_full_season_tmp_subtitles": true,
            "sub_type_priority": 1,
            "sub_name_formatter": 1,
            "save_multi_sub": true,
            "custom_video_exts": ["aaa", "bbb"],
            "fix_time_line": true,
            "topic": 0
        },
        "emby_settings": {
            "enable": true,
            "address_url": "123456",
            "api_key": "api123",
            "max_request_video_number": 1000,
            "skip_watched": true,
            "movie_directory_mapping": {
                "aa": "123",
                "bb": "456"
            },
            "series_directory_mapping": {
                "aab": "123",
                "bbc": "456"
            }
        },
        "developer_settings": {
            "bark_server_url": "bark"
        },
        "timeline_fixer_settings": null
    }
```



### 设置界面 -- 写入设置信息

`PUT /settings`

需要权限认证，这里获取到的 settings 信息与“应用初始化安装”填写的 settings 数据结构一致。

> 这里也需要填写 password 信息，，但是是 \*\*\*\*\*\* 6个 \* 号就行了。修改密码需要使用修改密码的接口

请求参数：

```json
{
	"settings": {
        "user_info": {
            "username": "abcd",
            "password": "123456"
        },
        "common_settings": {
            "scan_interval": "12h",
            "threads": 12,
            "run_scan_at_start_up": true,
            "movie_paths": ["aaa", "bbb"],
            "series_paths": ["ccc", "ddd"]
        },
        "advanced_settings": {
            "proxy_settings": {
                "use_http_proxy": true,
                "http_proxy_address": "123"
            },
            "debug_mode": true,
            "save_full_season_tmp_subtitles": true,
            "sub_type_priority": 1,
            "sub_name_formatter": 1,
            "save_multi_sub": true,
            "custom_video_exts": ["aaa", "bbb"],
            "fix_time_line": true,
            "topic": 0
        },
        "emby_settings": {
            "enable": true,
            "address_url": "123456",
            "api_key": "api123",
            "max_request_video_number": 1000,
            "skip_watched": true,
            "movie_directory_mapping": {
                "aa": "123",
                "bb": "456"
            },
            "series_directory_mapping": {
                "aab": "123",
                "bbc": "456"
            }
        },
        "developer_settings": {
            "bark_server_url": "bark"
        },
        "timeline_fixer_settings": null
    }
}
```

返回 HTTP 401：AccessToken 不正确

返回 HTTP 码204：

* Settings Request.Method Error

返回 HTTP 码200：

* Settings Save Success



### 检查代理服务器

`POST /check-proxy`

请求参数：

```javascript
{
  http_proxy_url: 'http://127.0.0.1:10809';
}
```

返回 HTTP 码 200：

```javascript
{
  status: 0; // OK 0 or ERROR 1
}
```



### 检查目录是否可用

`POST /check-path`

请求参数：

```javascript
{
  path: '/mnt/电影';
}
```

返回 HTTP 码 200：

```javascript
{
  status: 0; // OK 0 or ERROR 1
}
```

返回 HTTP 码 204



### 停止任务

停止正在运行的任务

`POST /jobs/stop`

请求参数：无

返回 HTTP 码 204



### 开始任务

`POST /jobs/start`

请求参数：无

返回 HTTP 码 204

## 通用错误码

### 401

未登录

### 404

请求内容不存在

### 400

参数验证错误

返回错误信息：

```javascript
{
  message: '代理URL不能为空';
}
```

### 500

其他意外情况导致的错误

```javascript
{
  message: 'xxx';
}
```
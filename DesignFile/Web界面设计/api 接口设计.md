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
		"movie_paths_mapping": {
			"aa": "123",
			"bb": "456"
		},
		"series_paths_mapping": {
			"aab": "123",
			"bbc": "456"
		}
	},
	"developer_settings": {
		"bark_server_address": "bark"
	},
	"timeline_fixer_settings": null,
    	"experimental_function": {
            "auto_change_sub_encode": {
                "enable": false,
                "des_encode_type": 0, // 默认 0 是 UTF-8，1 是 GBK
            },
           "chs_cht_changer": {
               "enable": false,
                "des_chinese_language_type": 0, // 默认 0 是 简体，1 是 繁体
           }
	}
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
            "movie_paths_mapping": {
                "aa": "123",
                "bb": "456"
            },
            "series_paths_mapping": {
                "aab": "123",
                "bbc": "456"
            }
        },
        "developer_settings": {
            "bark_server_address": "bark"
        },
        "timeline_fixer_settings": null,
        "experimental_function": {
            "auto_change_sub_encode": {
                "enable": false,
                "des_encode_type": 0, // 默认 0 是 UTF-8，1 是 GBK
            },
            "chs_cht_changer": {
               "enable": false,
                "des_chinese_language_type": 0, // 默认 0 是 简体，1 是 繁体
           }
		}
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
  valid: true; 
}
```

返回 HTTP 码 204

### 检查 Emby 目录是否可用

`POST /check-emby-path`

如果映射正确，应该需要返回对应的目录中的视频列表，如果检测到返回的列表为空，那么需要提示用户确认映射是否正确

请求参数：

```javascript
{
  address_url: "emby 的地址，因为这个时候很可能没有保存，所以需要额外传输过来"
  api_key： "emby ap，因为这个时候很可能没有保存，所以需要额外传输过来i"
  path_type: "movie"  // 或者是 series
  cfs_media_path: "X:\电影  这里的路径对应 check-path 中的路径，是本程序需要设置的媒体目录"
  emby_media_path: '/mnt/电影  这里的路径是 Emby 中的路径';
}
```

返回 HTTP 码 200：

```javascript
{
  media_list: ["电影AA", "连续剧BB"],; 
}
```

返回 HTTP 码 204



### 检查代理服务器

`POST /check-proxy`

请求参数：

```javascript
{
  http_proxy_address: 'http://127.0.0.1:10809';
}
```

返回 HTTP 码 200：

```javascript
{
	"sub_site_status": [{
		"name": "aa",
		"valid": true,
		"speed": 100  //ms
	}, {
		"name": "bb",
		"valid": false,
		"speed": 0
	}]
}
```



### 获取默认的设置数据结构

`GET /def-settings`

用于获取默认的设置界面使用的数据结构，无需登录。下面是示例，理论上数组和字典应该是空的

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
            "movie_paths_mapping": {
                "aa": "123",
                "bb": "456"
            },
            "series_paths_mapping": {
                "aab": "123",
                "bbc": "456"
            }
        },
        "developer_settings": {
            "bark_server_address": "bark"
        },
        "timeline_fixer_settings": null,
    	"experimental_function": {
            "auto_change_sub_encode": {
                "enable": false,
                "des_encode_type": 0, // 默认 0 是 UTF-8，1 是 GBK
            },
            "chs_cht_changer": {
               "enable": false,
                "des_chinese_language_type": 0, // 默认 0 是 简体，1 是 繁体
           }
		}
}
```



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



### 查询运行日志

`GET /running-log`

获取每一轮扫描字幕的运行日志，如果当前有正在运行的扫描任务，是不会获取到的。只能获取到完成的任务日志。

请求参数：

```json
?the_last_few_times=3 // 获取最后几次的运行日志，每次指的是一次字幕的扫描，默认获取最近运行三次的日志
```

返回 HTTP 401：AccessToken 不正确

返回 HTTP 码204：

* You need do Setup
* Org Password Error

返回 HTTP 200：

```json
{
	"recent_logs": [
		{
            "index": 0,
            "log_lines":[
                {"level": "INFO", "date_time": "2022-02-11 08:51:16", "content": "ChineseSubFinder Version: unknow"},
                {"level": "INFO", "date_time": "2022-02-11 08:51:16", "content": "Need do Setup"}
            ]
         },
        {
            "index": 1,
            "log_lines":[
                {"level": "INFO", "date_time": "2022-02-12 08:52:16", "content": "ChineseSubFinder Version: unknow"},
                {"level": "INFO", "date_time": "2022-02-12 08:52:16", "content": "Need do Setup"}
            ]
         },
	]
}
```



### V1 版本 API

#### 设置界面 -- 获取设置的信息

`GET /v1/settings`

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
            "movie_paths_mapping": {
                "aa": "123",
                "bb": "456"
            },
            "series_paths_mapping": {
                "aab": "123",
                "bbc": "456"
            }
        },
        "developer_settings": {
            "bark_server_address": "bark"
        },
        "timeline_fixer_settings": null,
    	"experimental_function": {
            "auto_change_sub_encode": {
                "enable": false,
                "des_encode_type": 0, // 默认 0 是 UTF-8，1 是 GBK
            },
            "chs_cht_changer": {
               "enable": false,
                "des_chinese_language_type": 0, // 默认 0 是 简体，1 是 繁体
           }
	 	}
    }
```



#### 设置界面 -- 写入设置信息

`PUT /v1/settings`

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
            "movie_paths_mapping": {
                "aa": "123",
                "bb": "456"
            },
            "series_paths_mapping": {
                "aab": "123",
                "bbc": "456"
            }
        },
        "developer_settings": {
            "bark_server_address": "bark"
        },
        "timeline_fixer_settings": null,
        "experimental_function": {
            "auto_change_sub_encode": {
                "enable": false,
                "des_encode_type": 0, // 默认 0 是 UTF-8，1 是 GBK
            },
            "chs_cht_changer": {
               "enable": false,
                "des_chinese_language_type": 0, // 默认 0 是 简体，1 是 繁体
           }
		}
    }
}
```

返回 HTTP 401：AccessToken 不正确

返回 HTTP 码204：

* Settings Request.Method Error

返回 HTTP 码200：

* Settings Save Success



#### 开始任务

`POST /v1/jobs/start`

请求参数：无

返回 HTTP 码 200：

```javascript
{
  message: "ok"; 
}
```

返回 HTTP 码 204



#### 停止任务

停止正在运行的任务

`POST /v1/jobs/stop`

请求参数：无

返回 HTTP 码 200：

```javascript
{
  message: "ok"; 
}
```

返回 HTTP 码 204



#### 查询任务的状态

`GET /v1/jobs/start`

请求参数：无

返回 HTTP 码 200：

```javascript
{
  status: "rinning"; // running or stoped
}
```

> running 或者是 stoped，前者是正在运行，后者是已经停止或者没有执行，如果开始的耗时的执行流程，停止的时候可以需要点时间，有可能上分钟级别。

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
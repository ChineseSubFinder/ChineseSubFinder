---
name: 这是提 Bug 的模板
about: 请尽可能填写模板中的信息，有助于定位问题
title: ""
labels: ""
assignees: ""
---

## 你使用的 chinesesubfinder 是什么版本，什么环境？

> chinesesubfinder 版本: vx.x.x
>
> 环境: docker or Window or Linux or MAC
>
> 在程序运行日志头部或者 Web UI 可以看到

## 你遇到什么问题了？

> 描述一下你遇到的问题

## 你的问题弄重现嘛？

> 能够重新，不能够重现

## 你期望的结果

> 描述以下你期望的结果

## 给出当前程序的配置文件

> /volume1/docker/chinesesubfinder/config:/config 下，ChineseSubFinderSettings.json，记得把你的用户名、密码、Emby API 敏感信息删除
>
> 可以复制上面的配置，使用类似在线 json 格式化的工具进行一次处理，方便阅读，[JSON 在线解析及格式化验证 - JSON.cn](https://www.json.cn/#)

```json
{
  "user_info": {
    "username": "xx", // 这里是用户名，需要替换掉
    "password": "xx" // 这里是密码，需要替换掉
  },
  "common_settings": {
    "scan_interval": "6h",
    "threads": 1,
    "run_scan_at_start_up": false,
    "movie_paths": ["X:\\电影"],
    "series_paths": ["X:\\连续剧"]
  },
  "advanced_settings": {
    "proxy_settings": {
      "use_http_proxy": false,
      "http_proxy_address": ""
    },
    "debug_mode": false,
    "save_full_season_tmp_subtitles": true,
    "sub_type_priority": 0,
    "sub_name_formatter": 0,
    "save_multi_sub": true,
    "custom_video_exts": [],
    "fix_time_line": false,
    "topic": 1
  },
  "emby_settings": {
    "enable": true,
    "address_url": "http://192.168.50.xxx:xxx", // 这里替换掉
    "api_key": "xxxxxxxxxxxxxx", // 这里替换掉
    "max_request_video_number": 100,
    "skip_watched": true,
    "movie_paths_mapping": {
      "X:\\电影": "/mnt/share1/电影"
    },
    "series_paths_mapping": {
      "X:\\连续剧": "/mnt/share1/连续剧"
    }
  },
  "developer_settings": {
    "enable": false,
    "bark_server_address": ""
  },
  "timeline_fixer_settings": {
    "max_offset_time": 120,
    "min_offset": 0.1
  }
}
```

## 如果是使用 Docker，请给出对应的配置信息

> 比如 Docker-Compose 信息，如果是其他方式使用忽略

## 给出媒体文件夹的目录结构

> 首先你需要去看主页的部署教程，确认你目录结构是规范的，如果你觉得没问题，那么请除非目录结构截图

> 是否看过对应的目录结构要求文档？

- [电影的推荐目录结构](https://github.com/ChineseSubFinder/ChineseSubFinder/blob/docs/DesignFile/%E7%94%B5%E5%BD%B1%E5%92%8C%E8%BF%9E%E7%BB%AD%E5%89%A7%E7%9B%AE%E5%BD%95%E7%BB%93%E6%9E%84%E7%A4%BA%E4%BE%8B.md)
- [连续剧目录结构要求](https://github.com/ChineseSubFinder/ChineseSubFinder/blob/docs/DesignFile/%E8%BF%9E%E7%BB%AD%E5%89%A7%E7%9B%AE%E5%BD%95%E7%BB%93%E6%9E%84%E8%A6%81%E6%B1%82.md)

## 请给出当次问题的完整日志

> 日志在程序的 Logs 目录下，如果你用的是 docker ，那么需要你映射出来。

> /volume1/docker/chinesesubfinder/config:/config 下，Logs 中

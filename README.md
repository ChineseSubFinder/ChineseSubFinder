# ChineseSubFinder

本项目的初衷仅仅是想自动化搞定限定条件下中文字幕下载。

> 开发中，会制作 Docker 镜像挂机用

## Why？

注意，因为近期参考《[高阶教程-追剧全流程自动化 | sleele的博客](https://sleele.com/tag/高阶教程-追剧全流程自动化/)》搞定了自动下载，美剧、电影没啥问题。但是遇到字幕下载的困难，里面推荐的都不好用，能下载一部分，大部分都不行。当然有可能是个人的问题。为此就打算自己整一个专用的下载器。

手动去下载再丢过去改名也不是不行，这不是懒嘛...

首先，明确一点，因为搞定了 sonarr 和 raddarr 以及 Emby，同时部分手动下载的视频也会使用 tinyMediaManager 去处理，所以可以认为所有的视频是都有 IMDB ID 的。那么就可以取巧，用 IMDB ID 去搜索（最差也能用标准的视频文件名称去搜索嘛）。

## 功能

支持的字幕下载站点：

* zimuku
* subhd
* shooter
* xunlei

## TODO

* 完成初版自动下载
  * ~~多个字幕网站的下载支持~~
  * ~~解析下载到的字幕是什么语言的(直接分析字幕文件)~~
  * ~~搜索视频文件~~
  * ~~日志支持~~
  * 配置文件支持
* 字幕的风评（有些字幕太差了，需要进行过滤，考虑排除，字幕组，关键词，机翻，以及评分等条件
* 加入 Web 设置界面
* docker 打包
* docker-compose 文件
  * 支持 go-rod 远程 browser
  * 本程序的部署

## 设计

![基础字幕搜索流程](DesignFile/基础字幕搜索流程.png)

## 限定条件

* 电影（暂时做这个类型，后续会考虑：连续剧、动画）

* 只搜索中文字幕

* 必要条件，视频文件经过削刮器处理

* 搜索优先级

  * 经过削刮器处理
    1. 视频经过削刮器（tinyMediaManager、Emby）处理，视频同级目录有 *.nfo 文件（Kodi 格式的）
    2. 使用 Raddarr 下载的电影， Metadata 设置 Emby，存在一个 movie.xml 文件
    3. 以上两个文件任意一个能读取到 IMDB ID
  * 通过视频文件的唯一ID（针对不同搜索方式不同）进行搜索
  * 视频文件名
  
* 支持的网站

  * subhd（根据优先级）

  * zimuku（根据优先级）

  * shooter（通过视频文件的唯一ID）

  * 迅雷（通过视频文件的唯一ID）

## 感谢

感谢下面项目的帮助

* [Andyfoo/GoSubTitleSearcher: 字幕搜索查询(go语言版)，支持4k 2160p,1080p,720p视频字幕搜索，集合了字幕库、迅雷、射手、SubHD查询接口。 (github.com)](https://github.com/Andyfoo/GoSubTitleSearcher)
* [go-rod/rod: A Devtools driver for web automation and scraping (github.com)](https://github.com/go-rod/rod)
* [ausaki/subfinder: 字幕查找器 (github.com)](https://github.com/ausaki/subfinder)

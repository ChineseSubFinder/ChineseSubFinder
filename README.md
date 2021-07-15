# ChineseSubFinder

本项目的初衷仅仅是想自动化搞定**限定条件**下 **中文** 字幕下载。

> 开发中，可能有不兼容性的调整（配置文件字段变更）
>
> 最新版本 v0.11.x 支持 Emby API，搜索效率更高，且支持刷新 Emby 加载刚下载的字幕。
>
> 关于日本动画字幕自动搜索功能的讨论，欢迎在这里提出：[关于动画字幕下载--日本](https://github.com/allanpk716/ChineseSubFinder/issues/1)

## Why？

注意，因为近期参考《[高阶教程-追剧全流程自动化 | sleele的博客](https://sleele.com/tag/高阶教程-追剧全流程自动化/)》搞定了自动下载，美剧、电影没啥问题。但是遇到字幕下载的困难，里面推荐的都不好用，能下载一部分，大部分都不行。当然有可能是个人的问题。为此就打算自己整一个专用的下载器。

手动去下载再丢过去改名也不是不行，这不是懒嘛...

首先，明确一点，因为搞定了 sonarr 和 radarr 以及 Emby，同时部分手动下载的视频也会使用 tinyMediaManager 去处理，所以可以认为所有的视频是都有 IMDB ID 的。那么就可以取巧，用 IMDB ID 去搜索（最差也能用标准的视频文件名称去搜索嘛）。

## 功能

本程序有什么功能见: [功能](https://github.com/allanpk716/ChineseSubFinder/blob/docs/DesignFile/功能.md)

## How to use

有两个文档可以参考：

* [How To Use - 原生文档](https://github.com/allanpk716/ChineseSubFinder/blob/docs/DesignFile/HowToUse.md)
* [Docker ChineseSubFinder--中文字幕自动下载 | sleele的博客 - 第三方教程](https://sleele.com/2021/06/25/docker-chinesesubfinder-中文字幕自动下载/)

高阶设置：

* [高阶设置 - Emby API 支持](https://github.com/allanpk716/ChineseSubFinder/blob/docs/DesignFile/%E9%AB%98%E9%98%B6%E8%AE%BE%E7%BD%AE%20-%20Emby%20API%20%E6%94%AF%E6%8C%81.md)

建议了解的文档，特别是对《连续剧目录结构要求》。

* [配置建议以及解释](https://github.com/allanpk716/ChineseSubFinder/blob/docs/DesignFile/%E9%85%8D%E7%BD%AE%E5%BB%BA%E8%AE%AE%E4%BB%A5%E5%8F%8A%E8%A7%A3%E9%87%8A.md)
* [连续剧目录结构要求](https://github.com/allanpk716/ChineseSubFinder/blob/docs/DesignFile/%E8%BF%9E%E7%BB%AD%E5%89%A7%E7%9B%AE%E5%BD%95%E7%BB%93%E6%9E%84%E8%A6%81%E6%B1%82.md)
* [物理路径与 docker 容器路劲映射指导](https://github.com/allanpk716/ChineseSubFinder/blob/docs/DesignFile/%E7%89%A9%E7%90%86%E8%B7%AF%E5%BE%84%E4%B8%8E%20docker%20%E5%AE%B9%E5%99%A8%E8%B7%AF%E5%8A%B2%E6%98%A0%E5%B0%84%E6%8C%87%E5%AF%BC.md)

如果文档没有及时更新，或者描述含糊、歧义的，欢迎提 [ISSUES](https://github.com/allanpk716/ChineseSubFinder/issues)。

## 其他文档

* [削刮器的推荐设置](https://github.com/allanpk716/ChineseSubFinder/blob/docs/DesignFile/%E5%89%8A%E5%88%AE%E5%99%A8%E7%9A%84%E6%8E%A8%E8%8D%90%E8%AE%BE%E7%BD%AE.md)
* [如何手动刷新 Emby 加载字幕](https://github.com/allanpk716/ChineseSubFinder/blob/docs/DesignFile/%E5%A6%82%E4%BD%95%E6%89%8B%E5%8A%A8%E5%88%B7%E6%96%B0%20Emby%20%E5%8A%A0%E8%BD%BD%E5%AD%97%E5%B9%95.md)
* [如何判断视频是否需要下载、更新字幕的](https://github.com/allanpk716/ChineseSubFinder/blob/docs/DesignFile/%E5%A6%82%E4%BD%95%E5%88%A4%E6%96%AD%E8%A7%86%E9%A2%91%E6%98%AF%E5%90%A6%E9%9C%80%E8%A6%81%E4%B8%8B%E8%BD%BD%E3%80%81%E6%9B%B4%E6%96%B0%E5%AD%97%E5%B9%95%E7%9A%84.md)
* [设计](https://github.com/allanpk716/ChineseSubFinder/blob/docs/DesignFile/%E8%AE%BE%E8%AE%A1.md)

## 版本

* v0.11.x 新增 Emby API 支持，以及其他细节修复和调整。 -- 2021年7月14日
* v0.10.x 添加额外的超时控制（最长超时时间设置为 20 min），修复特殊的双语字幕内容识别问题。 -- 2021年7月9日
* v0.9.x 新增 subhd zimuku 解析故障的通知接口，给维护人员用，可以尽快去修复解析问题。一般人员无需关心此设置。 -- 2021年6月25日
* v0.8.x 调整 docker 镜像结构 -- 2021年6月25日
* v0.7.x 提高搜索效率 -- 2021年6月25日
* v0.6.x 支持设置字幕格式的优先级 -- 2021年6月23日
* v0.5.x 支持连续剧字幕下载 -- 2021年6月19日
* v0.4.x 支持设置并发数 -- 2021年6月18日
* v0.3.x 支持连续剧字幕下载（连续剧暂时不支持 subhd） -- 2021年6月17日
* v0.2.0 docker 版本支持 subhd 的下载了，镜像体积也变大了 -- 2021年6月14日
* 完成初版，仅仅支持电影的字幕下载 -- 2021年6月13日

## TODO

见 [ToDo](https://github.com/allanpk716/ChineseSubFinder/projects/1#column-15141948)

## 感谢

感谢下面项目的帮助

* [Andyfoo/GoSubTitleSearcher: 字幕搜索查询(go语言版)](https://github.com/Andyfoo/GoSubTitleSearcher)
* [go-rod/rod: A Devtools driver for web automation and scraping](https://github.com/go-rod/rod)
* [ausaki/subfinder: 字幕查找器](https://github.com/ausaki/subfinder)
* [golandscape/sat: 高性能简繁体转换](https://github.com/golandscape/sat)


# 预览图
![Xnip2021-06-25_11-11-55](https://cdn.jsdelivr.net/gh/SuperNG6/pic@master/uPic/2021-06-25/Xnip2021-06-25_11-11-55.jpg)
![Xnip2021-06-25_11-12-33](https://cdn.jsdelivr.net/gh/SuperNG6/pic@master/uPic/2021-06-25/Xnip2021-06-25_11-12-33.jpg)
![Xnip2021-06-25_10-29-06](https://cdn.jsdelivr.net/gh/SuperNG6/pic@master/uPic/2021-06-25/Xnip2021-06-25_10-29-06.jpg)
![Xnip2021-06-25_10-24-22](https://cdn.jsdelivr.net/gh/SuperNG6/pic@master/uPic/2021-06-25/Xnip2021-06-25_10-24-22.jpg)
![Xnip2021-06-25_11-42-38](https://cdn.jsdelivr.net/gh/SuperNG6/pic@master/uPic/2021-06-25/Xnip2021-06-25_11-42-38.jpg)

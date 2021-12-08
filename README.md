# ChineseSubFinder

本项目的初衷仅仅是想自动化搞定**限定条件**下 **中文** 字幕下载。

> 开发中，可能有不兼容性的调整（配置文件字段变更）
>
> 发布的 Beta 版本可能是不稳定的，同时新增功能可能是没有文档支持的。如果没有特殊的需求，不建议使用该版本。
>
> v0.18.x 开始，暂时屏蔽了 subhd 的下载接口，后续下载字幕功能有待评估。
>
> v0.19.x 开始，升级字幕时间轴校正功能，见 [字幕时间轴校正 V2](https://github.com/allanpk716/ChineseSubFinder/blob/docs/DesignFile/%E5%AD%97%E5%B9%95%E6%97%B6%E9%97%B4%E8%BD%B4%E6%A0%A1%E6%AD%A3V2.md)

## Why？

注意，因为近期参考《[高阶教程-追剧全流程自动化 | sleele的博客](https://sleele.com/tag/高阶教程-追剧全流程自动化/)》搞定了自动下载，美剧、电影没啥问题。但是遇到字幕下载的困难，里面推荐的都不好用，能下载一部分，大部分都不行。当然有可能是个人的问题。为此就打算自己整一个专用的下载器。

手动去下载再丢过去改名也不是不行，这不是懒嘛...

首先，明确一点，因为搞定了 sonarr 和 radarr 以及 Emby，同时部分手动下载的视频也会使用 tinyMediaManager 去处理，所以可以认为所有的视频是都有 IMDB ID 的。那么就可以取巧，用 IMDB ID 去搜索（最差也能用标准的视频文件名称去搜索嘛）。

## 功能

本程序有什么功能见: [功能](https://github.com/allanpk716/ChineseSubFinder/blob/docs/DesignFile/功能.md)

## 前置要求

如果想顺利的用起来，还是对电影、连续剧的目录有一定的要求的。见文档，[电影的推荐目录结构](https://github.com/allanpk716/ChineseSubFinder/blob/docs/DesignFile/%E7%94%B5%E5%BD%B1%E5%92%8C%E8%BF%9E%E7%BB%AD%E5%89%A7%E7%9B%AE%E5%BD%95%E7%BB%93%E6%9E%84%E7%A4%BA%E4%BE%8B.md)

## How to use

有两个文档可以参考：

* [How To Use - 原生文档](https://github.com/allanpk716/ChineseSubFinder/blob/docs/DesignFile/HowToUse.md)
* [Docker ChineseSubFinder--中文字幕自动下载 | sleele的博客 - 第三方教程](https://sleele.com/2021/06/25/docker-chinesesubfinder-中文字幕自动下载/)

高阶设置：

* [高阶设置 - Emby API 支持](https://github.com/allanpk716/ChineseSubFinder/blob/docs/DesignFile/%E9%AB%98%E9%98%B6%E8%AE%BE%E7%BD%AE%20-%20Emby%20API%20%E6%94%AF%E6%8C%81.md)
* [字幕时间轴校正 V2](https://github.com/allanpk716/ChineseSubFinder/blob/docs/DesignFile/%E5%AD%97%E5%B9%95%E6%97%B6%E9%97%B4%E8%BD%B4%E6%A0%A1%E6%AD%A3V2.md)
* [强制扫描所有的视频文件下载字幕](https://github.com/allanpk716/ChineseSubFinder/blob/docs/DesignFile/强制扫描所有的视频文件下载字幕.md)

建议了解的文档，特别是对《连续剧目录结构要求》。

* [配置建议以及解释](https://github.com/allanpk716/ChineseSubFinder/blob/docs/DesignFile/%E9%85%8D%E7%BD%AE%E5%BB%BA%E8%AE%AE%E4%BB%A5%E5%8F%8A%E8%A7%A3%E9%87%8A.md)
* [连续剧目录结构要求](https://github.com/allanpk716/ChineseSubFinder/blob/docs/DesignFile/%E8%BF%9E%E7%BB%AD%E5%89%A7%E7%9B%AE%E5%BD%95%E7%BB%93%E6%9E%84%E8%A6%81%E6%B1%82.md)
* [关于字幕名称命名格式说明](https://github.com/allanpk716/ChineseSubFinder/blob/docs/DesignFile/%E5%85%B3%E4%BA%8E%E5%AD%97%E5%B9%95%E5%90%8D%E7%A7%B0%E5%91%BD%E5%90%8D%E6%A0%BC%E5%BC%8F%E8%AF%B4%E6%98%8E.md)
* [物理路径与 docker 容器路劲映射指导](https://github.com/allanpk716/ChineseSubFinder/blob/docs/DesignFile/%E7%89%A9%E7%90%86%E8%B7%AF%E5%BE%84%E4%B8%8E%20docker%20%E5%AE%B9%E5%99%A8%E8%B7%AF%E5%8A%B2%E6%98%A0%E5%B0%84%E6%8C%87%E5%AF%BC.md)

如果文档没有及时更新，或者描述含糊、歧义的，欢迎提 [ISSUES](https://github.com/allanpk716/ChineseSubFinder/issues)。

## 问题列表

如果遇到问题了，可以先看看这里总结的问题，如果未能解决，依然可以继续提问。[问题列表](https://github.com/allanpk716/ChineseSubFinder/blob/docs/DesignFile/%E9%97%AE%E9%A2%98%E5%88%97%E8%A1%A8.md)

## 其他文档

* [削刮器的推荐设置](https://github.com/allanpk716/ChineseSubFinder/blob/docs/DesignFile/%E5%89%8A%E5%88%AE%E5%99%A8%E7%9A%84%E6%8E%A8%E8%8D%90%E8%AE%BE%E7%BD%AE.md)
* [如何手动刷新 Emby 加载字幕](https://github.com/allanpk716/ChineseSubFinder/blob/docs/DesignFile/%E5%A6%82%E4%BD%95%E6%89%8B%E5%8A%A8%E5%88%B7%E6%96%B0%20Emby%20%E5%8A%A0%E8%BD%BD%E5%AD%97%E5%B9%95.md)
* [如何判断视频是否需要下载、更新字幕的](https://github.com/allanpk716/ChineseSubFinder/blob/docs/DesignFile/%E5%A6%82%E4%BD%95%E5%88%A4%E6%96%AD%E8%A7%86%E9%A2%91%E6%98%AF%E5%90%A6%E9%9C%80%E8%A6%81%E4%B8%8B%E8%BD%BD%E3%80%81%E6%9B%B4%E6%96%B0%E5%AD%97%E5%B9%95%E7%9A%84.md)
* [设计](https://github.com/allanpk716/ChineseSubFinder/blob/docs/DesignFile/%E8%AE%BE%E8%AE%A1.md)

## 如何编译此项目

go mod tidy ，然后需要设置 CGO=1 ，找到 cmd\chinesesubfinder\main.go 这个入口文件就好了。 :joy:

编译代码如下：

> cd ./cmd/chinesesubfinder \
>     && go build -ldflags="-s -w" -o /app/chinesesubfinder

跨平台是没有问题的，作者现在就是 Windows 开发的。因为手头没得 Mac OS ，也懒得整虚拟机去试，应该也是可以直接玩起来的。

## 如何参与开发

建议看 [关于中文字幕下载器的中长期规划讨论、求助](https://github.com/allanpk716/ChineseSubFinder/issues/20)，里面提及了后续的规划，需要大家的讨论。

目前阶段参与开发可以会遇到项目大范围重构，导致合并代码困难的问题。

可以协助规划和设计 Web 设置页面的需求，比如 api 接口设计什么的。

正式版本发布后，参与开发可以更加容易一些。

## 版本

* v0.19.x 调整，[字幕时间轴校正 V2](https://github.com/allanpk716/ChineseSubFinder/blob/docs/DesignFile/%E5%AD%97%E5%B9%95%E6%97%B6%E9%97%B4%E8%BD%B4%E6%A0%A1%E6%AD%A3V2.md) 功能，以及若干细节改动 --2021年12月8日
* v0.18.x 新增，[字幕时间轴自动校正](https://github.com/allanpk716/ChineseSubFinder/blob/docs/DesignFile/%E5%AD%97%E5%B9%95%E6%97%B6%E9%97%B4%E8%BD%B4%E6%A0%A1%E6%AD%A3.md)。暂时屏蔽 subhd 下载逻辑  -- 2021年10月17日
* v0.17.x 新增，代理检测模块，程序启动的时候会去 check 代理是否正常 -- 2021年9月22日
* v0.16.x 新增，启动容器/程序时，是否开始搜索并下载选项功能见[讨论](https://github.com/allanpk716/ChineseSubFinder/issues/50) -- 2021年9月18日
* v0.15.x 新增，[强制扫描所有的视频文件下载字幕](https://github.com/allanpk716/ChineseSubFinder/blob/docs/DesignFile/%E5%BC%BA%E5%88%B6%E6%89%AB%E6%8F%8F%E6%89%80%E6%9C%89%E7%9A%84%E8%A7%86%E9%A2%91%E6%96%87%E4%BB%B6%E4%B8%8B%E8%BD%BD%E5%AD%97%E5%B9%95.md)功能，但是依然跳过中文视频。 -- 2021年9月17日
* v0.14.x 修复，subhd 解析问题，新增支持[字幕命名格式转换的功能](https://github.com/allanpk716/ChineseSubFinder/blob/docs/DesignFile/%E5%85%B3%E4%BA%8E%E5%AD%97%E5%B9%95%E5%90%8D%E7%A7%B0%E5%91%BD%E5%90%8D%E6%A0%BC%E5%BC%8F%E8%AF%B4%E6%98%8E.md)。 -- 2021年9月16日
* v0.13.x 新增，高级配置，支持 Emby 任意用户看过的视频不下载字幕，修复字幕识别问题。 -- 2021年8月10日
* v0.12.x 重构，调整字幕的命名格式，移除 CGO 依赖。 -- 2021年7月26日
* v0.11.x 新增，Emby API 支持，以及其他细节修复和调整。 -- 2021年7月14日
* v0.10.x 添加额外的超时控制（最长超时时间设置为 20 min），修复特殊的双语字幕内容识别问题。 -- 2021年7月9日
* v0.9.x 新增，subhd zimuku 解析故障的通知接口，给维护人员用，可以尽快去修复解析问题。一般人员无需关心此设置。 -- 2021年6月25日
* v0.8.x 调整，docker 镜像结构 -- 2021年6月25日
* v0.7.x 提高搜索效率 -- 2021年6月25日
* v0.6.x 支持设置字幕格式的优先级 -- 2021年6月23日
* v0.5.x 支持连续剧字幕下载 -- 2021年6月19日
* v0.4.x 支持设置并发数 -- 2021年6月18日
* v0.3.x 支持连续剧字幕下载（连续剧暂时不支持 subhd） -- 2021年6月17日
* v0.2.0 docker 版本支持 subhd 的下载了，镜像体积也变大了 -- 2021年6月14日
* 完成初版，仅仅支持电影的字幕下载 -- 2021年6月13日

## TODO

见 [ChineseSubProject (github.com)](https://github.com/users/allanpk716/projects/2)

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

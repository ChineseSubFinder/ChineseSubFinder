# ChineseSubFinder

本项目的初衷仅仅是想自动化搞定**限定条件**下 **中文** 字幕下载。

> 正在实现共享字幕功能，前期欢迎讨论，也会在初版出来的时候需要有人参与内测。见：
>
> [大版本规划，以及新功能“共享字幕”功能的简介和讨论](https://github.com/allanpk716/ChineseSubFinder/issues/277)

> docker 如果拉取 latest 标签，可能在国内无法真正拉取到最新镜像，请手动指定具体的一个版本，比如: chinesesubfinder:v0.29.0

## 前置要求

如果想顺利的用起来，还是对电影、连续剧的目录有一定的要求的。见文档:

- [电影的推荐目录结构](https://github.com/allanpk716/ChineseSubFinder/blob/docs/DesignFile/%E7%94%B5%E5%BD%B1%E5%92%8C%E8%BF%9E%E7%BB%AD%E5%89%A7%E7%9B%AE%E5%BD%95%E7%BB%93%E6%9E%84%E7%A4%BA%E4%BE%8B.md)
- [连续剧目录结构要求](https://github.com/allanpk716/ChineseSubFinder/blob/docs/DesignFile/%E8%BF%9E%E7%BB%AD%E5%89%A7%E7%9B%AE%E5%BD%95%E7%BB%93%E6%9E%84%E8%A6%81%E6%B1%82.md)

## How to use

有两个文档可以参考：

- [v0.26 教程、更新说明](https://github.com/allanpk716/ChineseSubFinder/tree/docs/DesignFile/v0.26教程)
- [对外的 http api](https://github.com/allanpk716/ChineseSubFinder/tree/docs/DesignFile/ApiKey%E8%AE%BE%E8%AE%A1),以及[示例](https://github.com/allanpk716/ChineseSubFinder/issues/336)
- [Docker ChineseSubFinder--中文字幕自动下载 | sleele 的博客 - 第三方教程](https://sleele.com/2021/06/25/docker-chinesesubfinder-中文字幕自动下载/)

高阶设置：

- [字幕时间轴校正 V2](https://github.com/allanpk716/ChineseSubFinder/blob/docs/DesignFile/%E5%AD%97%E5%B9%95%E6%97%B6%E9%97%B4%E8%BD%B4%E6%A0%A1%E6%AD%A3V2.md)，有待更新 v0.20.x 对应的设置

建议了解的文档：

- [关于字幕名称命名格式说明](https://github.com/allanpk716/ChineseSubFinder/blob/docs/DesignFile/%E5%85%B3%E4%BA%8E%E5%AD%97%E5%B9%95%E5%90%8D%E7%A7%B0%E5%91%BD%E5%90%8D%E6%A0%BC%E5%BC%8F%E8%AF%B4%E6%98%8E.md)

如果文档没有及时更新，或者描述含糊、歧义的，欢迎提 [ISSUES](https://github.com/allanpk716/ChineseSubFinder/issues)。

## 问题列表

如果遇到问题了，可以先看看这里总结的问题，如果未能解决，依然可以继续提问。[问题列表](https://github.com/allanpk716/ChineseSubFinder/blob/docs/DesignFile/%E9%97%AE%E9%A2%98%E5%88%97%E8%A1%A8.md)

## 其他文档

- [削刮器的推荐设置](https://github.com/allanpk716/ChineseSubFinder/blob/docs/DesignFile/%E5%89%8A%E5%88%AE%E5%99%A8%E7%9A%84%E6%8E%A8%E8%8D%90%E8%AE%BE%E7%BD%AE.md)
- [如何判断视频是否需要下载、更新字幕的](https://github.com/allanpk716/ChineseSubFinder/blob/docs/DesignFile/%E5%A6%82%E4%BD%95%E5%88%A4%E6%96%AD%E8%A7%86%E9%A2%91%E6%98%AF%E5%90%A6%E9%9C%80%E8%A6%81%E4%B8%8B%E8%BD%BD%E3%80%81%E6%9B%B4%E6%96%B0%E5%AD%97%E5%B9%95%E7%9A%84.md)
- [设计](https://github.com/allanpk716/ChineseSubFinder/blob/docs/DesignFile/%E8%AE%BE%E8%AE%A1.md)
- [字幕时间轴校正功能实现解析(有待补全)](https://github.com/allanpk716/ChineseSubFinder/blob/docs/DesignFile/字幕时间轴校正功能实现解析/字幕时间轴校正功能实现解析.md)

## 如何编译此项目

首选需要编译 Web 部分，见 frontend/README.md

然后才能编译可执行程序部分

> 如果是 Windows，那么可以从这里下载 [MinGW-w64 - for 32 and 64 bit Windows - Browse /Toolchains targetting Win64 at SourceForge.net](https://sourceforge.net/projects/mingw-w64/files/Toolchains targetting Win64/)
>
> - [x86_64-posix-seh](https://sourceforge.net/projects/mingw-w64/files/Toolchains targetting Win64/Personal Builds/mingw-builds/8.1.0/threads-posix/seh/x86_64-8.1.0-release-posix-seh-rt_v6-rev0.7z)
>
> 后面的 CGO 编译需要：
>
> 1、新建变量: PATH，变量值为：xx\mingw64\bin
>
> 2、新建变量：LIB，变量值为：xx\mingw64\lib
>
> 3、新建变量：INCLUDE，变量值为：xx\mingw64\include
>
> 使用 gcc -v 验证是否生效

go mod tidy ，然后需要设置 CGO=1 ，找到 cmd\chinesesubfinder\main.go 这个入口文件就好了。 :joy:

编译代码如下：

> cd ./cmd/chinesesubfinder \
>  && go build -ldflags="-s -w" -o /app/chinesesubfinder

跨平台是没有问题的，作者现在就是 Windows 开发的。因为手头没得 Mac OS ，也懒得整虚拟机去试，应该也是可以直接玩起来的。

## 如何参与开发

建议看 [关于中文字幕下载器的中长期规划讨论、求助](https://github.com/allanpk716/ChineseSubFinder/issues/20)，里面提及了后续的规划，需要大家的讨论。

目前阶段参与开发可以会遇到项目大范围重构，导致合并代码困难的问题。

可以协助规划和设计 Web 设置页面的需求，比如 api 接口设计什么的。

正式版本发布后，参与开发可以更加容易一些。

## 版本

- v0.30.x 新增，“共享字幕”，低可信字幕收集功能 -- 2022 年 6 月 5 日
- v0.29.x 新增，“共享字幕”，详细见 WebUI “实验室页面”对应设置 -- 2022 年 5 月 29 日
- v0.28.x 优化，assrt 查询逻辑，对接 TMDB 中文信息查询 -- 2022 年 5 月 27 日
- v0.27.x 新增，assrt 字幕源，取消 zimuku 支持 -- 2022 年 5 月 19 日
- v0.26.x 大范围重构，详细教程和更新说明见，[v0.26 教程、更新说明](https://github.com/allanpk716/ChineseSubFinder/tree/docs/DesignFile/v0.26教程) -- 2022 年 5 月 13 日
- v0.25.x 调整细节，支持 cron 定时、指定时间、自定义 cron 规则，触发下载任务 -- 2022 年 4 月 6 日
- v0.24.x 调整细节，“实验室”添加远程 Chrome 设置 -- 2022 年 4 月 2 日
- v0.23.x 调整细节，“实验室”新增，简繁转换功能 -- 2022 年 4 月 1 日
- v0.22.x 调整细节，[v0.22.x 优化细节](https://github.com/allanpk716/ChineseSubFinder/issues/266) -- 2022 年 3 月 29 日
- v0.21.x 调整细节，[v0.21.x 优化细节](https://github.com/allanpk716/ChineseSubFinder/issues/240) -- 2022 年 2 月 6 日
- v0.20.x 重构，大范围重构，新增 Web 设置界面，支持多媒体路径 -- 2022 年 2 月 6 日
- v0.19.x 调整，[字幕时间轴校正 V2](https://github.com/allanpk716/ChineseSubFinder/blob/docs/DesignFile/%E5%AD%97%E5%B9%95%E6%97%B6%E9%97%B4%E8%BD%B4%E6%A0%A1%E6%AD%A3V2.md) 功能，以及若干细节改动 --2021 年 12 月 30 日
- v0.18.x 新增，[字幕时间轴自动校正 V1](https://github.com/allanpk716/ChineseSubFinder/blob/docs/DesignFile/%E5%AD%97%E5%B9%95%E6%97%B6%E9%97%B4%E8%BD%B4%E6%A0%A1%E6%AD%A3.md)。暂时屏蔽 subhd 下载逻辑 -- 2021 年 10 月 17 日
- v0.17.x 新增，代理检测模块，程序启动的时候会去 check 代理是否正常 -- 2021 年 9 月 22 日
- v0.16.x 新增，启动容器/程序时，是否开始搜索并下载选项功能见[讨论](https://github.com/allanpk716/ChineseSubFinder/issues/50) -- 2021 年 9 月 18 日
- v0.15.x 新增，[强制扫描所有的视频文件下载字幕](https://github.com/allanpk716/ChineseSubFinder/blob/docs/DesignFile/%E5%BC%BA%E5%88%B6%E6%89%AB%E6%8F%8F%E6%89%80%E6%9C%89%E7%9A%84%E8%A7%86%E9%A2%91%E6%96%87%E4%BB%B6%E4%B8%8B%E8%BD%BD%E5%AD%97%E5%B9%95.md)功能，但是依然跳过中文视频。 -- 2021 年 9 月 17 日
- v0.14.x 修复，subhd 解析问题，新增支持[字幕命名格式转换的功能](https://github.com/allanpk716/ChineseSubFinder/blob/docs/DesignFile/%E5%85%B3%E4%BA%8E%E5%AD%97%E5%B9%95%E5%90%8D%E7%A7%B0%E5%91%BD%E5%90%8D%E6%A0%BC%E5%BC%8F%E8%AF%B4%E6%98%8E.md)。 -- 2021 年 9 月 16 日
- v0.13.x 新增，高级配置，支持 Emby 任意用户看过的视频不下载字幕，修复字幕识别问题。 -- 2021 年 8 月 10 日
- v0.12.x 重构，调整字幕的命名格式，移除 CGO 依赖。 -- 2021 年 7 月 26 日
- v0.11.x 新增，Emby API 支持，以及其他细节修复和调整。 -- 2021 年 7 月 14 日
- v0.10.x 添加额外的超时控制（最长超时时间设置为 20 min），修复特殊的双语字幕内容识别问题。 -- 2021 年 7 月 9 日
- v0.9.x 新增，subhd zimuku 解析故障的通知接口，给维护人员用，可以尽快去修复解析问题。一般人员无需关心此设置。 -- 2021 年 6 月 25 日
- v0.8.x 调整，docker 镜像结构 -- 2021 年 6 月 25 日
- v0.7.x 提高搜索效率 -- 2021 年 6 月 25 日
- v0.6.x 支持设置字幕格式的优先级 -- 2021 年 6 月 23 日
- v0.5.x 支持连续剧字幕下载 -- 2021 年 6 月 19 日
- v0.4.x 支持设置并发数 -- 2021 年 6 月 18 日
- v0.3.x 支持连续剧字幕下载（连续剧暂时不支持 subhd） -- 2021 年 6 月 17 日
- v0.2.0 docker 版本支持 subhd 的下载了，镜像体积也变大了 -- 2021 年 6 月 14 日
- 完成初版，仅仅支持电影的字幕下载 -- 2021 年 6 月 13 日

## TODO

见 [ChineseSubProject](https://github.com/users/allanpk716/projects/2)

## 感谢

感谢 [iMyon (Myon) ](https://github.com/iMyon) 帮搞定 Web 前端部分

感谢下面项目的帮助

- [Andyfoo/GoSubTitleSearcher: 字幕搜索查询(go 语言版)](https://github.com/Andyfoo/GoSubTitleSearcher)
- [go-rod/rod: A Devtools driver for web automation and scraping](https://github.com/go-rod/rod)
- [ausaki/subfinder: 字幕查找器](https://github.com/ausaki/subfinder)
- [golandscape/sat: 高性能简繁体转换](https://github.com/golandscape/sat)
- [smacke/ffsubsync: Automagically synchronize subtitles with video](https://github.com/smacke/ffsubsync)

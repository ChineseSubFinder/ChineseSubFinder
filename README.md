# ChineseSubFinder

本项目的初衷仅仅是想自动化搞定**限定条件**下 **中文** 字幕，在**第三方**的网站或者接口的下载。

> CSF 使用交流目前只有一个 telegram 小组，https://t.me/chinesesubfinder 。
>
> 个人建议是以使用和讨论为主，bug 的反馈最好去 issue 按模板反馈和提出问题，也好有一个地方追溯。

## 提醒！

近期工作和生活繁琐事情比较多，更新的频率会下降。目前版本是足够日常简单使用的，不是严重的问题就不打算及时处理了。

这个项目写的越来越乱，还是计划重新整理一下，留下必要的功能，尽量简洁、简单，预期会用比较长的一段时间去重构，进度如何可以看 Refactor 分支。不建议过早使用该分支的输出程序。

提前祝各位新年快乐吧。2023年12月23日

## 前言

移除全功能版本，以后都是轻量级（Lite），tag 继续保留，实则都有是一个。不再直接支持某些字幕网站的下载（人多了，对方服务器扛不住），请使用第三方的字幕下载服务，subtitle best，具体请进入程序后去设置界面，会有引导。

最新的版本可以查看 [Docker Hub](https://hub.docker.com/repository/docker/allanpk716/chinesesubfinder) ，如果不在 telegram 群内，没有特殊的需求请不要选择 **Beta** 版本使用。

## 前置要求

如果想顺利的用起来，还是对电影、连续剧的目录有一定的要求的。见文档:

- [电影的推荐目录结构](https://github.com/ChineseSubFinder/ChineseSubFinder/blob/docs/DesignFile/%E7%94%B5%E5%BD%B1%E5%92%8C%E8%BF%9E%E7%BB%AD%E5%89%A7%E7%9B%AE%E5%BD%95%E7%BB%93%E6%9E%84%E7%A4%BA%E4%BE%8B.md)
- [连续剧目录结构要求](https://github.com/ChineseSubFinder/ChineseSubFinder/blob/docs/DesignFile/%E8%BF%9E%E7%BB%AD%E5%89%A7%E7%9B%AE%E5%BD%95%E7%BB%93%E6%9E%84%E8%A6%81%E6%B1%82.md)

## How to use

### 如何部署

- [Docker 部署教程](docker/readme.md)
- [如何在 Windows 上使用](https://github.com/ChineseSubFinder/ChineseSubFinder/blob/docs/DesignFile/v0.20教程/01.如何在Windows上使用.md)
- [Docker ChineseSubFinder--中文字幕自动下载 | sleele 的博客 - 第三方教程](https://sleele.com/2021/06/25/docker-chinesesubfinder-中文字幕自动下载/)

### 如何使用

* [使用教程](https://github.com/ChineseSubFinder/ChineseSubFinder/tree/docs/DesignFile/使用教程)
* [传参启动（v0.41.x 之后才支持）](https://github.com/ChineseSubFinder/ChineseSubFinder/blob/docs/DesignFile/传参启动/传参启动.md)

### API 文档文档

- [对外的 http api](https://github.com/ChineseSubFinder/ChineseSubFinder/tree/docs/DesignFile/ApiKey%E8%AE%BE%E8%AE%A1),以及[示例](https://github.com/ChineseSubFinder/ChineseSubFinder/issues/336)

### 高阶设置

- [字幕时间轴校正 V2](https://github.com/ChineseSubFinder/ChineseSubFinder/blob/docs/DesignFile/%E5%AD%97%E5%B9%95%E6%97%B6%E9%97%B4%E8%BD%B4%E6%A0%A1%E6%AD%A3V2.md)，有待更新 v0.20.x 对应的设置

建议了解的文档：

- [关于字幕名称命名格式说明](https://github.com/ChineseSubFinder/ChineseSubFinder/blob/docs/DesignFile/%E5%85%B3%E4%BA%8E%E5%AD%97%E5%B9%95%E5%90%8D%E7%A7%B0%E5%91%BD%E5%90%8D%E6%A0%BC%E5%BC%8F%E8%AF%B4%E6%98%8E.md)

如果文档没有及时更新，或者描述含糊、歧义的，欢迎提 [ISSUES](https://github.com/ChineseSubFinder/ChineseSubFinder/issues)。

## 问题列表

如果遇到问题了，可以先看看这里总结的问题，如果未能解决，依然可以继续提问。[问题列表](https://github.com/ChineseSubFinder/ChineseSubFinder/blob/docs/DesignFile/%E9%97%AE%E9%A2%98%E5%88%97%E8%A1%A8.md)

## 其他文档

- [削刮器的推荐设置](https://github.com/ChineseSubFinder/ChineseSubFinder/blob/docs/DesignFile/%E5%89%8A%E5%88%AE%E5%99%A8%E7%9A%84%E6%8E%A8%E8%8D%90%E8%AE%BE%E7%BD%AE.md)
- [如何判断视频是否需要下载、更新字幕的](https://github.com/ChineseSubFinder/ChineseSubFinder/blob/docs/DesignFile/%E5%A6%82%E4%BD%95%E5%88%A4%E6%96%AD%E8%A7%86%E9%A2%91%E6%98%AF%E5%90%A6%E9%9C%80%E8%A6%81%E4%B8%8B%E8%BD%BD%E3%80%81%E6%9B%B4%E6%96%B0%E5%AD%97%E5%B9%95%E7%9A%84.md)
- [设计](https://github.com/ChineseSubFinder/ChineseSubFinder/blob/docs/DesignFile/%E8%AE%BE%E8%AE%A1.md)
- [字幕时间轴校正功能实现解析(有待补全)](https://github.com/ChineseSubFinder/ChineseSubFinder/blob/docs/DesignFile/字幕时间轴校正功能实现解析/字幕时间轴校正功能实现解析.md)

## 如何编译此项目

* 首选需要编译 Web 部分，见 frontend/README.md

* 然后才能编译可执行程序部分

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

## 版本

请务必使用最新版本，这里忘记（懒得）写更新记录的话，可以去 [Releases](https://github.com/ChineseSubFinder/ChineseSubFinder/releases) 查看最新到什么版本了。

> 因为业余时间不多，都是断断续续做的，基本我只能记得最近两个版本的功能···

- v0.42.x 新增，支持手动上传字幕，以及在 Web 界面即可预览字幕效果，重写“库”的刷新逻辑。 -- 2022年10月31日
- ···
- 完成初版，仅仅支持电影的字幕下载 -- 2021 年 6 月 13 日

## 感谢

- [iMyon (Myon) ](https://github.com/iMyon) 帮搞定 Web 前端部分
- [devome](https://github.com/devome) 帮解决 Linux 和 Docker 编译、部署相关问题
- [宅宅还是度度](https://weibo.com/u/2884534224) 设计 Logo

感谢下面项目的帮助

- [Andyfoo/GoSubTitleSearcher: 字幕搜索查询(go 语言版)](https://github.com/Andyfoo/GoSubTitleSearcher)
- [go-rod/rod: A Devtools driver for web automation and scraping](https://github.com/go-rod/rod)
- [ausaki/subfinder: 字幕查找器](https://github.com/ausaki/subfinder)
- [golandscape/sat: 高性能简繁体转换](https://github.com/golandscape/sat)
- [smacke/ffsubsync: Automagically synchronize subtitles with video](https://github.com/smacke/ffsubsync)
- [shimberger/gohls: A server that exposes a directory for video streaming via web interface](https://github.com/shimberger/gohls)

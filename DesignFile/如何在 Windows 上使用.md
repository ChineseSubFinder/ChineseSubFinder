# 如何在 Windows 上使用

## 前言

使用本程序是有前提的，高度依赖 Emby、Jellyfin、Plex、tinyMediaManager  这类软件对你的视频（电影、连续剧）进行削刮。

### 推荐的电影目录结构

见[文档](https://github.com/allanpk716/ChineseSubFinder/blob/docs/DesignFile/%E7%94%B5%E5%BD%B1%E7%9A%84%E6%8E%A8%E8%8D%90%E7%9B%AE%E5%BD%95%E7%BB%93%E6%9E%84.md)

### 推荐的连续剧目录结构

见[文档](https://github.com/allanpk716/ChineseSubFinder/blob/docs/DesignFile/%E8%BF%9E%E7%BB%AD%E5%89%A7%E7%9B%AE%E5%BD%95%E7%BB%93%E6%9E%84%E8%A6%81%E6%B1%82.md)

## 如何使用

首先建议看一下相应的说明文档，建立一个大致的配置功能的印象。

* [配置建议以及解释](https://github.com/allanpk716/ChineseSubFinder/blob/docs/DesignFile/%E9%85%8D%E7%BD%AE%E5%BB%BA%E8%AE%AE%E4%BB%A5%E5%8F%8A%E8%A7%A3%E9%87%8A.md)

### 找到 Windows 上对应媒体文件夹目录

这里就举例找到两个目录：

* Y:\电影
* Y:\连续剧

> 并且确认你的两个文件夹进过相应的削刮处理，否则识别率很低，或者根本就会跳过识别扫描。

### 下载本程序

到以下的项目路径去下载最新的 exe，[Last Realse](https://github.com/allanpk716/ChineseSubFinder/releases)

![下载Windows版本的程序00](pics/下载Windows版本的程序00.png)

然后解压到一个地方，这里是举例，按你自己喜欢的来

![下载Windows版本的程序01](pics/下载Windows版本的程序01.png)

### 编写对应的配置文件

在上面的文件夹中新建一个文件叫做：

```
config.yaml
```

然后文件内容为

```yaml
UseProxy: false
HttpProxy: http://127.0.0.1:10809
EveryTime: 6h
Threads: 2
RunAtStartup: true
MovieFolder: Y:\电影
SeriesFolder: Y:\连续剧
```

这里具体配置的内容，请看上面提到的[配置建议以及解释](https://github.com/allanpk716/ChineseSubFinder/blob/docs/DesignFile/%E9%85%8D%E7%BD%AE%E5%BB%BA%E8%AE%AE%E4%BB%A5%E5%8F%8A%E8%A7%A3%E9%87%8A.md)，就不过多解释了。

### 运行本程序

正常来说就可以直接双击“Chinese苏北finder.exe” 就可以启动了。

![下载Windows版本的程序02](pics/下载Windows版本的程序02.png)

如果看到有一些 Warning 和 Error 一般不是很影响，等待它执行完毕吧。因为有一些是字幕文件是错误的，本程序解析的时候会提示而已。有问题可以提 issue。

## 如何设置服务启动

推荐一个 C# 的开源项目，用过很不错，[winsw/winsw: A wrapper executable that can run any executable as a Windows service, in a permissive license. (github.com)](https://github.com/winsw/winsw)

暂时偷个懒，就给网上的教程吧，一般没啥问题的。实在不行，再考虑写一下对应的教程吧。

[开机启动并隐藏 cmd 窗口程序在后台运行 - winsw 使用教程 - P3TERX ZONE](https://p3terx.com/archives/boot-and-hide-the-cmd-window-program-to-run-in-the-background-winsw-tutorial.html)
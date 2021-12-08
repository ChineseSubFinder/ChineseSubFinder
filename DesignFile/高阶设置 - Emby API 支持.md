# 高阶设置 - Emby API 支持

如题。这个功能要明确一点，这个只是本程序的加分项（吹一下），**不使用**此功能也是**足够**完成字幕的搜索和下载的。

那么目标是提供一下几个特性：

* 提高扫描影片和字幕的效率（可以知道那些是近期更新的视频，读取速度也是从 Emby 内存中拿数据，比硬盘读取快）
* 可以快速获取影片内置字幕列表（之前想做的功能，没找到相关资料怎么读取）
* 主动让 Emby 去刷新字幕列表（之前下载完字幕 Emby 很可能看不到字幕，需要等待间隔扫描或者手动刷新）

存在的问题和限制：

* 没有提供“提交字幕”到 Emby 的 API 接口（后面去 Emby 论坛提一下）
* 因为本程序和 Emby 是分开运行的，所以有一些设置需要有强关联性，否则无法使用

## 最低版本要求

> ChineseSubFinder Version > 0.11.0

## How to use

### 获取 Emby API KEY

如下图

![Emby-apikey-00](pics/Emby-apikey-00.png)

![Emby-apikey-01](pics/Emby-apikey-01.png)

### 编写 Emby Api 配置信息

> 这里都是以 docker 的部署方式来举例，请举一反三。

在原有的 ChineseSubFinder  config.yaml 中新增一下配置信息

```yaml
EmbyConfig:
    Url: http://192.168.50.x:8096
    ApiKey: 123456789
    LimitCount: 3000
    SkipWatched: false
```

那么新增后的 ChineseSubFinder  config.yaml 文件为:

```yaml
UseProxy: false
HttpProxy: http:/127.0.0.1:10809
EveryTime: 12h
Threads: 1
SaveMultiSub: true
MovieFolder: /media/电影
SeriesFolder: /media/连续剧

EmbyConfig:
    Url: http://192.168.50.x:8096
    ApiKey: 123456789
    LimitCount: 3000
    SkipWatched: false
```

### Emby 与 ChineseSubFinder 的目录映射关系

一般来说，Emby 映射的物理视频文件夹路径（ /volume1/Video ）和 ChineseSubFinder 的应该是一样的。没有什么特殊的设置，都应该直接可以支持搜索。

举例一下：

| APP              | Host/volume    | Path in container |
| :--------------- | -------------- | ----------------- |
| ChineseSubFinder | /volume1/Video | /media            |
| Emby             | /volume1/Video | /mnt/share1       |

如果是在两个 docker 中运行，那么最终映射到 docker 镜像中的路径是不一样的，所以再两个系统间需要进行一次文件 FullPath 的转换。正常来说这个可以由本程序自动完成。

这样就把 /media 与 /mnt/share1 关系起来了。那么下面的 “电影”、“连续剧” 文件夹就能够转换相对的路径了。

但是如果你的 Emby 和 ChineseSubFinder 物理视频文件夹路径都**不一样**，那么这个功能肯定是**无法用**的。

## 使用建议

* Emby 已经完成了一次所有媒体库的扫描
* 已经运行过多次或者挂机多天 ChineseSubFinder
* Threads: 1 并发数设置为 1。原因是扫描速度过快，请求给 subhd zimuku 太快，可能被拒接，所以强烈建议改为单线程，最多双线程即可···
* 合理的设置适合你自己的 LimitCount 参数

> 上面配置文件中的 LimitCount: 3000 是举例，比如你第一次使用 ChineseSubFinder 且设置了 Emby API 的时候，这个应该要设置一个较大的值，这样才能把所有Emby 扫描到的视频获取一次。同时你还需注意，如果你的 Emby 刚搭建起来还在扫描媒体库，那么此时，使用 Emby API 功能很可能效果很差的。
>
> 如果之前用过 ChineseSubFinder 的常规扫描功能做过几天的挂机处理，基本可以认为该下载的字幕早就有了。那么 LimitCount: 500 一般是足够的。
>
> 比如，你的间隔时间是 6h，除非你 6h 能丢 500+ 视频进 Emby，否则就是足够的，以此类推即可。

## 配置解释

```yaml
EmbyConfig:
    Url: http://192.168.50.x:8096
    ApiKey: 123456789
    LimitCount: 3000
    SkipWatched: false
```

* Url，Emby 的地址，目前只支持内网路径，且必须是 http
* ApiKey，Emby API Key，需要去 Emby 手动申请
* LimitCount，最多一次获取多少个近期更新的视频，包含电影和连续剧。测试设置了 3000 ，大概 10s 左右就能初步读取完信息，然后筛选出需要下载字幕的视频
* SkipWatched，默认值是 false，如果是 true 的时候，跳过Emby 任意用户看过的视频不进行字幕的搜索下载

> SkipWatched，这个功能是什么意思呢，你可以这样理解。你的 Emby 有两个用户，他们共享看一样的电影目录。假设，一个人**看完**了电影 A，那么是不是可以假设他使用了一个他**满意**的字幕去看的，那么 SkipWatched: true 的时候就会排除这个视频，不去再下载字幕了。同理，如果另一个人看了视频 B，以此类推。简单来说就是效率极高，也有风险，这个需要自己去思考一下，就不展开了。

## 可能遇到的问题

* 如果 Emby 没有进行新加入媒体文件的扫描，本功能是无法正常使用的

* subhd、zimuku 报错、提示异常增多，需要设置 Threads: 1

* [ChineseSubFinder在Dokcer bridge网络下，与EMBY（套件安装）的通信配置](https://github.com/allanpk716/ChineseSubFinder/issues/75)
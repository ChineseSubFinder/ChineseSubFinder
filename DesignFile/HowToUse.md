# How To Use

> 适用于 v0.8.x 版本的配置说明

使用本程序前，**强烈推荐**使用 Emby 或者 TinyMediaManager 对你的视频进行基础的削刮，整理好视频的命名，否则你**自行命名**连续剧是无法进行识别自动下载的。

## 配置建议

### Threads 

目前测试，设置到 6 ，群晖918+   8G 内存，是性能比较极限的数值。建议设置到 **4** 比较合适。太低就很慢，因为进行了大量的网络查询（依赖 IMDB API 以及各个字幕网站的查询接口）。太高的设置，这个看你的性能，也别太凶猛，不然被 ban IP。

### EveryTime

其实也无需经常扫描，按在下现在的使用情形举例。每天上午7点30群晖自动开机，然后本程序自动启动。设置 12h 的间隔，晚上回家吃完饭很可能电影剧集更新，正好观看。（后续考虑给出多个固定时间点的字幕扫描触发功能）

### SaveMultiSub

如果你担心本程序的自动选择最佳字幕的逻辑有问题（现在这个选择的逻辑写的很随意···），那么建议开启这个 SaveMultiSub: true。这样在视频的同级目录下会出现多个网站的 Top1 字幕。


## 使用 docker-compose 部署

> 支持 x86_64、ARM32，ARM64设备

编写以下的配置文件，注意 docker-compose 文件需要与本程序的 config.yaml 配套，特别是 MovieFolder、SeriesFolder  。

NAS 用户请注意填写用户 UID、GID，ssh进入NAS后输入 id 可获得对应账户的 UID、GID  

```yaml
version: "3"
services:
  chinesesubfinder:
    image: allanpk716/chinesesubfinder:latest
    volumes:
      - /volume1/docker/chinesesubfinder:/config
      - /volume1/Video:/media
    environment:
      - PUID=1026
      - PGID=100
      - TZ=Asia/Shanghai
    restart: unless-stopped
```

## docker 命令创建容器

````
docker create \
  --name=chinesesubfinder \
  -e PUID=1026 \
  -e PGID=100 \
  -e TZ=Asia/Shanghai \
  -v /volume1/docker/chinesesubfinder:/config \
  -v /volume1/Video:/media \
  --restart unless-stopped \
  allanpk716/chinesesubfinder:latest
````

第一次使用本容器时，请启动后立即关闭，修改 config.yaml 的媒体文件夹地址  

每次重启或更新 chinesesubfinder 容器时，系统会自动下载最新版的config.yaml.sample，可自行浏览最新配置文件并修改到config.yaml 

推荐 [使用watchtower自动更新](https://sleele.com/2019/06/16/docker更新容器镜像神器-watchtower/ ) 
 

## 配置文件解析

```yaml
UseProxy: false
HttpProxy: http://127.0.0.1:10809
EveryTime: 6h
Threads: 4
SubTypePriority: 0
DebugMode: false
SaveMultiSub: false
MovieFolder: /media/电影
SeriesFolder: /media/连续剧
```

* UseProxy，默认false。是否使用代理，需要配合 HttpProxy 设置

* HttpProxy，默认 http://127.0.0.1:10809 。http  代理这里不要留空，不适应就设置 UseProxy 为 false

* EveryTime，默认 6h。每隔多久触发一次下载逻辑。怎么用参考，[robfig/cron: a cron library for go](https://github.com/robfig/cron)

* Threads，并发数，最高到 20 个。看机器性能和网速来调整即可。

* SubTypePriority，字幕下载的优先级，0 是自动，1 是 srt 优先，2 是 ass/ssa 优先

* DebugMode，默认 false。调试模式，会在每个视频的文件夹下，新建一个  subtmp 文件夹，把所有匹配到的字幕都缓存到这个目录，没啥事可以不开。开的话就可以让你手动选择一堆的字幕啦。

* SaveMultiSub，默认值 false。true 会在每个视频下面保存每个网站找到的最佳字幕（见《[如何手动刷新 emby 加载字幕](https://github.com/allanpk716/ChineseSubFinder/blob/docs/DesignFile/%E5%A6%82%E4%BD%95%E6%89%8B%E5%8A%A8%E5%88%B7%E6%96%B0%20Emby%20%E5%8A%A0%E8%BD%BD%E5%AD%97%E5%B9%95.md)》，会举例）。false ，那么每个视频下面就一个最优字幕。

* MovieFolder，填写你的电影的目录

* SeriesFolder，填写你的连续剧的目录


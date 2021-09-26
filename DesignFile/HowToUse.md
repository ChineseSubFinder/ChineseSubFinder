# How To Use

## 前言

> 适用于 v0.12.x 版本的配置说明

使用本程序前，**强烈推荐**使用 Emby 或者 TinyMediaManager 对你的视频进行基础的削刮，整理好视频的命名，否则你**自行命名**连续剧是无法进行识别自动下载的。

本程序目前只实现了：电影、连续剧，两种类型的视频的字幕扫描支持。

> 在配置文件 config.yaml 中**必须**指定这两个目录，如果没有请指向到一个空的文件夹

## 如何在 Windows 下使用

见[文档](https://github.com/allanpk716/ChineseSubFinder/blob/docs/DesignFile/%E5%A6%82%E4%BD%95%E5%9C%A8%20Windows%20%E4%B8%8A%E4%BD%BF%E7%94%A8.md)


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

## 如何查看日志

映射日志目录出来即可，每7天回滚记录

```
- /volume1/docker/chinesesubfinder/log:/app/Logs
```

## 配置建议及解释

见，[配置建议以及解释](https://github.com/allanpk716/ChineseSubFinder/blob/docs/DesignFile/%E9%85%8D%E7%BD%AE%E5%BB%BA%E8%AE%AE%E4%BB%A5%E5%8F%8A%E8%A7%A3%E9%87%8A.md)

# 欢迎捐赠

![收款码](收款码/收款码.png)

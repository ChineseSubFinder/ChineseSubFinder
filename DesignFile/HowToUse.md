# How To Use

> 适用于 v0.8.x 版本的配置说明

使用本程序前，**强烈推荐**使用 Emby 或者 TinyMediaManager 对你的视频进行基础的削刮，整理好视频的命名，否则你**自行命名**连续剧是无法进行识别自动下载的。


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


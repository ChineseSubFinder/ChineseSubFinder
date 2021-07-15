# 物理路径与 docker 容器路劲映射指导

如题，本工具因为依赖 docker 来部署，配置尽可能简化，但是总归对一般新手的友好度还是欠佳的。那么这里写一下可能遇到的配置影视路径的情况。希望大家举一反三，不懂依然是可以提问的。

## 配置举例

### 电影、连续剧在一个根目录下

首先举例的是你的电影、连续剧都在一个根目录中的情况（这里是物理路径），如下

* /volume1/Video
* /volume1/Video/电影
* /volume1/Video/连续剧

那么，你 docker 路径映射就应该如下：

| 主机的物理路径                       | 容器中的路径 |
| ------------------------------------ | ------------ |
| /volume1/docker/chinesesubfinder     | /config      |
| /volume1/docker/chinesesubfinder/log | /app/Logs    |
| /volume1/Video                       | /media       |

本程序 config.yaml 配置信息：

```yaml
MovieFolder: /media/电影
SeriesFolder: /media/连续剧
```

### 电影、连续剧不在一个根目录下

你的电影、连续剧都在一个根目录中的情况（这里是物理路径），如下

* /volume1/AA/电影
* /volume2/BB/连续剧

那么，你 docker 路径映射就应该如下：

| 主机的物理路径                       | 容器中的路径  |
| ------------------------------------ | ------------- |
| /volume1/docker/chinesesubfinder     | /config       |
| /volume1/docker/chinesesubfinder/log | /app/Logs     |
| /volume1/AA/电影                     | /media/电影   |
| /volume2/BB/连续剧                   | /media/连续剧 |

本程序 config.yaml 配置信息：

```yaml
MovieFolder: /media/电影
SeriesFolder: /media/连续剧
```



## 遇到提示找不到 Movie 或者 Series 文件夹，如何提问

提问的时候，请给出以下信息：

你物理机器上影视的路径，比如：

 * /volume1/Video
 * /volume1/Video/电影
 * /volume1/Video/连续剧

你的 docker 配置的路径（volume）映射关系：

| 主机的物理路径                       | 容器中的路径 |
| ------------------------------------ | ------------ |
| /volume1/docker/chinesesubfinder     | /config      |
| /volume1/docker/chinesesubfinder/log | /app/Logs    |
| /volume1/Video                       | /media       |

本程序 config.yaml 配置信息：

```yaml
UseProxy: false
HttpProxy: http://192.168.50.252:20172
EveryTime: 6h
Threads: 1
DebugMode: false
SaveMultiSub: true
MovieFolder: /media/电影
SeriesFolder: /media/连续剧
```

有这些信息，一般就能帮助你定位问题了。

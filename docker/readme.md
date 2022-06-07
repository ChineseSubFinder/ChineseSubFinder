## 创建

**docker cli**

```
docker run -d \
    -v $(pwd)/config:/config   `# 冒号左边请修改为你想在主机上保存配置、日志等文件的路径` \
    -v $(pwd)/media:/media     `# 请修改为需要下载字幕的媒体目录，冒号右边可以改成你方便记忆的目录，多个媒体目录需要添加多个-v映射` \
    -e PUID=1000 \
    -e PGID=100 \
    -e PERMS=true       `# 是否重设/media权限` \
    -e TZ=Asia/Shanghai `# 时区` \
    -e UMASK=022        `# 权限掩码` \
    -p 19035:19035 `# 从0.20.0版本开始，通过webui来设置` \
    --name chinesesubfinder \
    --hostname chinesesubfinder \
    --log-driver "json-file" \
    --log-opt "max-size=100m" `# 限制日志大小，可自行调整` \
    allanpk716/chinesesubfinder
```

创建好后访问`http://<ip>:19035`来进行下一步设置。

**docker-compose**

新建`docker-compose.yml`文件如下，并以命令`docker-compose up -d`启动。

```
version: "3"
services:
  chinesesubfinder:
    image: allanpk716/chinesesubfinder:latest
    volumes:
      - ./config:/config  # 冒号左边请修改为你想在主机上保存配置、日志等文件的路径
      - ./media:/media    # 请修改为你的媒体目录，冒号右边可以改成你方便记忆的目录，多个媒体目录需要分别映射进来
    environment:
      - PUID=1000         # uid
      - PGID=100          # gid
      - PERMS=true        # 是否重设/media权限
      - TZ=Asia/Shanghai  # 时区
      - UMASK=022         # 权限掩码
    restart: always
    network_mode: bridge
    hostname: chinesesubfinder
    container_name: chinesesubfinder
    ports:
      - 19035:19035  # 从0.20.0版本开始，通过webui来设置
    logging:
        driver: "json-file"
        options:
          max-size: "100m" # 限制日志大小，可自行调整
```

创建好后访问`http://<ip>:19035`来进行下一步设置。

## 关于PUID/PGID的说明

如在使用诸如emby、jellyfin、plex、qbittorrent、transmission、deluge、jackett、sonarr、radarr等等的docker镜像，请在创建本容器时的设置和它们的PUID/PGID和它们一样，如若真的不想设置为一样，至少要保证本容器PUID/PGID所定义的用户拥有你设置的媒体目录（示例中是`/media`）的读取和写入权限。

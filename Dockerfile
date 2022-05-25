FROM library/node:16-alpine as frontBuilder

USER root
RUN mkdir -p /usr/src/app
WORKDIR /usr/src/app
add ./frontend/.npmrc /usr/src/app
add ./frontend/package.json /usr/src/app
add ./frontend/package-lock.json /usr/src/app
RUN npm ci
COPY ./frontend /usr/src/app
RUN ls -al
RUN npm run build && ls -al dist/spa


FROM golang:1.17-buster AS builder
ARG VERSION=0.0.10
LABEL stage=gobuilder

# 开始编译
ENV CGO_ENABLED 1
ENV GO111MODULE=on
ENV GOOS linux
#ENV GOPROXY https://goproxy.cn,direct

# 切换工作目录
WORKDIR /homelab/buildspace
COPY . .
# 把前端编译好的文件 copy 过来
COPY --from=frontBuilder /usr/src/app/dist/spa /homelab/buildspace/frontend/dist/spa


# 执行编译，-o 指定保存位置和程序编译名称
RUN --mount=type=secret,id=BASEKEY \
      --mount=type=secret,id=AESKEY16 \
      --mount=type=secret,id=AESIV16 \
    export BASEKEY=$(cat /run/secrets/BASEKEY) && \
      export AESKEY16=$(cat /run/secrets/AESKEY16) && \
      export AESIV16=$(cat /run/secrets/AESIV16) && \
    cd ./cmd/chinesesubfinder && \
    go build -ldflags="-s -w --extldflags '-static -fpic' -X main.AppVersion=${VERSION} -X main.BaseKey=$BASEKEY -X main.AESKey16=$AESKEY16 -X main.AESIv16=$AESIV16" -o /app/chinesesubfinder

# 运行时环境
FROM lsiobase/ubuntu:bionic

ENV TZ=Asia/Shanghai PERMS=true \
    PUID=1026 PGID=100

RUN  ln -s /root/.cache/rod/browser/$(ls /root/.cache/rod/browser)/chrome-linux/chrome /usr/bin/chrome && \
    # sed -i "s@http://archive.ubuntu.com@http://mirrors.aliyun.com@g" /etc/apt/sources.list && rm -Rf /var/lib/apt/lists/* && \
    apt-get update && \
    apt-get install --no-install-recommends -y \
    yasm ffmpeg \
    # C、C++ 支持库
    libgcc-6-dev libstdc++6 \
    # chromium dependencies
    libnss3 \
    libxss1 \
    libasound2 \
    libxtst6 \
    libgtk-3-0 \
    libgbm1 \
    # fonts
    fonts-liberation fonts-noto-color-emoji fonts-noto-cjk \
    # processs reaper
    dumb-init \
    # headful mode support, for example: $ xvfb-run chromium-browser --remote-debugging-port=9222
    xvfb \
    xorg gtk2-engines-pixbuf \
    dbus-x11 xfonts-base xfonts-100dpi xfonts-75dpi xfonts-cyrillic xfonts-scalable \
    imagemagick x11-apps \
    # 通用
    ca-certificates \
    wget \
    # cleanup
    && apt-get clean \
    && rm -rf \
       /tmp/* \
       /var/lib/apt/lists/* \
       /var/tmp/*

COPY Docker/root/ /
# 主程序
COPY --from=builder /app/chinesesubfinder /app/chinesesubfinder

VOLUME /config /media

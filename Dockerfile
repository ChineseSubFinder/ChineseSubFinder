FROM golang:1.17-buster AS builder
ARG VERSION=0.0.10
LABEL stage=gobuilder

# 开始编译
ENV CGO_ENABLED 1
ENV GO111MODULE=on
ENV GOOS linux
ENV GOPROXY https://goproxy.cn,direct

# 切换工作目录
WORKDIR /homelab/buildspace
COPY . .
# 执行编译，-o 指定保存位置和程序编译名称
RUN cd ./cmd/chinesesubfinder \
    && go build -ldflags="-s -w -X main.AppVersion=${VERSION}" -o /app/chinesesubfinder

# 运行时环境
FROM lsiobase/ubuntu:bionic

ENV TZ=Asia/Shanghai \
    PUID=1026 PGID=100

RUN ln -s /root/.cache/rod/chromium-856583/chrome-linux/chrome /usr/bin/chrome && \
    # sed -i "s@http://archive.ubuntu.com@http://mirrors.aliyun.com@g" /etc/apt/sources.list && rm -Rf /var/lib/apt/lists/* && \
    apt-get update && \
    apt-get install --no-install-recommends -y \
    yasm ffmpeg \
    # C、C++ 支持库
    libgcc-6-dev libstdc++6 \
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
# 配置文件
COPY --from=builder /homelab/buildspace/config.yaml.sample /app/config.yaml

VOLUME /config /media
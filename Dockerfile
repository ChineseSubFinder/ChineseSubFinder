FROM lsiobase/ubuntu:bionic AS builder

LABEL stage=gobuilder

# 设置环境变量，所有操作都是非交互式的
ENV DEBIAN_FRONTEND=noninteractive
ENV GO_USER=golang
ENV GO_LOG_DIR=/var/log/golang
# 这里的GOPATH路径是挂载的birdTracker项目的目录
ENV GOPATH=/home/golang/birdTracker
ENV PATH=$GOPATH/bin:/usr/local/go/bin:$PATH
ENV GOLANG_VERSION=1.15.2
ENV GOLANG_DOWNLOAD_URL https://golang.org/dl/go$GOLANG_VERSION.linux-amd64.tar.gz
# 替换 sources.list 的配置文件，并复制配置文件到对应目录下面。
# 这里使用的AWS国内的源，也可以替换成其他的源（例如：阿里云的源）
#COPY sources.list /etc/apt/sources.list
# 安装基础工具
RUN apt-get clean
RUN rm -rf /var/lib/apt/lists/*
RUN apt-get update
RUN apt-get install -y vim wget curl git
# 使用apt方式安装golang
RUN apt-get -y install golang
# 下载并安装golang
RUN curl -fsSL "$GOLANG_DOWNLOAD_URL" -o golang.tar.gz \
	&& echo "$GOLANG_DOWNLOAD_SHA256  golang.tar.gz" | sha256sum -c - \
	&& tar -C /usr/local -xzf golang.tar.gz \
	&& rm golang.tar.gz
# 创建用户和创建目录
RUN set -x && useradd $GO_USER && mkdir -p $GO_LOG_DIR $GOPATH && chown $GO_USER:$GO_USER $GO_LOG_DIR $GOPATH
WORKDIR $GOPATH

# 开始编译
ENV CGO_ENABLED 1
ENV GO111MODULE=on
ENV GOOS linux
ENV GOPROXY https://goproxy.cn,direct

# 切换工作目录
WORKDIR /homelab/buildspace

COPY . .
# 执行编译，-o 指定保存位置和程序编译名称
RUN go build -ldflags="-s -w" -o /app/chinesesubfinder

# 运行时环境
FROM lsiobase/ubuntu:bionic

ENV TZ=Asia/Shanghai \
    PUID=1026 PGID=100

RUN ln -s /root/.cache/rod/chromium-856583/chrome-linux/chrome /usr/bin/chrome && \
    # sed -i "s@http://archive.ubuntu.com@http://mirrors.aliyun.com@g" /etc/apt/sources.list && rm -Rf /var/lib/apt/lists/* && \
    apt-get update && \
    apt-get install --no-install-recommends -y \
    # C、C++ 支持库
    libgcc-6-dev libstdc++6 \
    # chromium dependencies
    libnss3 \
    libxss1 \
    libasound2 \
    libxtst6 \
    libgtk-3-0 \
    libgbm1 \
    ca-certificates \
    wget \
    # fonts
    fonts-liberation fonts-noto-color-emoji fonts-noto-cjk \
    # processs reaper
    dumb-init \
    # headful mode support, for example: $ xvfb-run chromium-browser --remote-debugging-port=9222
    xvfb \
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
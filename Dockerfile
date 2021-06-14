FROM golang:1.15-buster AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED 1
ENV GO111MODULE=on
ENV GOOS linux
ENV GOPROXY https://goproxy.cn,direct

# 切换工作目录
WORKDIR /homelab/buildspace

COPY . .
# 执行编译，-o 指定保存位置和程序编译名称
RUN go build -ldflags="-s -w" -o /app/chinesesubfinder

FROM ubuntu:bionic

RUN ln -s /root/.cache/rod/chromium-869685/chrome-linux/chrome /usr/bin/chrome && \
    sed -i "s@http://deb.debian.org@http://mirrors.aliyun.com@g" /etc/apt/sources.list && rm -Rf /var/lib/apt/lists/* && \
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
    # fonts
    fonts-liberation fonts-noto-color-emoji fonts-noto-cjk \
    # timezone
    tzdata \
    # processs reaper
    dumb-init \
    # headful mode support, for example: $ xvfb-run chromium-browser --remote-debugging-port=9222
    xvfb \
    # cleanup
    && rm -rf /var/lib/apt/lists/*

ENV TZ Asia/Shanghai

WORKDIR /app
# 主程序
COPY --from=builder /app/chinesesubfinder /app/chinesesubfinder
# 配置文件
COPY --from=builder /homelab/buildspace/config.yaml.sample /app/config.yaml
RUN chmod -R 777 /app
EXPOSE 1200

ENTRYPOINT ["/app/chinesesubfinder"]
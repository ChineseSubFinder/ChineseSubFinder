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
FROM jrottenberg/ffmpeg:4.4-alpine

# Add s6-overlay
ENV S6_OVERLAY_VERSION=v1.22.1.0 \
    GO_DNSMASQ_VERSION=1.0.7

RUN apk add --update --no-cache bind-tools curl libcap && \
    curl -sSL https://github.com/just-containers/s6-overlay/releases/download/${S6_OVERLAY_VERSION}/s6-overlay-amd64.tar.gz \
    | tar xfz - -C /

ENV TZ=Asia/Shanghai \
    PUID=1026 PGID=100

RUN apk update --no-cache \
   && apk add --no-cache ca-certificates tzdata libc6-compat libgcc libstdc++ wget

COPY Docker/root/ /
# 主程序
COPY --from=builder /app/chinesesubfinder /app/chinesesubfinder
# 配置文件
COPY --from=builder /homelab/buildspace/config.yaml.sample /app/config.yaml

VOLUME /config /media

CMD [""]
ENTRYPOINT [""]
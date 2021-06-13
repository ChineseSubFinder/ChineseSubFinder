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

FROM alpine:latest

RUN set -eux && sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories \
    && apk update --no-cache \
    && apk add --no-cache ca-certificates tzdata libc6-compat libgcc libstdc++
ENV TZ Asia/Shanghai

WORKDIR /app
# 主程序
COPY --from=builder /app/chinesesubfinder /app/chinesesubfinder
# 配置文件
COPY --from=builder /homelab/buildspace/config.yaml.sample /app/config.yaml
RUN chmod -R 777 /app
EXPOSE 1200

ENTRYPOINT ["/app/chinesesubfinder"]
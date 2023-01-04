#!/bin/bash

set -e

export CGO_ENABLED=1
export GO111MODULE=on
export GOOS=linux
export MUSL=/musl

## 准备必要的环境
[[ ! -d ~/go/src ]] && mkdir -p ~/go/src
[[ ! -d ~/go/bin ]] && mkdir -p ~/go/bin

## 下载源码并下载go mod
cd ~/go/src
curl -sSL https://github.com/ChineseSubFinder/ChineseSubFinder/archive/refs/tags/${VERSION}.tar.gz | tar xvz --strip-components 1
npm --prefix frontend ci
npm --prefix frontend run build
go mod tidy

## 编译共用函数
cross_make() {
    export PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
    if [[ -n ${CROSS_NAME} ]]; then
        export CPLUS_VERSION=$(${MUSL}/${CROSS_NAME}-cross/bin/${CROSS_NAME}-g++ --version | grep -oE '\d+\.\d+\.\d+' | head -1)
        export PATH=${MUSL}/${CROSS_NAME}-cross/bin:${MUSL}/${CROSS_NAME}-cross/${CROSS_NAME}/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
        export C_INCLUDE_PATH=${MUSL}/${CROSS_NAME}-cross/${CROSS_NAME}/include
        export CPLUS_INCLUDE_PATH=${MUSL}/${CROSS_NAME}-cross/${CROSS_NAME}/include/c++/${CPLUS_VERSION}
        export LIBRARY_PATH=${MUSL}/${CROSS_NAME}-cross/${CROSS_NAME}/lib
        export CC=${CROSS_NAME}-gcc
        export CXX=${CROSS_NAME}-g++
        export AR=${CROSS_NAME}-ar
    fi
    echo "[$(date +'%H:%M:%S')] 开始编译 ${GOARCH} 平台..."
    go build \
        -ldflags="-s -w --extldflags '-static -fpic' -X main.AppVersion=${VERSION} -X main.LiteMode=true -X main.BaseKey=${BASEKEY} -X main.AESKey16=${AESKEY16} -X main.AESIv16=${AESIV16}" \
        -o ~/go/out/${GOARCH}/chinesesubfinder \
        ./cmd/chinesesubfinder
    if [[ -n ${CROSS_NAME} ]]; then
        unset -v CPLUS_VERSION PATH C_INCLUDE_PATH CPLUS_INCLUDE_PATH LIBRARY_PATH CC CXX AR
    fi
    export PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
}

## 开始交叉编译
arches=( amd64 386 arm64 arm )
crosses=( "" i686-linux-musl aarch64-linux-musl armv7l-linux-musleabihf )
for ((i=0; i<${#arches[@]}; i++)); do
    export GOARCH=${arches[i]}
    export CROSS_NAME=${crosses[i]}
    [[ ${GOARCH} == amd64 ]] && export GOAMD64=v1
    [[ ${GOARCH} == arm ]] && export GOARM=7
    cross_make
    unset -v GOARCH CROSS_NAME GOAMD64 GOARM
done

## 列出文件
ls -lR ~/go/out
exit 0

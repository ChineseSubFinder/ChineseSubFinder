#!/bin/bash

## 重设权限
chown -R "${PUID}:${PGID}" /config
if [[ ${PERMS} == true ]]; then
    echo "已设置 PERMS=true，重设 '/media' 目录权限为 ${PUID}:${PGID} 所有..."
    chown -R ${PUID}:${PGID} /media
fi

## 兼容旧的缓存目录
if [[ -d /app/cache ]]; then
    echo "检测到映射了 '/app/cache'，创建软连接 '/config/cache' -> '/app/cache'"
    chown -R ${PUID}:${PGID} /app
    if [[ -L /config/cache && $(readlink -f /config/cache) != /app/cache ]]; then
        rm -rf /config/cache &>/dev/null
    fi
    if [[ ! -e /config/cache ]]; then
        gosu ${PUID}:${PGID} ln -sf /app/cache /config/cache
    fi
else
    if [[ -L /config/cache ]]; then
        echo "检测到 '/config/cache' 指向了不存在的目录 '/app/cache'，删除之，如想保留缓存，请将旧的 'cache' 目录移动到 '/config' 路径下..."
        rm -rf /config/cache &>/dev/null
    fi
fi

## 启动
umask ${UMASK}
cd /config
Xvfb -ac ${DISPLAY} -screen 0 1280x1024x16 &
gosu ${PUID}:${PGID} chinesesubfinder

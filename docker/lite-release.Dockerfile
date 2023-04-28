FROM alpine
ENV TZ=Asia/Shanghai \
    PERMS=true \
    PUID=1026 \
    PGID=100 \
    UMASK=022 \
    PS1="\u@\h:\w \$ "
ARG TARGETARCH
RUN apk add --no-cache \
       bash \
       ffmpeg \
       ca-certificates \
       tini \
       su-exec \
       tzdata \
    && ln -sf /usr/share/zoneinfo/${TZ} /etc/localtime \
    && echo "${TZ}" > /etc/timezone \
    && rm -rf /tmp/* /var/cache/apk/*
COPY go/out/${TARGETARCH}/chinesesubfinder /usr/bin/chinesesubfinder
COPY lite-entrypoint.sh /usr/bin/entrypoint.sh
VOLUME ["/config", "/media"]
WORKDIR /config
EXPOSE 19035
ENTRYPOINT ["tini", "entrypoint.sh"]

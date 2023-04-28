FROM allanpk716/chinesesubfinder-base:latest
ARG VERSION
ENV TZ=Asia/Shanghai \
    PERMS=true \
    PUID=1026 \
    PGID=100 \
    UMASK=022 \
    DISPLAY=:99 \
    PS1="\u@\h:\w \$ "
RUN cd /tmp \
    && arch=$(uname -m | sed -e 's|aarch64|arm64|' -e 's|armv7l|arm|') \
    && wget -q --no-check-certificate https://github.com/ChineseSubFinder/ChineseSubFinder/releases/download/${VERSION}/chinesesubfinder-${VERSION#*v}-Linux-${arch}.tar.gz \
    && tar xvf chinesesubfinder-${VERSION#*v}-Linux-${arch}.tar.gz \
    && mv chinesesubfinder /usr/local/bin \
    && chmod +x /usr/local/bin/chinesesubfinder \
    && ln -sf /usr/share/zoneinfo/${TZ} /etc/localtime \
    && echo "${TZ}" > /etc/timezone \
    && rm -rf /tmp/* \
EXPOSE 19035
COPY full-rootfs /
ENTRYPOINT ["/init"]
WORKDIR /config
VOLUME ["/config", "/media"]


FROM debian
ARG DEBIAN_FRONTEND=noninteractive
RUN apt-get update \
    && apt-get install --no-install-recommends -y \
       ca-certificates \
       chromium \
       dbus-x11 \
       dumb-init \
       ffmpeg \
       fonts-liberation \
       fonts-noto-cjk \
       fonts-noto-color-emoji \
       gosu \
       gtk2-engines-pixbuf \
       imagemagick \
       libasound2 \
       libgbm1 \
       libgcc-9-dev \
       libgtk-3-0 \
       libnss3 \
       libstdc++6 \
       libxss1 \
       libxtst6 \
       tini \
       tzdata \
       wget \
       x11-apps \
       xfonts-100dpi \
       xfonts-75dpi \
       xfonts-base \
       xfonts-cyrillic \
       xfonts-scalable \
       xorg \
       xvfb \
       yasm \
    && ln -s /usr/bin/chromium /usr/bin/chrome \
    && apt-get clean \
    && rm -rf \
       /tmp/* \
       /var/lib/apt/lists/* \
       /var/tmp/*


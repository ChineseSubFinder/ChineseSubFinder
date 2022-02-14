FROM library/node:16-alpine as builder

USER root
RUN mkdir -p /usr/src/app
WORKDIR /usr/src/app
add ./package.json /usr/src/app
add ./package-lock.json /usr/src/app
add ./.npmrc /usr/src/app
RUN npm ci
ADD . /usr/src/app
RUN ls -al
RUN npm run build && ls -al dist/spa

FROM library/nginx:alpine
COPY --from=builder /usr/src/app/dist/spa /usr/share/nginx/html
COPY nginx.conf /etc/nginx/

EXPOSE 80
EXPOSE 443

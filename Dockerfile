#FROM node:11.10.1-alpine
#ARG MAKE_ENV_ARGV
#ARG ALPINE_MIRROR_HOST=mirrors.tuna.tsinghua.edu.cn
#
#COPY ./ /app/
#RUN set -xe;\
#    sed -Ei "s/dl-cdn\.alpinelinux\.org/"$ALPINE_MIRROR_HOST"/g" /etc/apk/repositories;\
#    mkdir /apk-cache;\
#    apk update --cache-dir /apk-cache;\
#    apk add -t build-deps make;\
#    npm config set registry $NPM_REGISTRY;\
#    cd /app/;\
#    make bin/dashboard $MAKE_ENV_ARGV;\
#    apk del build-deps;\
#    rm -rf /apk-cache;

# Backend
FROM golang:1.12-alpine

ARG MAKE_ENV_ARGV
ARG ALPINE_MIRROR_HOST=mirrors.ustc.edu.cn
ARG NPM_REGISTRY=http://registry.npm.taobao.org

COPY ./ /app/
RUN set -xe;\
    sed -Ei "s/dl-cdn\.alpinelinux\.org/"$ALPINE_MIRROR_HOST"/g" /etc/apk/repositories;\
    mkdir /apk-cache;\
    apk update --cache-dir /apk-cache;\
    apk add -t build-deps make git gcc g++ nodejs nodejs-npm python;\
    apk add bash vim;\
    npm set strict-ssl false;\
    npm config set registry $NPM_REGISTRY;\
    cd /app/;\
    make bin/dashboard;\
    make bin/wing $MAKE_ENV_ARGV SKIP_FE_BUILD=1;\
    apk del build-deps;\
    cp bin/wing /bin/wing;\
    rm -rf /apk-cache /app;


ENTRYPOINT ["/bin/wing"]
CMD ["-config=/etc/wing/config.yml"]

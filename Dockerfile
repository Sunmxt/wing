FROM golang:1.12-alpine

COPY ./ /app/
ARG MAKE_ENV_ARGV

RUN set -xe;\
    sed -Ei "s/dl-cdn\.alpinelinux\.org/mirrors.tuna.tsinghua.edu.cn/g" /etc/apk/repositories;\
    mkdir /apk-cache;\
    apk update --cache-dir /apk-cache;\
    apk add -t build-deps make git gcc g++; \
    cd /app/;\
    make $MAKE_ENV_ARGV;\
    apk del build-deps;\
    cp bin/wing /bin/wing;\
    rm -rf /apk-cache /app;


ENTRYPOINT ["/bin/wing"]
CMD [""]

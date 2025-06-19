FROM golang:1.23.10-alpine3.21
ARG GIT_TAG=master
ENV CGO_CFLAGS="-D_LARGEFILE64_SOURCE"
RUN apk update && apk add git cmake build-base m4
RUN git clone --branch ${GIT_TAG} --recursive https://github.com/aergoio/aergo.git \
    && cd aergo \
    && make aergosvr polaris colaris aergocli aergoluac brick


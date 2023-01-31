FROM golang:1.19.5-alpine3.17
ARG GIT_TAG=master
RUN apk update && apk add git cmake build-base m4
RUN git clone --branch ${GIT_TAG} --recursive https://github.com/aergoio/aergo.git \
    && cd aergo \
    && make aergosvr polaris colaris aergocli aergoluac brick


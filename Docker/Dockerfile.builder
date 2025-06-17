FROM golang:1.13.5-alpine3.11
ARG GIT_TAG=master
RUN apk update && apk add git cmake build-base m4
RUN git clone --branch ${GIT_TAG} --recursive https://github.com/aergoio/aergo.git \
    && cd aergo \
    && make aergosvr polaris colaris aergocli aergoluac brick


FROM golang:1.19.0-bullseye as builder
RUN apt-get -y update && apt-get -y install build-essential git cmake binutils m4 file
RUN git clone --branch ${GIT_TAG} --recursive https://github.com/aergoio/aergo.git \
    && cd aergo \
    && make aergosvr polaris colaris aergocli aergoluac brick

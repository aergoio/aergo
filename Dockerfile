FROM golang:alpine as builder
RUN apk update && apk add git glide cmake build-base m4
ENV GOPATH $HOME/go
ARG GIT_TAG
RUN go get -d github.com/aergoio/aergo
WORKDIR ${GOPATH}/src/github.com/aergoio/aergo
RUN git checkout --detach ${GIT_TAG} && git submodule init && git submodule update && cmake .
RUN make aergosvr

FROM alpine:3.8
RUN apk add libgcc
COPY --from=builder $HOME/go/src/github.com/aergoio/aergo/bin/aergosvr /usr/local/bin/
COPY --from=builder $HOME/go/src/github.com/aergoio/aergo/libtool/lib/* /usr/local/lib/

ADD node/testnet.toml /aergo/testnet.toml
ADD node/testnet.toml /root/.aergo/config.toml
ADD node/mainnet.toml /aergo/mainnet.toml
ADD node/local.toml /aergo/local.toml
ADD node/testmode.toml /aergo/testmode.toml
ADD node/arglog.toml /aergo/arglog.toml
ENV LD_LIBRARY_PATH="/usr/local/lib:${LD_LIBRARY_PATH}"

WORKDIR /aergo/
CMD ["aergosvr", "--config", "/aergo/mainnet.toml"]
EXPOSE 7845 7846 6060

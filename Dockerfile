FROM golang:1.12.5-alpine3.9 as builder
RUN apk update && apk add git cmake build-base m4
COPY . aergo
RUN cd aergo && make aergosvr

FROM alpine:3.9
RUN apk add libgcc
COPY --from=builder /go/aergo/bin/aergosvr /usr/local/bin/
COPY --from=builder /go/aergo/libtool/lib/* /usr/local/lib/
COPY --from=builder /go/aergo/Docker/conf/* /aergo/
ENV LD_LIBRARY_PATH="/usr/local/lib:${LD_LIBRARY_PATH}"
WORKDIR /aergo/
CMD ["aergosvr", "--home", "/aergo"]
EXPOSE 7845 7846 6060

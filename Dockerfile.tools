FROM golang:1.12.5-alpine3.9 as builder
RUN apk update && apk add git cmake build-base m4
COPY . aergo
RUN cd aergo && make aergocli aergoluac brick

FROM alpine:3.9
RUN apk add libgcc
COPY --from=builder /go/aergo/bin/* /usr/local/bin/
COPY --from=builder /go/aergo/cmd/brick/arglog.toml /tools/arglog.toml
COPY --from=builder /go/aergo/libtool/lib/* /usr/local/lib/
ENV LD_LIBRARY_PATH="/usr/local/lib:${LD_LIBRARY_PATH}"
WORKDIR /tools/
CMD ["aergocli"]

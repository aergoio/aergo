FROM golang:1.12.5-alpine3.9 as builder
RUN apk update && apk add git cmake build-base m4
COPY . aergo
RUN cd aergo && make polaris colaris

FROM alpine:3.9
RUN apk add libgcc
COPY --from=builder go/aergo/bin/?olaris /usr/local/bin/
WORKDIR /tools/
CMD ["polaris"]

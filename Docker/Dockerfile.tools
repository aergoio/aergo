FROM alpine:3.9
RUN apk add libgcc
WORKDIR /tools/
COPY bin/aergocli bin/aergoluac bin/brick /usr/local/bin/
COPY bin/brick-arglog.toml arglog.toml
COPY lib/* /usr/local/lib/
ENV LD_LIBRARY_PATH="/usr/local/lib:${LD_LIBRARY_PATH}"
CMD ["aergocli"]

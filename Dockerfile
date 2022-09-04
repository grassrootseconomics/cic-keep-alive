FROM debian:11-slim

RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /cic-keep-alive

COPY cic-keep-alive .
COPY config.toml .

CMD ["./cic-keep-alive"]
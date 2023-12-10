FROM golang:1.21.4 as builder


RUN set -eux; \
    apt-get update \
    && apt-get install -y --no-install-recommends \
    ca-certificates \
    make \
    vim \
    unzip \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
ENV GO111MODULE=on

COPY go.mod go.sum ./
RUN go clean --modcache
RUN go mod download

COPY . .
ENTRYPOINT ["/usr/bin/make", "-C", "migrations"]
CMD ["bash"]
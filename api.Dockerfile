FROM golang:1.21.4 as builder

ENV PATH="$GOPATH/bin:$PATH"

WORKDIR /opt/sample
COPY . .

RUN go clean --modcache
RUN go mod download

# use local
RUN go install github.com/pilu/fresh@latest

EXPOSE 8080

CMD ["/go/bin/fresh"]

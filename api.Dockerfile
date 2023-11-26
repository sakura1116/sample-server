FROM golang:1.21.4 as base

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

FROM base as development
RUN go install github.com/pilu/fresh@latest
EXPOSE 8080
CMD ["fresh"]

FROM base as builder
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM alpine:latest as production
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/app .
EXPOSE 8080
CMD ["./app"]
FROM golang:1.24-alpine3.20 as builder
RUN mkdir /app
ADD . /app
WORKDIR /app
RUN go clean --modcache &&\
    go mod download &&\
    CGO_ENABLED=0 GOOS=linux go build -o cmd/pvz/bin/main ./cmd/pvz/

FROM alpine:latest
WORKDIR /pvz
COPY /.env .
COPY /config/config.yaml ./config/
COPY --from=builder /app/cmd/pvz/bin/main .
CMD ["./main"]
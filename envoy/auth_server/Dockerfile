FROM golang:1.23-alpine3.19 AS builder
RUN apk add --no-cache git

WORKDIR /go/src/github.com/buildwithgrove/auth-server
COPY . .
RUN apk add --no-cache make build-base
RUN go build -tags auth_server -o /go/bin/auth-server ./main.go

FROM alpine:3.19
WORKDIR /app

ARG IMAGE_TAG
ENV IMAGE_TAG=${IMAGE_TAG}
ENV CONFIG_PATH=/etc/auth_server.yaml

COPY --from=builder /go/bin/auth-server ./

CMD ["./auth-server"]

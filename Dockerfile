FROM golang:1.23-alpine3.19 AS builder
RUN apk add --no-cache git

WORKDIR /go/src/github.com/buildwithgrove/path
COPY . .
RUN apk add --no-cache make build-base
RUN go build -o /go/bin/path ./cmd

FROM alpine:3.19
WORKDIR /app

ARG IMAGE_TAG
ENV IMAGE_TAG=${IMAGE_TAG}

COPY --from=builder /go/bin/path ./

CMD ["./path"]

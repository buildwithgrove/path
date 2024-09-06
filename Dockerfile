FROM golang:1.22.3-alpine3.19 AS builder
RUN apk add --no-cache git

WORKDIR /go/src/github.com/pokt-foundation/portal-middleware
COPY . .
RUN apk add --no-cache make build-base
RUN go build -o /go/bin/path ./cmd/main.go

FROM alpine:3.19
WORKDIR /app

ARG IMAGE_TAG
ENV IMAGE_TAG=${IMAGE_TAG}

COPY --from=builder /go/bin/path ./

CMD ["./path"]

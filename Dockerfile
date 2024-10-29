FROM golang:1.23-alpine3.19
RUN apk add git make build-base

WORKDIR /app
COPY . .
RUN go build -o /go/bin/path ./cmd

ARG IMAGE_TAG
ENV IMAGE_TAG=${IMAGE_TAG}

CMD ["/go/bin/path"]

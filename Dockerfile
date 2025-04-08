FROM golang:1.23-alpine3.19 AS builder
RUN apk add --no-cache git make build-base

WORKDIR /go/src/github.com/buildwithgrove/path

# Copy only go.mod and go.sum first to cache dependency installation
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the files
COPY . .

# Build the application
RUN go build -o /go/bin/path ./cmd

FROM alpine:3.19
WORKDIR /app

ARG IMAGE_TAG
ENV IMAGE_TAG=${IMAGE_TAG}

# Create config directory
RUN mkdir -p /app/config

COPY --from=builder /go/bin/path ./

CMD ["./path"]
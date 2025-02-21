FROM golang:1.23-alpine3.19 AS builder

# Install all build dependencies in one layer
RUN apk add --no-cache git make build-base

WORKDIR /go/src/github.com/buildwithgrove/path/envoy/auth_server

# Copy and download dependencies first to leverage caching
COPY go.mod go.sum ./
RUN go mod download

# Copy rest of the code
COPY . .

# Build the application
RUN go build -o /go/bin/auth-server ./main.go

FROM alpine:3.19
WORKDIR /app

ARG IMAGE_TAG
ENV IMAGE_TAG=${IMAGE_TAG}
ENV CONFIG_PATH=/app/config/.config.yaml

# Create config directory
RUN mkdir -p /app/config

COPY --from=builder /go/bin/auth-server ./

CMD ["./auth-server"]
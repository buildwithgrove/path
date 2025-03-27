FROM --platform=$BUILDPLATFORM golang:1.23-alpine3.19 AS builder
RUN apk add --no-cache git make build-base

WORKDIR /go/src/github.com/buildwithgrove/path

# Copy only go.mod and go.sum first to cache dependency installation
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the files
COPY . .

# Build the application with platform-specific optimizations
ARG TARGETPLATFORM
ARG BUILDPLATFORM
RUN GOOS=$(echo $TARGETPLATFORM | cut -d/ -f1) \
    GOARCH=$(echo $TARGETPLATFORM | cut -d/ -f2) \
    CGO_ENABLED=0 \
    GOARM=7 \
    go build -ldflags="-w -s" -o /go/bin/path ./cmd

FROM --platform=$TARGETPLATFORM alpine:3.19
WORKDIR /app

ARG IMAGE_TAG
ENV IMAGE_TAG=${IMAGE_TAG}

# Create config directory
RUN mkdir -p /app/config

COPY --from=builder /go/bin/path ./

CMD ["./path"]
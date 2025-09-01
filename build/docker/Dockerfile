FROM golang:alpine AS builder

LABEL stage=gobuilder

# Set environment variables (fixed ENV format)
ENV CGO_ENABLED=0
ENV GOPROXY=https://goproxy.cn,direct
ENV TZ=Asia/Shanghai

# Use Aliyun mirror for Alpine
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# Install base tools
RUN apk update --no-cache && \
    apk add --no-cache tzdata

WORKDIR /build

# Copy dependency files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build application with optimizations
RUN go build -ldflags="-s -w" -o /app/app ./cmd/service1/main.go

# Use lightweight base image
FROM alpine:latest

# Copy timezone and certificates
COPY --from=builder /usr/share/zoneinfo/Asia/Shanghai /usr/share/zoneinfo/Asia/Shanghai
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

# Set timezone (fixed ENV format)
ENV TZ=Asia/Shanghai

WORKDIR /app

# Copy build artifacts
COPY --from=builder /app/app ./
COPY cmd/service1/application.yaml ./

# Expose ports (should match config file)
EXPOSE 8080 50051

# Run as non-root user for security
RUN adduser -D -g '' appuser && \
    chown -R appuser:appuser /app

USER appuser

# Start command
CMD ["./app"]
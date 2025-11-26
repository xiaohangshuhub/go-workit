FROM golang:alpine AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED=0 \
    GOPROXY=https://goproxy.cn,direct \
    TZ=Asia/Shanghai

# Use Aliyun mirror for Alpine
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apk/repositories \
    && apk add --no-cache tzdata

WORKDIR /build


# Copy source code
COPY . .

# Copy dependency files first for better caching
RUN go mod download

# Build application with optimizations
RUN go build  -tags=go_json -ldflags="-s -w" -o /app/app ./cmd/service1/main.go


# ---------------- Runtime stage ----------------
FROM alpine:latest

# Copy timezone and certificates
COPY --from=builder /usr/share/zoneinfo/Asia/Shanghai /usr/share/zoneinfo/Asia/Shanghai
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

ENV TZ=Asia/Shanghai

# Create app directory and user
WORKDIR /app
RUN adduser -D -g '' appuser

# Copy build artifacts and set ownership directly
COPY --from=builder --chown=appuser:appuser /app/app ./app
COPY --chown=appuser:appuser cmd/service1/application.yaml ./

EXPOSE 8080 50051

USER appuser

CMD ["./app"]
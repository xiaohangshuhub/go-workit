FROM golang:alpine AS builder

LABEL stage=gobuilder

# 设置环境变量
ENV CGO_ENABLED 0
ENV GOPROXY https://goproxy.cn,direct
ENV TZ Asia/Shanghai

# 修改 Alpine 源为阿里云源
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# 安装基础工具
RUN apk update --no-cache && apk add --no-cache tzdata

WORKDIR /build

# 首先复制依赖文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN go build -ldflags="-s -w" -o /app/service1 ./cmd/service1/main.go

# 使用轻量级基础镜像
FROM alpine:latest

# 复制时区和证书文件
COPY --from=builder /usr/share/zoneinfo/Asia/Shanghai /usr/share/zoneinfo/Asia/Shanghai
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

# 设置时区
ENV TZ Asia/Shanghai

WORKDIR /app

# 复制构建产物
COPY --from=builder /app/service1 ./
COPY cmd/service1/application.yaml ./

# 暴露端口(与配置文件中的端口保持一致)
EXPOSE 8080 50051

# 启动命令
CMD ["./service1"]
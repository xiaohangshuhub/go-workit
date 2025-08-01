# workit

workit 俚语,努力去做。

一个整合DDD(领域驱动设计)、 Echo框架、Zap日志、Fx依赖注入、Viper配置管理的轻量级、高扩展性的 Golang Web 应用快速开发模板，是模板不是框架!

> 🚀 帮助你快速构建清晰可扩展的 Golang 微服务 / API 应用。

---

# Branch

- main: 基于 Gin 框架源码
- echo: 基于 Echo 框架源码
- dev: 功能开发迭代
- cli-template:  基于Gin开发模板
- cli-echo:  基于Echo开发模板

# Features

- 🚀 模块化 WebHost 架构
- 🔥 依赖注入（DI）服务管理（基于 fx.Option）
- ⚙️ 灵活配置管理（Viper封装，多源支持）
- 🖋️ 高性能日志系统（Zap，支持 console 彩色和 file JSON输出）
- 🛡️ 支持中间件链路（自定义中间件注册）
- 📦 内置健康检查、静态文件服务、Swagger文档集成
- 🌐 支持环境区分（Debug/Release模式自动适配）
- 🏗️ 标准生命周期管理（构建 → 启动 → 关闭）

---

# Getting Started

## Installation

```bash
#  安装CLI
go install github.com/xiaohangshuhub/workit-cli/cmd/workit@latest
# 查看版本
workit -v
# 创建项目
workit new myapp 
```

## Hello World Example

```go
package main

import (
	"fmt"

<<<<<<< HEAD
	"github.com/labstack/echo/v4"
	"github.com/xiaohangshuhub/go-workit/pkg/cache"
	"github.com/xiaohangshuhub/go-workit/pkg/database"
	"github.com/xiaohangshuhub/go-workit/pkg/host"
	"github.com/xiaohangshuhub/go-workit/pkg/middleware"
=======
	_ "github.com/xiaohangshuhub/go-workit/api/service1/docs" // swagger 一定要有这行,指向你的文档地址
	"github.com/xiaohangshuhub/go-workit/internal/service1/grpcapi/hello"
	"github.com/xiaohangshuhub/go-workit/internal/service1/webapi"
	"github.com/xiaohangshuhub/go-workit/pkg/workit"
	"go.uber.org/fx"
>>>>>>> dev
	"go.uber.org/zap"
)

// 结构体
type HelloService struct {
	log *zap.Logger
}

// 构建服务
func NewHelloService(log *zap.Logger) *HelloService {
	return &HelloService{log: log}
}

func main() {

	// web应用构建器
	builder := workit.NewWebAppBuilder()

	// 配置构建器(注册即生效)
	builder.AddConfig(func(build workit.ConfigBuilder) {
		build.AddYamlFile("./config.yaml")
	})
	// 配置构建器读取数据
	var port = builder.Config().Get("server.port")
	fmt.Println("server port:", port)

	// 服务注册
	builder.AddServices(fx.Provide(NewHelloService))

	app, err := builder.Build()

	if err != nil {
		fmt.Printf("Failed to build application: %v\n", err)
		return
	}

	if app.Env.IsDevelopment {
		app.UseSwagger()
	}

	// 配置路由
	app.MapRoutes(webapi.Hello)

	// 配置路由
	app.MapRoutes(func(router *echo.Echo) {
		router.GET("/ping", func(c echo.Context) error {
			return c.JSON(200, map[string]string{
				"message": "hello world",
			})
		})
	})

	// 运行应用
	if err := app.Run(); err != nil {
		app.Logger().Error("Error running application", zap.Error(err))
	}
}

```

---

# Core Concepts

## Dependency Injection (依赖注入)

**Design Philosophy**

- 基于 Uber Fx 理念，通过 `fx.Option` 管理服务依赖
- Builder模式动态注册，支持应用启动时灵活装配服务
- 解耦组件间依赖关系，提升可测试性和可维护性

**How to Use**

注册依赖：

```go
builder.AddServices(
	fx.Provide(NewDatabase),
	fx.Provide(NewCache),
)
```

使用依赖：

```go
func NewHandler(db *Database, cache *Cache) *Handler {
	return &Handler{db: db, cache: cache}
}
```

---

## Configuration Management (配置管理)

**Design Philosophy**

- 基于 Viper 封装
- 支持 YAML、ENV环境变量、命令行多源加载
- 层级合并，适合开发、测试、生产环境
- 简化配置绑定，统一管理

**How to Use**

加载配置：

```go
builder.AddConfig(func(cfg host.ConfigBuilder) {
	_ = cfg.AddYamlFile("./config.yaml")
})
```

## Web Server Configuration

```go
server:
  port: 8080
  grpc_port: 50051

```

---

## Logging System (日志系统)

**Design Philosophy**

- 基于 Zap，极致性能
- Console 彩色输出（Dev模式）
- JSON结构化日志（Prod模式）
- 多目标输出：控制台 + 文件
- 自动切换输出格式，适配不同环境

**How to Use**

配置日志：

```go
log:
  level: info # 日志级别，可选值：debug, info, warn, error, fatal, panic
  filename: ./logs/app.log
  maxsize: 100    # 每个日志文件的最大尺寸(MB)
  maxbackups: 3   # 保留的旧日志文件最大数量 
  maxage: 7       # 保留的旧日志文件最大天数
  compress: true  # 是否压缩旧日志文件
  console: true   # 是否同时输出到控制台
```

日志输出示例：

```go
logger.Info("HTTP server starting...", zap.String("port", "8080"))
```

---

## WebApplicationBuilder (Web应用构建器)

**Design Philosophy**

- 参考 Builder 设计模式
- 统一应用生命周期管理
- 配置-服务-应用分层清晰
- 支持灵活扩展和插件化开发

**How to Use**

标准流程：

```go
builder := host.NewWebAppBuilder().
	AddConfig(...) .
	AddServices(...) .
	ConfigureWebServer(...)

app, err := builder.Build()
app.Run()
```

---

# Advanced Guide

- 中间件管理（UseMiddleware）
- 静态文件托管（UseStaticFiles）
- 健康检查（UseHealthCheck）
- Swagger集成（UseSwagger）
- 支持分组路由（gin）

---

# Deployment

- Release模式部署前，建议：
  - 修改 `config.yaml` 中 `gin.mode=release`
  - 关闭 console 日志，仅保存文件日志
  - 使用 `docker-compose` 或 `k8s` 管理服务
- 支持优雅关闭（待完善 graceful shutdown）

---

# Contribute

欢迎贡献代码、提出建议或者提交 PR！

---

# License

This project is licensed under the MIT License.

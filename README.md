# workit

workit 俚语,努力去做。

一个整合DDD(领域驱动设计)、 Gin框架、Zap日志、Fx依赖注入、Viper配置管理的轻量级、高扩展性的 Golang Web 应用快速开发模板，是模板不是框架!

> 🚀 帮助你快速构建清晰可扩展的 Golang 微服务 / API 应用。

---
# Branch
- main: 基于 Gin 框架
- echo: 基于 Echo 框架
- dev: 功能开发迭代
- cli: 开发模板

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
git get  git@github.com:lxhanghub/workit
```

## Hello World Example

```go
package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/lxhanghub/workit/pkg/cache"
	"github.com/lxhanghub/workit/pkg/database"
	"github.com/lxhanghub/workit/pkg/host"
	"github.com/lxhanghub/workit/pkg/middleware"
	"go.uber.org/zap"
	//_ "xxx/docs" // swagger 一定要有这行,指向你的文档地址
)

func main() {

	// 创建服务主机构建器
	builder := host.NewWebHostBuilder()

	// 配置应用配置,内置环境变量读取和命令行参数读取
	builder.ConfigureAppConfiguration(func(build host.ConfigBuilder) {
		build.AddYamlFile("../../configs/config.yaml")
	})

	// 配置依赖注入
	builder.ConfigureServices(database.PostgresModule())

	builder.ConfigureServices(cache.RedisModule())

	//配置请求中间件,支持跳过

	//构建应用
	app, err := builder.Build()

	if err != nil {
		fmt.Printf("Failed to build application: %v\n", err)
		return
	}

	app.UseMiddleware(middleware.NewAuthorizationMiddleware([]string{"/hello"}))

	//app.UseSwagger()

	// 配置路由
	app.MapRoutes(func(router *gin.Engine) {
		router.GET("/ping", func(c *gin.Context) {

			c.JSON(200, gin.H{"message": "hello world"})
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
builder.ConfigureServices(
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
builder.ConfigureAppConfiguration(func(cfg host.ConfigBuilder) {
	_ = cfg.AddYamlFile("./configs/config.yaml")
})
```

## Web Server Configuration

```go
builder.ConfigureWebServer(host.WebHostOptions{
	Service: host.ServiceOptions{Port: "8080"},
})
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
Log: host.LogOptions{
	Level:    "info",
	Console:  true,
	Filename: "./logs/app.log",
}
```

日志输出示例：

```go
logger.Info("HTTP server starting...", zap.String("port", "8080"))
```

---

## WebHostBuilder (Web应用宿主构建器)

**Design Philosophy**

- 参考 ASP.NET Core HostBuilder 模式
- 统一应用生命周期管理
- 配置-服务-应用分层清晰
- 支持灵活扩展和插件化开发

**How to Use**

标准流程：

```go
builder := host.NewWebHostBuilder().
	ConfigureAppConfiguration(...) .
	ConfigureServices(...) .
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

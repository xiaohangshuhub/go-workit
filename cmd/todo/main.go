package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/lxhanghub/newb/pkg/cache"
	"github.com/lxhanghub/newb/pkg/database"
	"github.com/lxhanghub/newb/pkg/host"
	"github.com/lxhanghub/newb/pkg/middleware"
	"go.uber.org/zap"
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
	builder.UseMiddleware(middleware.NewAuthorizationMiddleware([]string{"/hello"}))

	// 配置路由
	builder.MapGet("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "hello world",
		})
	})
	builder.MapGet("/hello", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "hello world",
		})
	})

	//构建应用
	app, err := builder.Build()

	if err != nil {
		fmt.Printf("Failed to build application: %v\n", err)
		return
	}

	// 运行应用
	if err := app.Run(); err != nil {
		app.Logger().Error("Error running application", zap.Error(err))
	}
}

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

	// 1、创建服务主机构建器
	builder := host.NewWebHostBuilder()

	// 2、配置中间件,依赖注入,应用配置等等
	// ......
	builder.ConfigureAppConfiguration(func(build host.ConfigBuilder) {
		build.AddYamlFile("../../configs/config.yaml")
	})

	builder.ConfigureServices(database.PostgresModule())

	builder.ConfigureServices(cache.RedisModule())

	builder.UseMiddleware(middleware.NewAuthorizationMiddleware([]string{"/hello"}))

	// 3、配置路由
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

	//4、构建应用
	app, err := builder.Build()

	if err != nil {
		fmt.Printf("Failed to build application: %v\n", err)
		return
	}

	// 5、运行应用
	if err := app.Run(); err != nil {
		app.Logger().Error("Error running application", zap.Error(err))
	}
}

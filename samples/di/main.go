// Package main API文档
//
// @title           我的服务 API
// @version         1.0
// @description     这是一个示例 API 文档
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description 输入格式: Bearer {token}
package main

import (
	"github.com/gin-gonic/gin"
	_ "github.com/xiaohangshuhub/go-workit/api/service1/docs" // swagger 一定要有这行,指向你的文档地址
	"github.com/xiaohangshuhub/go-workit/pkg/webapp"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type MyType struct {
	*zap.Logger
}

func NewMyType(logger *zap.Logger) *MyType {
	return &MyType{
		Logger: logger,
	}
}

func main() {
	// web应用构建器
	builder := webapp.NewBuilder()

	builder.AddServices(fx.Provide(NewMyType))

	// 构建Web应用
	app := builder.Build()

	// 配置路由
	app.MapRouter(func(router *gin.Engine, myType *MyType) {
		router.GET("/hello", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "Hello, World!",
			})
		})
	})

	// 运行应用
	app.Run()
}

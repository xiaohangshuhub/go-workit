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
	"github.com/minio/minio-go/v7"
	_ "github.com/xiaohangshu-dev/go-workit/api/service1/docs" // swagger 一定要有这行,指向你的文档地址

	"github.com/xiaohangshu-dev/go-workit/pkg/components/miniox"
	"github.com/xiaohangshu-dev/go-workit/pkg/webapp"
	"github.com/xiaohangshu-dev/go-workit/pkg/webapp/minioctx"
)

func main() {

	builder := webapp.NewBuilder()

	builder.AddMinioContext(func(opts *minioctx.Options) {
		opts.UseClient("default", func(cfg *miniox.Options) {
			cfg.Endpoint = "117.72.15.185:9000"
		})
	})

	app := builder.Build()

	app.MapRoute(func(router *gin.Engine, mc *minio.Client) {
		router.GET("/hello", func(c *gin.Context) {

			c.JSON(200, gin.H{
				"message": "Hello, World!",
			})
		})
	})

	app.Run()
}

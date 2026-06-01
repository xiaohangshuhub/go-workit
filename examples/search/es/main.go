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
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/gin-gonic/gin"
	_ "github.com/xiaohangshu-dev/go-workit/api/service1/docs" // swagger 一定要有这行,指向你的文档地址

	"github.com/xiaohangshu-dev/go-workit/pkg/components/elasticsearchx"
	"github.com/xiaohangshu-dev/go-workit/pkg/webapp"
	"github.com/xiaohangshu-dev/go-workit/pkg/webapp/esctx"
)

func main() {

	builder := webapp.NewBuilder()

	builder.AddEsContext(func(opts *esctx.Options) {
		opts.UseClient("default", func(cfg *elasticsearchx.Options) {
			cfg.Addresses = []string{"http://localhost:9200"}
		})
	})

	app := builder.Build()

	app.MapRoute(func(router *gin.Engine, es *elasticsearch.Client) {
		router.GET("/hello", func(c *gin.Context) {
			info, err := es.Info()
			if err != nil {
				c.JSON(500, gin.H{
					"error": err.Error(),
				})
				return
			}
			println(info.String())

			c.JSON(200, gin.H{
				"message": "Hello, World!",
			})
		})
	})

	app.Run()
}

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
	"github.com/olivere/elastic/v7"
	_ "github.com/xiaohangshu-dev/go-workit/api/service1/docs" // swagger 一定要有这行,指向你的文档地址

	"github.com/xiaohangshu-dev/go-workit/pkg/components/elasticx"
	"github.com/xiaohangshu-dev/go-workit/pkg/webapp"
	"github.com/xiaohangshu-dev/go-workit/pkg/webapp/elasticctx"
)

func main() {

	builder := webapp.NewBuilder()

	builder.AddElasticContext(func(opts *elasticctx.Options) {
		opts.UseClient("default", func(cfg *elasticx.Options) {
			cfg.Func = []elastic.ClientOptionFunc{
				elastic.SetURL("http://117.72.15.185:9200"),
				elastic.SetSniff(false),
			}
		})
	})

	app := builder.Build()

	app.MapRoute(func(router *gin.Engine, es *elastic.Client) {
		router.GET("/hello", func(c *gin.Context) {

			c.JSON(200, gin.H{
				"message": "Hello, World!",
			})
		})
	})

	app.Run()
}

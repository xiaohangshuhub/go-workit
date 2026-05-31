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
	_ "github.com/xiaohangshu-dev/go-workit/api/service1/docs" // swagger 一定要有这行,指向你的文档地址
	"github.com/xiaohangshu-dev/go-workit/pkg/db/mongox"
	"github.com/xiaohangshu-dev/go-workit/pkg/webapp/mongoctx"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/xiaohangshu-dev/go-workit/pkg/webapp"
)

func main() {

	builder := webapp.NewBuilder()

	builder.AddMongoContext(func(opts *mongoctx.Options) {
		opts.UseClient("default", func(cfg *mongox.Options) {
			cfg.ApplyURI("mongodb://localhost:27017")
		})
	})

	app := builder.Build()

	app.MapRoute(func(router *gin.Engine, mongo *mongo.Client) {
		router.GET("/hello", func(c *gin.Context) {
			// 检测是否连接成功
			err := mongo.Ping(c, nil)
			if err != nil {
				c.JSON(500, gin.H{
					"error": err.Error(),
				})
				return
			}

			// 连接成功, 打印数据库列表
			databases, err := mongo.ListDatabaseNames(c, bson.M{})
			if err != nil {
				c.JSON(500, gin.H{
					"error": err.Error(),
				})
				return
			}
			println("数据库列表:")
			for _, db := range databases {
				println(db)
			}

			c.JSON(200, gin.H{
				"message": "Hello, World!",
			})
		})
	})

	app.Run()
}

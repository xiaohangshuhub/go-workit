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
	"github.com/xiaohangshuhub/go-workit/pkg/database"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/dbcontext"

	"github.com/xiaohangshuhub/go-workit/pkg/webapp"
	"gorm.io/gorm"
)

func main() {
	// web应用构建器
	builder := webapp.NewBuilder()

	builder.AddDbContext(func(opts *dbcontext.Options) {
		opts.UsePostgresSQL("default", func(cfg *database.PostgresConfig) {
			cfg.PgSQLCfg.DSN = builder.Config.GetString("database.dsn")
		})

	})

	// 构建Web应用
	app := builder.Build()

	// 配置路由
	app.MapRouter(func(router *gin.Engine, orm *gorm.DB) {
		router.GET("/hello", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "Hello, World!",
			})
		})
	})

	// 运行应用
	app.Run()
}

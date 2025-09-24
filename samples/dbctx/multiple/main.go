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
	"github.com/xiaohangshuhub/go-workit/pkg/db"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/dbctx"

	"github.com/xiaohangshuhub/go-workit/pkg/webapp"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

type DBs struct {
	fx.In

	Other *gorm.DB `name:"other"`
}

func main() {

	builder := webapp.NewBuilder()

	builder.AddDbContext(func(opts *dbctx.Options) {

		opts.UseMySQL("default", func(cfg *db.MySQLConfigOptions) {
			cfg.MySQLCfg.DSN = builder.Config.GetString("database.dsn")

		})

		opts.UsePostgresSQL("other", func(cfg *db.PostgresConfigOptions) {
			cfg.PgSQLCfg.DSN = db.PostgresDefaultDns
		})

	})

	app := builder.Build()

	app.MapRoute(func(router *gin.Engine, orm *gorm.DB, db DBs) {
		router.GET("/hello", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "Hello, World!",
			})
		})
	})

	app.Run()
}

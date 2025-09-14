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
	"github.com/xiaohangshuhub/go-workit/pkg/app"
	"github.com/xiaohangshuhub/go-workit/pkg/database"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

type DBs struct {
	fx.In

	Other *gorm.DB `name:"other"`
}

func main() {
	// web应用构建器
	builder := webapp.NewBuilder()

	// 配置构建器(注册即生效)
	builder.AddConfig(func(build *app.ConfigOptions) {
		build.AddYamlFile("./application.yaml")
	})

	builder.AddDbContext(func(opts *webapp.DbContextOptions) {

		opts.UseMySQL("default", func(cfg *database.MySQLConfigOptions) {
			cfg.MySQLCfg.DSN = builder.Config.GetString("database.dsn")

		})

		opts.UsePostgresSQL("other", func(cfg *database.PostgresConfig) {
			cfg.PgSQLCfg.DSN = database.PostgresDefaultDns
		})

	})

	// 构建Web应用
	app := builder.Build()

	// 配置路由
	app.MapRouter(func(router *gin.Engine, orm *gorm.DB, db DBs) {
		router.GET("/hello", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "Hello, World!",
			})
		})
	})

	// 运行应用
	app.Run()
}

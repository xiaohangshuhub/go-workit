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
	"database/sql"

	"github.com/gin-gonic/gin"
	_ "github.com/xiaohangshu-dev/go-workit/api/service1/docs" // swagger 一定要有这行,指向你的文档地址
	"github.com/xiaohangshu-dev/go-workit/pkg/db/gormx/mysqlx"
	mysqldb "github.com/xiaohangshu-dev/go-workit/pkg/db/mysqlx"
	"github.com/xiaohangshu-dev/go-workit/pkg/webapp/dbctx"
	"github.com/xiaohangshu-dev/go-workit/pkg/webapp/gormctx"

	"github.com/xiaohangshu-dev/go-workit/pkg/webapp"
	"gorm.io/gorm"
)

func main() {

	builder := webapp.NewBuilder()

	builder.AddGormContext(func(opts *gormctx.Options) {

		opts.UseMySQL("default", func(cfg *mysqlx.Options) {
			cfg.MySQLCfg.DSN = builder.Config().GetString("database.dsn")

		})
	})

	builder.AddDbContext(func(opts *dbctx.Options) {
		opts.UseMySQL("default", func(cfg *mysqldb.Options) {
			cfg.DSN = builder.Config().GetString("database.dsn")

		})
	})

	app := builder.Build()

	app.MapRoute(func(router *gin.Engine, orm *gorm.DB, db *sql.DB) {
		router.GET("/hello", func(c *gin.Context) {
			err := db.Ping()
			if err != nil {
				c.JSON(500, gin.H{
					"error": err.Error(),
				})
				return
			}

			c.JSON(200, gin.H{
				"message": "Hello, World!",
			})
		})
	})

	app.Run()
}

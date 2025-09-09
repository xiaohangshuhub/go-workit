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
	"github.com/go-redis/redis/v8"
	_ "github.com/xiaohangshuhub/go-workit/api/service1/docs" // swagger 一定要有这行,指向你的文档地址
	"github.com/xiaohangshuhub/go-workit/pkg/cache"
	"github.com/xiaohangshuhub/go-workit/pkg/workit"
)

func main() {
	// web应用构建器
	builder := workit.NewWebAppBuilder()

	// 配置构建器(注册即生效)
	builder.AddConfig(func(build workit.ConfigBuilder) {
		build.AddYamlFile("./application.yaml")
	})

	builder.AddCacheContext(func(opts *workit.CacheContextOptions) {

		opts.UseRedis("default", func(cfg *cache.RedisConfigOptions) {
			cfg.Addr = builder.Config.GetString("redis.addr")
			cfg.Password = builder.Config.GetString("redis.password")
			cfg.DB = builder.Config.GetInt("redis.db")
			cfg.PoolSize = builder.Config.GetInt("redis.pool_size")
		})
	})

	// 构建Web应用
	app := builder.Build()

	// 配置路由
	app.MapRouter(func(router *gin.Engine, rc *redis.Client) {
		router.GET("/hello", func(c *gin.Context) {
			rc.Set(c, "hello", "Hello World", 0)

			c.JSON(200, gin.H{
				"message": rc.Get(c, "hello").Val(),
			})
		})
	})

	// 运行应用
	app.Run()
}

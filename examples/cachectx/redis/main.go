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
	_ "github.com/xiaohangshu-dev/go-workit/api/service1/docs" // swagger 一定要有这行,指向你的文档地址
	"github.com/xiaohangshu-dev/go-workit/internal/service1/webapi"

	r "github.com/xiaohangshu-dev/go-workit/pkg/cache/redis"
	"github.com/xiaohangshu-dev/go-workit/pkg/webapp"
	"github.com/xiaohangshu-dev/go-workit/pkg/webapp/cachectx"
)

func main() {

	builder := webapp.NewBuilder()

	builder.AddCacheContext(func(opts *cachectx.Options) {

		opts.UseRedis("default", func(cfg *r.Options) {
			cfg.Addr = builder.Config().GetString("redis.addr")
			cfg.Password = builder.Config().GetString("redis.password")
			cfg.DB = builder.Config().GetInt("redis.db")
			cfg.PoolSize = builder.Config().GetInt("redis.pool_size")
		})
	})

	app := builder.Build()

	app.MapRoute(webapi.Cache)

	// 运行应用
	app.Run()
}

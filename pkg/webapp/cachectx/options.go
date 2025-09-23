package cachectx

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/xiaohangshuhub/go-workit/pkg/app"

	r "github.com/xiaohangshuhub/go-workit/pkg/cache/redis"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Options struct {
	container []fx.Option         // 持有容器引用
	cacheMap  map[string]struct{} // 数据库实例名称集合
}

// NewOptions
func NewOptions() *Options {

	return &Options{
		container: make([]fx.Option, 0),
		cacheMap:  make(map[string]struct{}),
	}
}

func (c *Options) UseRedis(instanceName string, fn func(cfg *r.Options)) *Options {
	if instanceName == "" {
		// 默认单库，无 name
		instanceName = "default"
	}

	if _, ok := c.cacheMap[instanceName]; ok {
		panic("database instance name already exists")
	}

	cfg := &r.Options{
		Options: redis.Options{},
	}

	fn(cfg)

	if instanceName == "default" {
		// 单库，第一次注册 default，提供不带 name 的数据库
		c.container = append(c.container,
			fx.Provide(func(lc fx.Lifecycle, logger *zap.Logger, appCtx *app.AppContext) (*redis.Client, error) {
				return r.NewRedisClient(lc, cfg, logger, appCtx)
			}),
		)
	} else {
		// 多库，或显式传名字的数据库，使用 name 标签
		c.container = append(c.container,
			fx.Provide(
				fx.Annotate(
					func(lc fx.Lifecycle, logger *zap.Logger, appCtx context.Context) (*redis.Client, error) {
						return r.NewRedisClient(lc, cfg, logger, appCtx)
					},
					fx.ResultTags(`name:"`+instanceName+`"`),
				),
			),
		)
	}

	c.cacheMap[instanceName] = struct{}{}

	return c
}

func (o *Options) Container() []fx.Option {
	return o.container
}

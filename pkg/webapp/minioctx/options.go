package minioctx

import (
	"github.com/minio/minio-go/v7"
	"github.com/xiaohangshu-dev/go-workit/pkg/components/miniox"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Options struct {
	container []fx.Option         // 持有容器引用
	cacheMap  map[string]struct{} // Minio 实例实例名称集合
}

// NewOptions
func NewOptions() *Options {

	return &Options{
		container: make([]fx.Option, 0),
		cacheMap:  make(map[string]struct{}),
	}
}

// UseClient  使用Minio 作为缓存
func (c *Options) UseClient(instanceName string, fn func(*miniox.Options)) *Options {
	if instanceName == "" {
		// 默认单库，无 name
		instanceName = "default"
	}

	if _, ok := c.cacheMap[instanceName]; ok {
		panic("redis instance name already exists")
	}

	cfg := &miniox.Options{
		Options: minio.Options{},
	}

	fn(cfg)

	if instanceName == "default" {
		// 单库，第一次注册 default，提供不带 name 的数据库
		c.container = append(c.container,
			fx.Provide(func(lc fx.Lifecycle, logger *zap.Logger) *minio.Client {
				return miniox.NewClient(lc, cfg, logger)
			}),
		)
	} else {
		// 多库，或显式传名字的数据库，使用 name 标签
		c.container = append(c.container,
			fx.Provide(
				fx.Annotate(
					func(lc fx.Lifecycle, logger *zap.Logger) *minio.Client {
						return miniox.NewClient(lc, cfg, logger)
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

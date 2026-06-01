package kafkactx

import (
	"github.com/segmentio/kafka-go"
	"github.com/xiaohangshu-dev/go-workit/pkg/components/kafkax"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Options struct {
	container []fx.Option         // 持有容器引用
	readerMap map[string]struct{} // Kafka Reader 实例名称集合
	writerMap map[string]struct{} // Kafka Writer 实例名称集合
}

// NewOptions
func NewOptions() *Options {

	return &Options{
		container: make([]fx.Option, 0),
		readerMap: make(map[string]struct{}),
		writerMap: make(map[string]struct{}),
	}
}

// UseReaderClient  使用Kafka Reader 实例
func (c *Options) UseReaderClient(instanceName string, fn func(*kafkax.ReaderOptions)) *Options {
	if instanceName == "" {
		// 默认单库，无 name
		instanceName = "default"
	}

	if _, ok := c.readerMap[instanceName]; ok {
		panic("kafka reader instance name already exists")
	}

	cfg := &kafkax.ReaderOptions{
		ReaderConfig: kafka.ReaderConfig{},
	}

	fn(cfg)

	if instanceName == "default" {
		// 单库，第一次注册 default，提供不带 name 的数据库
		c.container = append(c.container,
			fx.Provide(func(lc fx.Lifecycle, logger *zap.Logger) *kafka.Reader {
				return kafkax.NewReader(lc, cfg, logger)
			}),
		)
	} else {
		// 多库，或显式传 name 的 reader，使用 name 标签
		c.container = append(c.container,
			fx.Provide(
				fx.Annotate(
					func(lc fx.Lifecycle, logger *zap.Logger) *kafka.Reader {
						return kafkax.NewReader(lc, cfg, logger)
					},
					fx.ResultTags(`name:"`+instanceName+`"`),
				),
			),
		)
	}

	c.readerMap[instanceName] = struct{}{}

	return c
}

// UseWriterClient  使用Kafka Writer 实例
func (c *Options) UseWriterClient(instanceName string, fn func(*kafkax.WriterOptions)) *Options {
	if instanceName == "" {
		// 默认单库，无 name
		instanceName = "default"
	}

	if _, ok := c.writerMap[instanceName]; ok {
		panic("kafka writer instance name already exists")
	}

	cfg := &kafkax.WriterOptions{
		WriterConfig: kafka.WriterConfig{},
	}

	fn(cfg)

	if instanceName == "default" {
		// 单库，第一次注册 default，提供不带 name 的数据库
		c.container = append(c.container,
			fx.Provide(func(lc fx.Lifecycle, logger *zap.Logger) *kafka.Writer {
				return kafkax.NewWriter(lc, cfg, logger)
			}),
		)
	} else {
		// 多库，或显式传 name 的 writer，使用 name 标签
		c.container = append(c.container,
			fx.Provide(
				fx.Annotate(
					func(lc fx.Lifecycle, logger *zap.Logger) *kafka.Writer {
						return kafkax.NewWriter(lc, cfg, logger)
					},
					fx.ResultTags(`name:"`+instanceName+`"`),
				),
			),
		)
	}

	c.writerMap[instanceName] = struct{}{}

	return c
}

func (o *Options) Container() []fx.Option {
	return o.container
}

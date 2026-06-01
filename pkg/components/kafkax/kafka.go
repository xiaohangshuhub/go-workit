package kafkax

import (
	"github.com/segmentio/kafka-go"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type ReaderOptions struct {
	kafka.ReaderConfig
}

type WriterOptions struct {
	AllowAutoTopicCreation bool
	kafka.WriterConfig
}

func NewReader(lc fx.Lifecycle, cfg *ReaderOptions, logger *zap.Logger) *kafka.Reader {

	// 创建Reader
	r := kafka.NewReader(cfg.ReaderConfig)
	return r
}

func NewWriter(lc fx.Lifecycle, cfg *WriterOptions, logger *zap.Logger) *kafka.Writer {
	w := kafka.NewWriter(cfg.WriterConfig)
	w.AllowAutoTopicCreation = cfg.AllowAutoTopicCreation
	return w
}

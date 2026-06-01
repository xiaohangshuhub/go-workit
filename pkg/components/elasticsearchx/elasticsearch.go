package elasticsearchx

import (
	"github.com/elastic/go-elasticsearch/v7"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Options struct {
	elasticsearch.Config
}

func NewClient(lc fx.Lifecycle, cfg *Options, logger *zap.Logger) *elasticsearch.Client {
	client, err := elasticsearch.NewClient(cfg.Config)
	if err != nil {
		logger.Error("Elasticsearch client creation failed", zap.Error(err))
		panic(err)
	}
	return client
}

package miniox

import (
	"github.com/minio/minio-go/v7"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Options struct {
	Endpoint string
	minio.Options
}

func NewClient(lc fx.Lifecycle, cfg *Options, logger *zap.Logger) *minio.Client {
	client, err := minio.New(cfg.Endpoint, &cfg.Options)
	if err != nil {
		logger.Error("Minio client creation failed", zap.Error(err))
		panic(err)
	}
	return client
}

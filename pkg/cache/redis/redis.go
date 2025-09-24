package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Options struct {
	redis.Options
}

func NewRedisClient(lc fx.Lifecycle, cfg *Options, logger *zap.Logger) *redis.Client {
	client := redis.NewClient(&cfg.Options)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// 用启动时传进来的 ctx 做连接检测，避免阻塞太久
			pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			if err := client.Ping(pingCtx).Err(); err != nil {
				logger.Error("Redis ping failed", zap.Error(err))
				return err
			}

			logger.Info("Connected to Redis", zap.String("addr", cfg.Addr))
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Closing Redis client")
			return client.Close()
		},
	})

	return client
}

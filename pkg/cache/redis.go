package cache

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type RedisConfigOptions struct {
	redis.Options
}

func NewRedisClient(lc fx.Lifecycle, cfg *RedisConfigOptions, logger *zap.Logger) (*redis.Client, error) {

	client := redis.NewClient(&cfg.Options)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		logger.Error("Redis ping failed", zap.Error(err))
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			logger.Info("Closing Redis client")
			return client.Close()
		},
	})

	logger.Info("Connected to Redis", zap.String("addr", cfg.Addr))

	return client, nil
}

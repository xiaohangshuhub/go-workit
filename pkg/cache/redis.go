package cache

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

func NewRedisClient(lc fx.Lifecycle, v *viper.Viper, logger *zap.Logger) (*redis.Client, error) {
	var redisCfg RedisConfig

	if err := v.UnmarshalKey("redis", &redisCfg); err != nil {
		logger.Error("Failed to unmarshal redis config", zap.Error(err))
		return nil, err
	}

	client := redis.NewClient(&redis.Options{
		Addr:     redisCfg.Addr,
		Password: redisCfg.Password,
		DB:       redisCfg.DB,
		PoolSize: redisCfg.PoolSize,
	})

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

	logger.Info("Connected to Redis", zap.String("addr", redisCfg.Addr))

	return client, nil
}

// 提供模块注册
func RedisModule() fx.Option {
	return fx.Provide(NewRedisClient)
}

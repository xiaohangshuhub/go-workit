package database

import (
	"github.com/spf13/viper"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// PostgresConfig = 公共字段 + 扩展字段
type PostgresConfig struct {
	CommonDatabaseConfig
	PreferSimpleProtocol bool `mapstructure:"prefer_simple_protocol"`
}

func NewPostgresDB(lc fx.Lifecycle, v *viper.Viper, zapLogger *zap.Logger) (*gorm.DB, error) {
	var cfg PostgresConfig
	if err := v.UnmarshalKey("database", &cfg); err != nil {
		zapLogger.Error("Failed to unmarshal postgres config", zap.Error(err))
		return nil, err
	}

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  cfg.DSN,
		PreferSimpleProtocol: cfg.PreferSimpleProtocol,
	}), &gorm.Config{
		Logger: NewGormZapLogger(zapLogger, cfg.LogLevel, cfg.SlowThreshold),
		DryRun: cfg.DryRun,
	})
	if err != nil {
		zapLogger.Error("Failed to open GORM postgres", zap.Error(err))
		return nil, err
	}

	configureConnectionPool(db, cfg.CommonDatabaseConfig, zapLogger, lc)
	return db, nil
}

func PostgresModule() fx.Option {
	return fx.Provide(NewPostgresDB)
}

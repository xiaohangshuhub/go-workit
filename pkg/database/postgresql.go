package database

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// PostgresConfig = 公共字段 + 扩展字段
type PostgresConfig struct {
	CommonDatabaseConfig `mapstructure:",squash"`
	PreferSimpleProtocol bool `mapstructure:"prefer_simple_protocol"`
}

func NewPostgresDB(lc fx.Lifecycle, cfg *PostgresConfig, zapLogger *zap.Logger) (*gorm.DB, error) {

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

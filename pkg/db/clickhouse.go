package db

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/driver/clickhouse"
	"gorm.io/gorm"
)

type ClickhouseConfig struct {
	DatabaseConfig
	ClickhouseCfg clickhouse.Config
}

// NewClickhouse
func NewClickhouse(lc fx.Lifecycle, cfg *ClickhouseConfig, logger *zap.Logger) (*gorm.DB, error) {

	if cfg.Config.Logger == nil {
		cfg.Config.Logger = NewGormZapLogger(logger, cfg.LogLevel, cfg.SlowThreshold)
	}

	db, err := gorm.Open(clickhouse.New(cfg.ClickhouseCfg), cfg.Config)

	if err != nil {
		logger.Error("Failed to open GORM Clickhouse", zap.Error(err))
		return nil, err
	}

	configureConnectionPool(db, cfg.DatabaseConfig, logger, lc)

	return db, nil
}

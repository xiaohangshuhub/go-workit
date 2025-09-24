package db

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/driver/gaussdb"
	"gorm.io/gorm"
)

type GaussDBConfig struct {
	DatabaseConfig
	GaussDBCfg gaussdb.Config
}

// NewGaussDB
func NewGaussDB(lc fx.Lifecycle, cfg *GaussDBConfig, logger *zap.Logger) (*gorm.DB, error) {

	if cfg.Config.Logger == nil {
		cfg.Config.Logger = NewGormZapLogger(logger, cfg.LogLevel, cfg.SlowThreshold)
	}

	db, err := gorm.Open(gaussdb.New(cfg.GaussDBCfg), cfg.Config)

	if err != nil {
		logger.Error("Failed to open GORM GaussDB", zap.Error(err))
		return nil, err
	}

	configureConnectionPool(db, cfg.DatabaseConfig, logger, lc)

	return db, nil
}

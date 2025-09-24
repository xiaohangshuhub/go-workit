package db

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

type SQLServerConfigOptions struct {
	DatabaseConfig
	SQLServerCfg sqlserver.Config
}

// NewSQLServer
func NewSQLServer(lc fx.Lifecycle, cfg *SQLServerConfigOptions, logger *zap.Logger) (*gorm.DB, error) {

	if cfg.Config.Logger == nil {
		cfg.Config.Logger = NewGormZapLogger(logger, cfg.LogLevel, cfg.SlowThreshold)
	}

	db, err := gorm.Open(sqlserver.New(cfg.SQLServerCfg), cfg.Config)

	if err != nil {
		logger.Error("Failed to open GORM SQLServer", zap.Error(err))
		return nil, err
	}

	configureConnectionPool(db, cfg.DatabaseConfig, logger, lc)

	return db, nil
}

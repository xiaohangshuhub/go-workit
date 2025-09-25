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
func NewSQLServer(lc fx.Lifecycle, opts *SQLServerConfigOptions, logger *zap.Logger) (*gorm.DB, error) {

	if opts.Config.Logger == nil {
		opts.Config.Logger = NewGormZapLogger(logger, opts.LogLevel, opts.SlowThreshold)
	}

	db, err := gorm.Open(sqlserver.New(opts.SQLServerCfg), opts.Config)

	if err != nil {
		logger.Error("Failed to open GORM SQLServer", zap.Error(err))
		return nil, err
	}

	configureConnectionPool(db, opts.DatabaseConfig, logger, lc)

	return db, nil
}

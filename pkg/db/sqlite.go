package db

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type SQLiteConfigOptions struct {
	DatabaseConfig
	SQLiteCfg sqlite.Config
}

// NewSQLServer
func NewSQLite(lc fx.Lifecycle, opts *SQLiteConfigOptions, logger *zap.Logger) (*gorm.DB, error) {

	if opts.Config.Logger == nil {
		opts.Config.Logger = NewGormZapLogger(logger, opts.LogLevel, opts.SlowThreshold)
	}

	db, err := gorm.Open(sqlite.New(opts.SQLiteCfg), opts.Config)

	if err != nil {
		logger.Error("Failed to open GORM SQLite", zap.Error(err))
		return nil, err
	}

	configureConnectionPool(db, opts.DatabaseConfig, logger, lc)

	return db, nil
}

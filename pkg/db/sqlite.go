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
func NewSQLite(lc fx.Lifecycle, cfg *SQLiteConfigOptions, logger *zap.Logger) (*gorm.DB, error) {

	if cfg.Config.Logger == nil {
		cfg.Config.Logger = NewGormZapLogger(logger, cfg.LogLevel, cfg.SlowThreshold)
	}

	db, err := gorm.Open(sqlite.New(cfg.SQLiteCfg), cfg.Config)

	if err != nil {
		logger.Error("Failed to open GORM SQLite", zap.Error(err))
		return nil, err
	}

	configureConnectionPool(db, cfg.DatabaseConfig, logger, lc)

	return db, nil
}

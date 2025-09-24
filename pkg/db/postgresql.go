package db

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// PostgresConfigOptions
type PostgresConfigOptions struct {
	DatabaseConfig `mapstructure:",squash"`
	PgSQLCfg       postgres.Config
}

// NewPostgresDB
func NewPostgresDB(lc fx.Lifecycle, cfg *PostgresConfigOptions, logger *zap.Logger) (*gorm.DB, error) {

	if cfg.Config.Logger == nil {
		cfg.Config.Logger = NewGormZapLogger(logger, cfg.LogLevel, cfg.SlowThreshold)
	}

	db, err := gorm.Open(postgres.New(cfg.PgSQLCfg), cfg.Config)

	if err != nil {
		logger.Error("Failed to open GORM PostgreSQL", zap.Error(err))
		return nil, err
	}

	configureConnectionPool(db, cfg.DatabaseConfig, logger, lc)
	return db, nil
}

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
func NewPostgresDB(lc fx.Lifecycle, opts *PostgresConfigOptions, logger *zap.Logger) (*gorm.DB, error) {

	if opts.Config.Logger == nil {
		opts.Config.Logger = NewGormZapLogger(logger, opts.LogLevel, opts.SlowThreshold)
	}

	db, err := gorm.Open(postgres.New(opts.PgSQLCfg), opts.Config)

	if err != nil {
		logger.Error("Failed to open GORM PostgreSQL", zap.Error(err))
		return nil, err
	}

	configureConnectionPool(db, opts.DatabaseConfig, logger, lc)
	return db, nil
}

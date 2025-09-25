package db

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// MySQLConfigOptions  公共字段 + 扩展字段
type MySQLConfigOptions struct {
	DatabaseConfig
	MySQLCfg mysql.Config
}

// NewMySQLDB
func NewMysqlDB(lc fx.Lifecycle, opts *MySQLConfigOptions, logger *zap.Logger) (*gorm.DB, error) {

	if opts.Config.Logger == nil {
		opts.Config.Logger = NewGormZapLogger(logger, opts.LogLevel, opts.SlowThreshold)
	}

	db, err := gorm.Open(mysql.New(opts.MySQLCfg), opts.Config)

	if err != nil {
		logger.Error("Failed to open GORM MySQL", zap.Error(err))
		return nil, err
	}

	configureConnectionPool(db, opts.DatabaseConfig, logger, lc)

	return db, nil
}

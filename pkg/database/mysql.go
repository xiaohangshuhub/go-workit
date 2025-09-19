package database

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
func NewMysqlDB(lc fx.Lifecycle, cfg *MySQLConfigOptions, logger *zap.Logger) (*gorm.DB, error) {

	if cfg.Config.Logger == nil {
		cfg.Config.Logger = NewGormZapLogger(logger, cfg.LogLevel, cfg.SlowThreshold)
	}

	db, err := gorm.Open(mysql.New(cfg.MySQLCfg), cfg.Config)

	if err != nil {
		logger.Error("Failed to open GORM mysql", zap.Error(err))
		return nil, err
	}

	configureConnectionPool(db, cfg.DatabaseConfig, logger, lc)

	return db, nil
}

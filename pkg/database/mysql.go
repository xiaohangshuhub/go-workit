package database

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// MysqlConfig = 公共字段 + 扩展字段
type MysqlConfigOptions struct {
	CommonDatabaseConfig `mapstructure:",squash"`
	// 比如扩展字段（如果以后有）
	SomeMysqlSpecialOption bool `mapstructure:"some_mysql_special_option"`
}

func NewMysqlDB(lc fx.Lifecycle, cfg *MysqlConfigOptions, zapLogger *zap.Logger) (*gorm.DB, error) {

	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{
		Logger: NewGormZapLogger(zapLogger, cfg.LogLevel, cfg.SlowThreshold),
		DryRun: cfg.DryRun,
	})
	if err != nil {
		zapLogger.Error("Failed to open GORM mysql", zap.Error(err))
		return nil, err
	}

	configureConnectionPool(db, cfg.CommonDatabaseConfig, zapLogger, lc)
	return db, nil
}

package database

import (
	"github.com/spf13/viper"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// MysqlConfig = 公共字段 + 扩展字段
type MysqlConfig struct {
	CommonDatabaseConfig
	// 比如扩展字段（如果以后有）
	SomeMysqlSpecialOption bool `mapstructure:"some_mysql_special_option"`
}

func NewMysqlDB(lc fx.Lifecycle, v *viper.Viper, zapLogger *zap.Logger) (*gorm.DB, error) {
	var cfg MysqlConfig
	if err := v.UnmarshalKey("database", &cfg); err != nil {
		zapLogger.Error("Failed to unmarshal mysql config", zap.Error(err))
		return nil, err
	}

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

func MysqlModule() fx.Option {
	return fx.Provide(NewMysqlDB)
}

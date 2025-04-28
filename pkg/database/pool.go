package database

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func configureConnectionPool(db *gorm.DB, cfg CommonDatabaseConfig, zapLogger *zap.Logger, lc fx.Lifecycle) {
	sqlDB, err := db.DB()
	if err != nil {
		zapLogger.Error("Failed to get sql.DB from GORM", zap.Error(err))
		return
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			zapLogger.Info("Closing database connection")
			return sqlDB.Close()
		},
	})
}

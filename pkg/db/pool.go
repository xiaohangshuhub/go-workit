package db

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// configureConnectionPool 配置连接池并注册 Fx 生命周期
func configureConnectionPool(db *gorm.DB, cfg DatabaseConfig, logger *zap.Logger, lc fx.Lifecycle) {
	sqlDB, err := db.DB()
	if err != nil {
		logger.Error("Failed to get sql.DB from GORM", zap.Error(err))
		return
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// 注册 Fx 生命周期
	lc.Append(fx.Hook{
		// OnStart 时检查数据库连接
		OnStart: func(ctx context.Context) error {
			if err := sqlDB.PingContext(ctx); err != nil {
				logger.Error("Failed to ping database on start", zap.Error(err))
				return err
			}
			logger.Info("Database connection established")
			return nil
		},
		// OnStop 时关闭数据库连接
		OnStop: func(ctx context.Context) error {
			logger.Info("Closing database connection")
			return sqlDB.Close()
		},
	})
}

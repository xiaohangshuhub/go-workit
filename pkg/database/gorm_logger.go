package database

import (
	"context"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm/logger"
)

// GormZapLogger 让GORM日志走zap
type GormZapLogger struct {
	logger        *zap.Logger
	level         logger.LogLevel
	slowThreshold time.Duration
}

// 构造器
func NewGormZapLogger(l *zap.Logger, logLevel string, slowThreshold time.Duration) logger.Interface {
	return &GormZapLogger{
		logger:        l,
		level:         mapLogLevel(logLevel),
		slowThreshold: slowThreshold,
	}
}

func mapLogLevel(level string) logger.LogLevel {
	switch level {
	case "silent":
		return logger.Silent
	case "error":
		return logger.Error
	case "warn":
		return logger.Warn
	case "info":
		return logger.Info
	default:
		return logger.Info
	}
}

// 实现 logger.Interface 四个方法
func (g *GormZapLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *g
	newLogger.level = level
	return &newLogger
}

func (g *GormZapLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if g.level >= logger.Info {
		g.logger.Sugar().Infof(msg, data...)
	}
}

func (g *GormZapLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if g.level >= logger.Warn {
		g.logger.Sugar().Warnf(msg, data...)
	}
}

func (g *GormZapLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if g.level >= logger.Error {
		g.logger.Sugar().Errorf(msg, data...)
	}
}

func (g *GormZapLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if g.level <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	switch {
	case err != nil && g.level >= logger.Error:
		g.logger.Error("GORM SQL Error",
			zap.Error(err),
			zap.Duration("elapsed", elapsed),
			zap.Int64("rows", rows),
			zap.String("sql", sql),
		)
	case elapsed > g.slowThreshold && g.slowThreshold != 0 && g.level >= logger.Warn:
		g.logger.Warn("GORM SQL Slow",
			zap.Duration("elapsed", elapsed),
			zap.Int64("rows", rows),
			zap.String("sql", sql),
		)
	case g.level >= logger.Info:
		g.logger.Info("GORM SQL",
			zap.Duration("elapsed", elapsed),
			zap.Int64("rows", rows),
			zap.String("sql", sql),
		)
	}
}

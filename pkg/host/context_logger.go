package host

import (
	"context"

	"go.uber.org/zap"
)

type ctxLoggerKey struct {
}

// WithContext 将日志记录器添加到上下文中
func WithContext(ctx context.Context, log *zap.Logger, fields ...zap.Field) context.Context {
	logger := log.With(fields...)
	return context.WithValue(ctx, ctxLoggerKey{}, logger)
}

// FromContext 从上下文中获取日志记录器
func FromContext(ctx context.Context, log *zap.Logger) *zap.Logger {
	if ctx == nil {
		return log
	}
	if logger, ok := ctx.Value(ctxLoggerKey{}).(*zap.Logger); ok {
		return logger
	}
	return log
}

// DebugContext 使用上下文输出Debug级别日志
func DebugContext(ctx context.Context, msg string, fields ...zap.Field) {
	FromContext(ctx, zap.L()).Debug(msg, fields...)
}

// InfoContext 使用上下文输出Info级别日志
func InfoContext(ctx context.Context, msg string, fields ...zap.Field) {
	FromContext(ctx, zap.L()).Info(msg, fields...)
}

// WarnContext 使用上下文输出Warn级别日志
func WarnContext(ctx context.Context, msg string, fields ...zap.Field) {
	FromContext(ctx, zap.L()).Warn(msg, fields...)
}

// ErrorContext 使用上下文输出Error级别日志
func ErrorContext(ctx context.Context, msg string, fields ...zap.Field) {
	FromContext(ctx, zap.L()).Error(msg, fields...)
}

package echo

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

// logRequest 统一请求日志记录方法
func logRequest(logger *zap.Logger, status int, method, uri, ip string, latency time.Duration, isDebug bool) {
	// Release模式只记录 status >= 400
	if !isDebug && status < 400 {
		return
	}

	fields := []zap.Field{
		zap.Int("status", status),
		zap.String("method", method),
		zap.String("path", uri),
		zap.String("ip", ip),
		zap.Duration("latency", latency),
	}

	switch {
	case status >= 500:
		logger.Error("HTTP Request", fields...)
	case status >= 400:
		logger.Warn("HTTP Request", fields...)
	default:
		logger.Info("HTTP Request", fields...)
	}
}

// newZapLogger 返回 Echo 的请求日志中间件
func newZapLogger(logger *zap.Logger, isDebug bool) echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			start := time.Now()
			c.Response().Before(func() {
				latency := time.Since(start)
				method := c.Request().Method
				ip := c.RealIP()
				logRequest(logger, v.Status, method, v.URI, ip, latency, isDebug)
			})
			return nil
		},
	})
}

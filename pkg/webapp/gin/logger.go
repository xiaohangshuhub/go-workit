package gin

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// newZapLogger returns a gin.HandlerFunc that logs requests using zap.
func newZapLogger(logger *zap.Logger) gin.HandlerFunc {
	isDebug := gin.Mode() == gin.DebugMode

	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.Path
		clientIP := c.ClientIP()

		// release模式，只记录4xx、5xx
		if !isDebug && statusCode < 400 {
			return
		}

		fields := []zap.Field{
			zap.Int("status", statusCode),
			zap.String("method", method),
			zap.String("path", path),
			zap.String("ip", clientIP),
			zap.Duration("latency", latency),
		}

		switch {
		case statusCode >= 500:
			logger.Error("HTTP Request", fields...)
		case statusCode >= 400:
			logger.Warn("HTTP Request", fields...)
		default:
			logger.Info("HTTP Request", fields...)
		}
	}
}

package webapp

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

// recoveryWithZapEcho returns a middleware function that recovers from panics
func newEchoRecoveryWithZap(logger *zap.Logger) echo.MiddlewareFunc {
	return middleware.RecoverWithConfig(middleware.RecoverConfig{
		LogErrorFunc: func(c echo.Context, err error, stack []byte) error {
			req := c.Request()
			logger.Error("panic recovered",
				zap.Error(err),
				zap.String("path", req.URL.Path),
				zap.String("method", req.Method),
				zap.String("ip", c.RealIP()),
				zap.ByteString("stack", stack),
			)

			// 返回统一错误
			return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
		},
	})
}

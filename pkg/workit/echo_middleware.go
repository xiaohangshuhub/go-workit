package workit

import "github.com/labstack/echo/v4"

// EchoMiddleware echo中间件接口
type EchoMiddleware interface {
	Handle() echo.MiddlewareFunc
	ShouldSkip(path string, method string) bool
}

package workit

import "github.com/labstack/echo/v4"

type EchoMiddleware interface {
	Handle() echo.MiddlewareFunc
	ShouldSkip(path string, method string) bool
}

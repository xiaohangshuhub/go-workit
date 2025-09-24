package echo

import "github.com/labstack/echo/v4"

// Middleware echo中间件接口
type Middleware interface {
	Handle() echo.MiddlewareFunc
}

package webapp

import (
	"github.com/labstack/echo/v4"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"go.uber.org/zap"
)

// GinAuthorizationMiddleware 授权中间件
type EchoLocalizationMiddleware struct {
	*LocalizationOptions
	logger *zap.Logger
}

// newGinAuthorizationMiddleware 初始国际化中间件
func newEchoLocalizationMiddleware(opts *LocalizationOptions, logger *zap.Logger) *EchoLocalizationMiddleware {
	return &EchoLocalizationMiddleware{
		LocalizationOptions: opts,
		logger:              logger,
	}
}

// Handle 授权中间件处理函数
func (l *EchoLocalizationMiddleware) Handle() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			// 从请求头获取语言
			lang := c.Request().Header.Get("Accept-Language")
			if lang == "" {
				lang = l.DefaultLanguage
			}

			// 创建本地化器
			localizer := i18n.NewLocalizer(l.Bundle, lang)

			// 将本地化器存储到上下文中
			c.Set("localizer", localizer)
			// 继续执行后续中间件
			return next(c)
		}
	}
}
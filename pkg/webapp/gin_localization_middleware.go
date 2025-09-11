package webapp

import (
	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"go.uber.org/zap"
)

// GinAuthorizationMiddleware 授权中间件
type GinLocalizationMiddleware struct {
	*LocalizationOptions
	logger *zap.Logger
}

// newGinAuthorizationMiddleware 初始国际化中间件
func newGinLocalizationMiddleware(opts *LocalizationOptions, logger *zap.Logger) *GinLocalizationMiddleware {
	return &GinLocalizationMiddleware{
		LocalizationOptions: opts,
		logger:              logger,
	}
}

// Handle 授权中间件处理函数
func (l *GinLocalizationMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := c.GetHeader("Accept-Language")
		if lang == "" {
			lang = l.DefaultLanguage
		}

		// 创建本地化器
		localizer := i18n.NewLocalizer(l.Bundle, lang)

		// 将本地化器存储到上下文中
		c.Set("localizer", localizer)

		// 继续执行后续中间件
		c.Next()
	}
}

// ShouldSkip 跳过逻辑
func (a *GinLocalizationMiddleware) ShouldSkip(path string, method string) bool {
	return false
}

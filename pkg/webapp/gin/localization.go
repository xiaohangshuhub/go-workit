package gin

import (
	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/web"
	"go.uber.org/zap"
)

// newAuthorize 授权中间件
type Localization struct {
	web.Localization
	logger *zap.Logger
}

// newLocalization 初始国际化中间件
func newLocalization(provider web.Localization, logger *zap.Logger) Middleware {
	return &Localization{
		Localization: provider,
		logger:       logger,
	}
}

// Handle 授权中间件处理函数
func (l *Localization) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := c.GetHeader("Accept-Language")
		if lang == "" {
			lang = l.DefaultLanguage()
		}

		// 创建本地化器
		localizer := i18n.NewLocalizer(l.Bundle(), lang)

		// 将本地化器存储到上下文中
		c.Set("localizer", localizer)

		// 继续执行后续中间件
		c.Next()
	}
}

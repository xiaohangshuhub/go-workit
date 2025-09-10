// Package main API文档
//
// @title           我的服务 API
// @version         1.0
// @description     这是一个示例 API 文档
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description 输入格式: Bearer {token}
package main

import (
	"github.com/gin-gonic/gin"
	"github.com/mattermost/mattermost/server/public/pluginapi/i18n"
	_ "github.com/xiaohangshuhub/go-workit/api/service1/docs" // swagger 一定要有这行,指向你的文档地址
	"github.com/xiaohangshuhub/go-workit/pkg/workit"
)

func main() {
	// web应用构建器
	builder := workit.NewWebAppBuilder()

	// 配置构建器(注册即生效)
	builder.AddConfig(func(build workit.ConfigBuilder) {
		build.AddYamlFile("./application.yaml")
	})

	builder.AddLocalization(func(opts *workit.LocalizationOptions) {
		opts.DefaultLanguage = "zh-CN"
		opts.SupportedLanguages = []string{"en-US", "zh-CN"}
		opts.TranslationsDir = "locales"

	})

	// 构建Web应用
	app := builder.Build()

	app.UseLocalization()

	// 配置路由
	app.MapRouter(func(router *gin.Engine) {
		router.GET("/hello", func(c *gin.Context) {
			// 获取本地化器
			localizer := c.MustGet("localizer").(*i18n.Localizer)

			// 翻译消息
			msg, _ := localizer.Localize(&i18n.LocalizeConfig{
				MessageID: "hello",
			})
			c.JSON(200, gin.H{
				"message": msg,
			})
		})
	})

	// 运行应用
	app.Run()
}

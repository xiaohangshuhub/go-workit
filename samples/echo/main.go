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
	"github.com/labstack/echo/v4"
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
		opts.DefaultLanguage = "en-US"
		opts.SupportedLanguages = []string{"en-US", "zh-CN"}
		opts.TranslationsDir = "locales"
		opts.FileType = workit.LocalizationFileTypeJSON

	})

	// 构建Web应用
	app := builder.Build(func(b *workit.WebApplicationBuilder) workit.WebApplication {

		return workit.NewEchoWebApplication(workit.WebApplicationOptions{

			Config:    b.Config,
			Logger:    b.Logger,
			Container: b.Container,
		})

	})

	app.UseLocalization()

	// 配置路由
	app.MapRouter(func(router *echo.Echo) {
		router.GET("/hello", func(c echo.Context) error {

			// 获取本地化器
			localizer := c.Get("localizer").(*i18n.Localizer)

			// 翻译消息
			msg, _ := localizer.Localize(&i18n.LocalizeConfig{
				MessageID: "hello",
			})

			return c.JSON(200, map[string]string{
				"message": msg,
			})
		})
	})

	// 运行应用
	app.Run()
}

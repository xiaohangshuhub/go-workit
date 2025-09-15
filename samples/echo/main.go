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
	"github.com/nicksnyder/go-i18n/v2/i18n"
	_ "github.com/xiaohangshuhub/go-workit/api/service1/docs" // swagger 一定要有这行,指向你的文档地址
	"github.com/xiaohangshuhub/go-workit/pkg/app"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp"
)

func main() {
	// web应用构建器
	builder := webapp.NewBuilder()

	// 配置构建器(注册即生效)
	builder.AddConfig(func(build *app.ConfigOptions) {
		build.AddYamlFile("./application.yaml")
	})

	builder.AddRouter(func(ro *webapp.RouterOptions) {
		ro.MapGet("/v1/hello", func() echo.HandlerFunc {

			return func(c echo.Context) error {

				return c.JSON(200, map[string]string{
					"message": "hello word",
				})
			}
		})
	})

	builder.AddLocalization(func(opts *webapp.LocalizationOptions) {
		opts.DefaultLanguage = "en-US"
		opts.SupportedLanguages = []string{"en-US", "zh-CN"}
		opts.TranslationsDir = "locales"
		opts.FileType = webapp.LocalizationFileTypeJSON

	})

	// 构建Web应用
	app := builder.Build(func(b *webapp.WebApplicationBuilder) webapp.WebApplication {

		return webapp.NewEchoWebApplication(webapp.WebApplicationOptions{

			Config:        b.Config,
			Logger:        b.Logger,
			Container:     b.Container,
			App:           b.Application,
			RouterOptions: b.RouterOptions,
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

	app.UseRouting()
	// 运行应用
	app.Run()
}

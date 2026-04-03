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
	"github.com/nicksnyder/go-i18n/v2/i18n"
	_ "github.com/xiaohangshu-dev/go-workit/api/service1/docs" // swagger 一定要有这行,指向你的文档地址
	"github.com/xiaohangshu-dev/go-workit/pkg/webapp"
	"github.com/xiaohangshu-dev/go-workit/pkg/webapp/localiza"
)

func main() {

	builder := webapp.NewBuilder()

	builder.AddLocalization(func(opts *localiza.Options) {
		opts.DefaultLanguage = "en-US"
		opts.SupportedLanguages = []string{"en-US", "zh-CN"}
		opts.TranslationsDir = "./locales"
		opts.FileType = localiza.LocalizationFileTypeJSON
	})

	app := builder.Build()

	app.UseLocalization()

	app.MapRoute(func(router *gin.Engine) {
		router.GET("/hello", func(c *gin.Context) {
			// 获取本地化器
			localizer := c.MustGet("localizer").(*i18n.Localizer)

			msg, _ := localizer.Localize(&i18n.LocalizeConfig{
				MessageID: "hello",
			})
			c.JSON(200, gin.H{
				"message": msg,
			})
		})
	})

	app.Run()
}

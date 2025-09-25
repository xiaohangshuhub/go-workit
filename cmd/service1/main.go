package main

import (
	_ "cli-echo/api/service1/docs" // swagger 一定要有这行,指向你的文档地址
	"cli-echo/internal/service1/webapi"

	"github.com/xiaohangshuhub/go-workit/pkg/webapp"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/echo"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/web"
)

func main() {

	builder := webapp.NewBuilder()

	app := builder.Build(func(b *webapp.WebApplicationBuilder) web.Application {

		return echo.NewWebApplication(b.App(), b.Router())
	})

	if app.Env().IsDevelopment {
		app.UseSwagger()
	}

	app.MapRoute(webapi.Hello)

	app.Run()
}

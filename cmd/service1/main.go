package main

import (
	_ "cli-echo/api/service1/docs" // swagger 一定要有这行,指向你的文档地址
	"cli-echo/internal/service1/grpcapi/hello"
	"cli-echo/internal/service1/webapi"

	"github.com/xiaohangshuhub/go-workit/pkg/webapp"
)

func main() {

	builder := webapp.NewBuilder()

	app := builder.Build(func(b *webapp.WebApplicationBuilder) webapp.WebApplication {

		return webapp.NewEchoWebApplication(webapp.WebApplicationOptions{
			Logger:    b.Logger,
			Config:    b.Config,
			Container: b.Container,
			App:       b.Application,
		})
	})

	if app.Env().IsDevelopment {
		app.UseSwagger()
	}

	app.MapRouter(webapi.Hello)

	app.MapGrpcServices(hello.NewHelloService)

	app.Run()
}

package main

import (
	_ "cli-echo/api/service1/docs" // swagger 一定要有这行,指向你的文档地址
	"cli-echo/internal/service1/grpcapi/hello"
	"cli-echo/internal/service1/webapi"

	"github.com/xiaohangshuhub/go-workit/pkg/workit"
)

func main() {

	builder := workit.NewWebAppBuilder()

	builder.AddConfig(func(build workit.ConfigBuilder) {
		build.AddYamlFile("./application.yaml")
	})

	app := builder.Build(func(b *workit.WebApplicationBuilder) workit.WebApplication {

		return workit.NewEchoWebApplication(workit.WebApplicationOptions{
			Logger:    b.Logger,
			Config:    b.Config,
			Container: b.Container,
		})
	})

	if app.Environment().IsDevelopment {
		app.UseSwagger()
	}

	app.MapRoutes(webapi.Hello)

	app.MapGrpcServices(hello.NewHelloService)

	app.Run()
}

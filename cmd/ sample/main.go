package main

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/lxhanghub/newb/pkg/host"
)

func main() {

	builder := host.NewWebHostBuilder()

	builder.MapGet("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "hello world",
		})
	})

	app, err := builder.Build()

	if err != nil {
		fmt.Println("Error starting the server:", err)
		return
	}

	app.Run(context.Background())
}

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
	"github.com/segmentio/kafka-go"
	_ "github.com/xiaohangshu-dev/go-workit/api/service1/docs" // swagger 一定要有这行,指向你的文档地址
	"github.com/xiaohangshu-dev/go-workit/examples/mq/kafka/bs"
	"github.com/xiaohangshu-dev/go-workit/pkg/mq/kafkax"
	"github.com/xiaohangshu-dev/go-workit/pkg/webapp"
	"github.com/xiaohangshu-dev/go-workit/pkg/webapp/kafkactx"
)

func main() {

	builder := webapp.NewBuilder()

	builder.AddBackgroundService(bs.NewSubscriberService)
	builder.AddBackgroundService(bs.NewProducerService)

	builder.AddKafkaContext(func(opts *kafkactx.Options) {
		opts.UseReaderClient("default", func(cfg *kafkax.ReaderOptions) {
			cfg.Brokers = []string{"localhost:9092"}
			cfg.Topic = "test"
			cfg.GroupID = "test-group-1"
			cfg.MaxBytes = 10e6 // 10MB
			cfg.StartOffset = kafka.LastOffset
			cfg.CommitInterval = 0
		})
		opts.UseWriterClient("default", func(cfg *kafkax.WriterOptions) {
			cfg.Brokers = []string{"localhost:9092"}
			cfg.Topic = "test"
			cfg.Balancer = &kafka.LeastBytes{} // 指定分区的balancer模式为最小字节分布
			cfg.BatchBytes = 900 * 1024
			cfg.BatchSize = 10
			cfg.RequiredAcks = int(kafka.RequireAll) // ack模式
			cfg.Async = true
			cfg.AllowAutoTopicCreation = true
		})
	})

	app := builder.Build()

	app.MapRoute(func(router *gin.Engine) {
		router.GET("/hello", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "Hello, World!",
			})
		})
	})

	app.Run()
}

package bs

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/xiaohangshu-dev/go-workit/pkg/app"
	"go.uber.org/zap"
)

// 生产者服务
type ProducerService struct {
	log    *zap.Logger
	writer *kafka.Writer
}

func NewProducerService(log *zap.Logger, writer *kafka.Writer) app.BackgroundService {
	return &ProducerService{
		writer: writer,
		log:    log,
	}
}

func (b *ProducerService) Start(ctx context.Context) error {
	for {
		err := b.writer.WriteMessages(context.Background(),
			kafka.Message{
				Key:   []byte(time.Now().Format("20060102150405")),
				Value: []byte(time.Now().Format("20060102150405")),
			},
		)
		if err != nil {
			b.log.Fatal("failed to write messages:", zap.Error(err))
		}
		log.Println("send message:", time.Now().Format("20060102150405"))

		time.Sleep(time.Second)
	}
	return nil
}
func (b *ProducerService) Stop(ctx context.Context) error {
	b.log.Info("ProducerService Stop")
	fmt.Println("ProducerService Stop")
	if err := b.writer.Close(); err != nil {
		log.Fatal("failed to close writer:", err)
	}
	return nil
}

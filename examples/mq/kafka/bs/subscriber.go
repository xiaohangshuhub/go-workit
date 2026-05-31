package bs

import (
	"context"
	"fmt"

	"github.com/segmentio/kafka-go"
	"github.com/xiaohangshu-dev/go-workit/pkg/app"
	"go.uber.org/zap"
)

type SubscriberService struct {
	log    *zap.Logger
	reader *kafka.Reader
}

func NewSubscriberService(log *zap.Logger, reader *kafka.Reader) app.BackgroundService {
	return &SubscriberService{
		reader: reader,
		log:    log,
	}
}

func (b *SubscriberService) Start(ctx context.Context) error {

	// 接收消息
	for {
		// 获取消息
		m, err := b.reader.FetchMessage(ctx)
		if err != nil {
			b.log.Error("failed to fetch message:", zap.Error(err))
			break
		}
		// 处理消息
		fmt.Printf("message at topic/partition/offset %v/%v/%v: %s = %s\n", m.Topic, m.Partition, m.Offset, string(m.Key), string(m.Value))
		// 显式提交
		if err := b.reader.CommitMessages(ctx, m); err != nil {
			b.log.Fatal("failed to commit messages:", zap.Error(err))
		}
	}

	// 程序退出前关闭Reader
	if err := b.reader.Close(); err != nil {
		b.log.Fatal("failed to close reader:", zap.Error(err))
	}

	return nil
}
func (b *SubscriberService) Stop(ctx context.Context) error {
	b.log.Info("SubscriberService Stop")
	fmt.Println("SubscriberService Stop")
	return nil
}

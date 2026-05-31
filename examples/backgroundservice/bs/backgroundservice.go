package bs

import (
	"context"
	"fmt"

	"github.com/xiaohangshu-dev/go-workit/pkg/app"
	"go.uber.org/zap"
)

type BackService struct {
	log *zap.Logger
}

func NewBackService(log *zap.Logger) app.BackgroundService {
	return &BackService{
		log: log,
	}
}

func (b *BackService) Start(ctx context.Context) error {
	b.log.Info("BackService Start")
	fmt.Println("BackService Start")
	return nil
}
func (b *BackService) Stop(ctx context.Context) error {
	b.log.Info("BackService Stop")
	fmt.Println("BackService Stop")
	return nil
}

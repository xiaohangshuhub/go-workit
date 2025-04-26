package host

import "context"

// BackgroundService 定义后台任务标准接口
type BackgroundService interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

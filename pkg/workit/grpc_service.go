package workit

import (
	"context"
	"net"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// GrpcService 定义 gRPC 服务接口
type GrpcService interface {
	Register(server *grpc.Server)
}

// NewGrpcServer 启动 gRPC 服务
func NewGrpcServer(lc fx.Lifecycle, logger *zap.Logger, opt ServerOptions) *grpc.Server {
	grpcServer := grpc.NewServer()

	// 使用 fx 生命周期启动 gRPC
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				listener, err := net.Listen("tcp", ":"+opt.GrpcPort)
				if err != nil {
					logger.Fatal("failed to listen", zap.Error(err))
				}
				logger.Info("starting gRPC server", zap.String("GRPC Port", opt.GrpcPort))
				if err := grpcServer.Serve(listener); err != nil {
					logger.Fatal("gRPC Serve failed", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("stopping gRPC server...")
			grpcServer.GracefulStop()
			return nil
		},
	})

	return grpcServer
}

package hello

import (
	context "context"

	"go.uber.org/zap"
	grpc "google.golang.org/grpc"
)

type HelloService struct {
	log *zap.Logger
	UnimplementedHelloServiceServer
}

func NewHelloService(log *zap.Logger) *HelloService {
	return &HelloService{log: log}
}

func (s *HelloService) SayHello(ctx context.Context, req *HelloRequest) (*HelloReply, error) {
	return &HelloReply{Message: "Hello, " + req.Name}, nil
}

func (s *HelloService) Register(server *grpc.Server) {
	RegisterHelloServiceServer(server, s)
}

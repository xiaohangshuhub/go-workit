package workit

import (
	"github.com/spf13/viper"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	development = "dev"
	testing     = "test"
	production  = "prod"
	port        = "8080"
	g_port      = "50051"
)

// WebApplicationOptions is the struct for web application options
type WebApplicationOptions struct {
	Config    *viper.Viper
	Logger    *zap.Logger
	Container []fx.Option
}

// ServerOptions is the struct for server options
type ServerOptions struct {
	HttpPort          string
	GrpcPort          string
	Environment       string
	UseDefaultRecover bool
	UseDefaultLogger  bool
}

// Environment is the struct for environment variables
type Environment struct {
	IsDevelopment bool
	Env           string
}

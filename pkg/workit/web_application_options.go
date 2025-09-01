package workit

import (
	"github.com/spf13/viper"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// WebApplicationOptions is the struct for web application options
type WebApplicationOptions struct {
	Config    *viper.Viper
	Logger    *zap.Logger
	Container []fx.Option
}

// ServerOptions is the struct for server options
type ServerOptions struct {
	HttpPort    string `mapstructure:"http_port"`
	GrpcPort    string `mapstructure:"grpc_port"`
	Environment string `mapstructure:"environment"`
}

// Environment is the struct for environment variables
type Environment struct {
	IsDevelopment bool
	Env           string
}

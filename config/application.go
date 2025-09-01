package config

import (
	"time"

	"github.com/spf13/viper"
)

type ApplicationConfig struct {
	Server   ServerConfig   `mapstructure:"server"`
	Log      LogConfig      `mapstructure:"log"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Database DatabaseConfig `mapstructure:"database"`
}

type ServerConfig struct {
	HTTPPort    int    `mapstructure:"http_port"`
	GRPCPort    int    `mapstructure:"grpc_port"`
	Environment string `mapstructure:"environment"`
}

type LogConfig struct {
	Level      string `mapstructure:"level"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"maxsize"`
	MaxBackups int    `mapstructure:"maxbackups"`
	MaxAge     int    `mapstructure:"maxage"`
	Compress   bool   `mapstructure:"compress"`
	Console    bool   `mapstructure:"console"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

type DatabaseConfig struct {
	DSN             string        `mapstructure:"dsn"`
	LogLevel        string        `mapstructure:"log_level"`
	SlowThreshold   time.Duration `mapstructure:"slow_threshold"`
	DryRun          bool          `mapstructure:"dry_run"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

func NewApplicationConfig(cfg *viper.Viper) *ApplicationConfig {

	// 绑定配置
	var appCfg ApplicationConfig
	if err := cfg.Unmarshal(&appCfg); err != nil {
		panic(err)
	}

	return &appCfg
}

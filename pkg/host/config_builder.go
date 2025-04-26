package host

import (
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type ConfigBuilder interface {
	AddYamlFile(path string) error
	AddJsonFile(path string) error
	AddEnvironmentVariables(prefix string)
	AddCommandLine(args []string)
	AddConfigFile(path string, fileType string) error
}

type configBuilder struct {
	v *viper.Viper
}

func newConfigBuilder(v *viper.Viper) ConfigBuilder {
	return &configBuilder{v: v}
}

func (c *configBuilder) AddYamlFile(path string) error {
	return c.AddConfigFile(path, "yaml")
}

func (c *configBuilder) AddJsonFile(path string) error {
	return c.AddConfigFile(path, "json")
}

func (c *configBuilder) AddEnvironmentVariables(prefix string) {
	if prefix != "" {
		c.v.SetEnvPrefix(prefix)
	}
	c.v.AutomaticEnv()
	c.v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
}

func (c *configBuilder) AddCommandLine(args []string) {
	flags := pflag.NewFlagSet("app", pflag.ContinueOnError)
	flags.String("port", "", "server port")
	flags.String("env", "", "environment")

	_ = flags.Parse(args)
	_ = c.v.BindPFlags(flags)
}

func (c *configBuilder) AddConfigFile(path string, fileType string) error {
	subViper := viper.New()
	subViper.SetConfigFile(path)
	subViper.SetConfigType(fileType)
	if err := subViper.ReadInConfig(); err != nil {
		return err
	}
	return c.v.MergeConfigMap(subViper.AllSettings())
}

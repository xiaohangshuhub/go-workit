package workit

import (
	"os"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type ConfigBuilder interface {
	AddYamlFile(path string) error
	AddJsonFile(path string) error
	addEnvironmentVariables()
	addCommandLine()
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

func (c *configBuilder) addEnvironmentVariables() {
	c.v.AutomaticEnv()
	c.v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
}

func (c *configBuilder) addCommandLine() {
	flags := pflag.NewFlagSet("app", pflag.ContinueOnError)

	allSettings := c.v.AllSettings()
	for key, value := range allSettings {
		switch v := value.(type) {
		case string:
			flags.String(key, v, "override for "+key)
		case int:
			flags.Int(key, v, "override for "+key)
		case bool:
			flags.Bool(key, v, "override for "+key)
		case float64:
			flags.Float64(key, v, "override for "+key)
		}
	}

	_ = flags.Parse(os.Args[1:])
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

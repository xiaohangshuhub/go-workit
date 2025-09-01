package workit

import (
	"os"
	"strings"

	"github.com/fsnotify/fsnotify"
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
	v         *viper.Viper
	subVipers []*viper.Viper
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

	// 展平所有配置
	allSettings := map[string]interface{}{}
	flattenSettings("", c.v.AllSettings(), allSettings)

	for key, value := range allSettings {
		// 转换 flag 名字：server.http_port → server-http-port
		flagName := strings.ReplaceAll(key, ".", "-")

		switch v := value.(type) {
		case string:
			flags.String(flagName, v, "override for "+key)
		case int:
			flags.Int(flagName, v, "override for "+key)
		case bool:
			flags.Bool(flagName, v, "override for "+key)
		case float64:
			flags.Float64(flagName, v, "override for "+key)
		case []string:
			flags.StringSlice(flagName, v, "override for "+key)
		}

		// 绑定原始 key（server.http_port），而不是 flagName（server-http-port）
		_ = c.v.BindPFlag(key, flags.Lookup(flagName))
	}

	_ = flags.Parse(os.Args[1:])
}

func flattenSettings(prefix string, settings map[string]interface{}, out map[string]interface{}) {
	for k, v := range settings {
		fullKey := k
		if prefix != "" {
			fullKey = prefix + "." + k
		}
		switch child := v.(type) {
		case map[string]interface{}:
			flattenSettings(fullKey, child, out)
		default:
			out[fullKey] = v
		}
	}
}

func (c *configBuilder) AddConfigFile(path string, fileType string) error {
	subViper := viper.New()
	subViper.SetConfigFile(path)
	subViper.SetConfigType(fileType)
	if err := subViper.ReadInConfig(); err != nil {
		return err
	}

	// 设置监听和回调函数
	subViper.WatchConfig()
	subViper.OnConfigChange(func(e fsnotify.Event) {
		if err := subViper.ReadInConfig(); err != nil {
			// 这里可以添加日志输出，例如：log.Printf("Error reading config file: %v", err)
			return
		}
		c.v.MergeConfigMap(subViper.AllSettings())
	})

	// 保存子 Viper 实例以避免被垃圾回收
	c.subVipers = append(c.subVipers, subViper)

	return c.v.MergeConfigMap(subViper.AllSettings())
}

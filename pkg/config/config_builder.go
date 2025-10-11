package config

import (
	"os"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// ConfigBuilder 定义配置构建器接口
type Builder interface {
	AddYamlFile(path string) error
	AddJsonFile(path string) error
	AddEnvironmentVariables()
	AddCommandLine()
	AddConfigFile(path string, fileType string) error
}

// configBuilder 实现 ConfigBuilder 接口
type configBuilder struct {
	v         *viper.Viper
	subVipers []*viper.Viper
	loaded    map[string]bool
}

// newConfigBuilder 创建配置构建器实例
func NewBuilder(v *viper.Viper) Builder {
	return &configBuilder{
		v:      v,
		loaded: make(map[string]bool)}
}

// AddYamlFile 添加 YAML 配置文件
func (c *configBuilder) AddYamlFile(path string) error {
	return c.AddConfigFile(path, "yaml")
}

// AddJsonFile 添加 JSON 配置文件
func (c *configBuilder) AddJsonFile(path string) error {
	return c.AddConfigFile(path, "json")
}

// addEnvironmentVariables 添加环境变量
func (c *configBuilder) AddEnvironmentVariables() {
	c.v.AutomaticEnv()
	c.v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
}

// addCommandLine 添加命令行参数
func (c *configBuilder) AddCommandLine() {
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

// flattenSettings 展平配置
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

// AddConfigFile 添加配置文件
func (c *configBuilder) AddConfigFile(path string, fileType string) error {

	// 避免重复加载同一个文件
	if c.loaded[path] {
		return nil
	}

	subViper := viper.New()
	subViper.SetConfigFile(path)
	subViper.SetConfigType(fileType)

	if err := subViper.ReadInConfig(); err != nil {

		// 如果文件不存在，可以选择忽略
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil // 不报错，继续运行
		}

		panic(err)
	}

	// 设置监听和回调函数
	subViper.WatchConfig()

	subViper.OnConfigChange(func(e fsnotify.Event) {

		if err := subViper.ReadInConfig(); err != nil {
			// 这里可以添加日志输出
			return
		}

		c.v.MergeConfigMap(subViper.AllSettings())
	})

	// 保存子 Viper 实例，避免被 GC
	c.subVipers = append(c.subVipers, subViper)

	// 标记为已加载
	c.loaded[path] = true

	return c.v.MergeConfigMap(subViper.AllSettings())
}

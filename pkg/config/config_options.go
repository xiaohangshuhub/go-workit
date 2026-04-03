package config

type Options struct {
	configBuilder Builder
}

func NewOptions(configBuilder Builder) *Options {

	return &Options{configBuilder: configBuilder}
}

func (co *Options) UseYamlFile(path string) {
	co.configBuilder.AddYamlFile(path)
}

func (co *Options) UseJsonFile(path string) {
	co.configBuilder.AddJsonFile(path)
}

func (co *Options) UseFile(path, fileType string) {
	co.configBuilder.AddConfigFile(path, fileType)
}

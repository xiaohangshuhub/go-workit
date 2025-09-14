package app

type ConfigOptions struct {
	configBuilder ConfigBuilder
}

func newConfigOptions(configBuilder ConfigBuilder) *ConfigOptions {

	return &ConfigOptions{configBuilder: configBuilder}
}

func (co *ConfigOptions) AddYamlFile(path string) error {
	return co.configBuilder.AddYamlFile(path)
}

func (co *ConfigOptions) AddJsonFile(path string) error {
	return co.configBuilder.AddJsonFile(path)
}

func (co *ConfigOptions) AddFile(path, fileType string) error {
	return co.configBuilder.AddConfigFile(path, fileType)
}

package app

type ConfigOptions struct {
	configBuilder ConfigBuilder
}

func newConfigOptions(configBuilder ConfigBuilder) *ConfigOptions {

	return &ConfigOptions{configBuilder: configBuilder}
}

func (co *ConfigOptions) UseYamlFile(path string) error {
	return co.configBuilder.AddYamlFile(path)
}

func (co *ConfigOptions) UseJsonFile(path string) error {
	return co.configBuilder.AddJsonFile(path)
}

func (co *ConfigOptions) UseFile(path, fileType string) error {
	return co.configBuilder.AddConfigFile(path, fileType)
}

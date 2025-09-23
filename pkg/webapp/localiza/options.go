package localiza

type Options struct {
	// 默认语言
	DefaultLanguage string
	// 支持的语言列表
	SupportedLanguages []string
	// 翻译文件目录
	TranslationsDir string
	// 文件类型，支持 "json" 和 "toml"
	FileType LocalizationFileType
}

// NewLocalizerOptions 返回默认的国际化配置
func NewOptions() *Options {

	opts := &Options{
		DefaultLanguage:    "zh-CN",
		SupportedLanguages: []string{"zh-CN", "en-US"},
		TranslationsDir:    "./locales",
	}
	return opts
}

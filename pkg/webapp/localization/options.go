package localization

import "github.com/nicksnyder/go-i18n/v2/i18n"

type Options struct {
	// 默认语言
	DefaultLanguage string
	// 支持的语言列表
	SupportedLanguages []string
	// 翻译文件目录
	TranslationsDir string
	// Bundle 实例
	Bundle *i18n.Bundle
	// 文件类型，支持 "json" 和 "toml"
	FileType LocalizationFileType

	*Builder
}

// NewLocalizerOptions 返回默认的国际化配置
func NewLocalizerOptions() *Options {

	opts := &Options{
		DefaultLanguage:    "zh-CN",
		SupportedLanguages: []string{"zh-CN", "en-US"},
		TranslationsDir:    "locales",
	}

	opts.Builder = NewBuilder(opts)

	return opts
}

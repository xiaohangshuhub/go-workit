package localization

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

type LocalizationFileType string

const (
	LocalizationFileTypeJSON LocalizationFileType = ".json"
	LocalizationFileTypeToml LocalizationFileType = ".toml"
)

// Builder 国际化构建器
type Builder struct {
	defaultLanguage string
	translationsDir string
	fileType        LocalizationFileType
	*Options
}

// NewBuilder 创建国际化构建器
func NewBuilder(options *Options) *Builder {

	return &Builder{

		Options: options,
	}
}

// Build 构建国际化服务
func (b *Builder) Build() (*Provider, error) {

	bundle := i18n.NewBundle(language.Make(b.defaultLanguage))

	if b.fileType == LocalizationFileTypeJSON {
		bundle.RegisterUnmarshalFunc(string(b.fileType), json.Unmarshal)
	}

	if b.fileType == LocalizationFileTypeToml {
		bundle.RegisterUnmarshalFunc(string(b.fileType), toml.Unmarshal)
	}

	// 使用绝对路径进行文件遍历
	err := filepath.Walk(b.translationsDir, func(path string, info os.FileInfo, err error) error {

		if err != nil {
			return err
		}

		if !info.IsDir() && (filepath.Ext(path) == string(LocalizationFileTypeJSON) || filepath.Ext(path) == string(LocalizationFileTypeToml)) {

			_, err = bundle.LoadMessageFile(path)

			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return newProvider(b.defaultLanguage, *b.Bundle), nil
}

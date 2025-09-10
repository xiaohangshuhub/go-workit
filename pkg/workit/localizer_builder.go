package workit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

// I18nBuilder 国际化构建器
type LocalizerBuilder struct {
	defaultLanguage    string
	supportedLanguages []string
	translationsDir    string
}

// NewI18nBuilder 创建国际化构建器
func newLocalizerBuilder(defaultLanguage string, supportedLanguages []string, translationsDir string) *LocalizerBuilder {
	return &LocalizerBuilder{
		defaultLanguage:    defaultLanguage,
		supportedLanguages: supportedLanguages,
		translationsDir:    translationsDir,
	}
}

// Build 构建国际化服务
func (b *LocalizerBuilder) Build() (*i18n.Bundle, error) {
	// 获取工作目录
	wd, err := os.Getwd()
	if err != nil {

		return nil, fmt.Errorf("获取工作目录失败: %w", err)
	}

	// 构建本地化文件目录的绝对路径
	localesPath := filepath.Join(wd, b.translationsDir)

	// 确保路径存在
	if _, err := os.Stat(localesPath); os.IsNotExist(err) {

		return nil, fmt.Errorf("本地化目录不存在: %s", localesPath)
	}

	// 创建 bundle
	bundle := i18n.NewBundle(language.Make(b.defaultLanguage))
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	// 使用绝对路径进行文件遍历
	err = filepath.Walk(localesPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {

			return fmt.Errorf("遍历路径错误 %s: %w", path, err)
		}

		// 只处理 json 文件
		if !info.IsDir() && filepath.Ext(path) == ".json" {

			_, err = bundle.LoadMessageFile(path)
			if err != nil {

				return fmt.Errorf("加载翻译文件失败 %s: %w", path, err)
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return bundle, nil
}

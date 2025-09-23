package web

import "github.com/nicksnyder/go-i18n/v2/i18n"

type Localization interface {
	DefaultLanguage() string //默认语言
	Bundle() *i18n.Bundle
}

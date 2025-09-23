package localiza

import "github.com/nicksnyder/go-i18n/v2/i18n"

type Provider struct {
	defaultLanguage string       //默认语言
	bundle          *i18n.Bundle // 国际化
}

func newProvider(defaultLanguage string, bundle *i18n.Bundle) *Provider {
	return &Provider{
		defaultLanguage: defaultLanguage,
		bundle:          bundle,
	}
}

func (p *Provider) DefaultLanguage() string {
	return p.defaultLanguage
}

func (p *Provider) Bundle() *i18n.Bundle {
	return p.bundle
}

package reqdecp

import (
	"fmt"

	"github.com/xiaohangshuhub/go-workit/pkg/webapp/reqdecp/br"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/reqdecp/deflate"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/reqdecp/gzip"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/web"
)

// Options 管理所有请求解压提供者
type Options struct {
	providers map[string]web.ReqDecompression
}

// NewOptions 初始化 Options 并注册默认 provider
func NewOptions() *Options {
	opts := &Options{
		providers: make(map[string]web.ReqDecompression),
	}

	// 注册默认提供者
	opts.DecompressionProvider("br", br.New())
	opts.DecompressionProvider("deflate", deflate.New())
	opts.DecompressionProvider("gzip", gzip.New())

	return opts
}

// RegisterProvider 注册一个解压提供者，如果已存在则 panic
func (opts *Options) DecompressionProvider(name string, provider web.ReqDecompression) {
	if _, exists := opts.providers[name]; exists {
		panic(fmt.Errorf("provider 已存在: %s", name))
	}

	opts.providers[name] = provider
}

func (opts *Options) Decompressions() map[string]web.ReqDecompression {

	return opts.providers
}

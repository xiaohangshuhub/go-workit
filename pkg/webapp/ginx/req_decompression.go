package ginx

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xiaohangshu-dev/go-workit/pkg/webapp/web"
	"go.uber.org/zap"
)

// Authorize 授权中间件
type ReqDecompression struct {
	web.ReqDecompressor
	logger *zap.Logger
}

// newAuthorize 初始化授权中间件
func newDecompression(dec web.ReqDecompressor, logger *zap.Logger) *ReqDecompression {
	return &ReqDecompression{
		ReqDecompressor: dec,
		logger:          logger,
	}
}

// Handle 授权中间件处理函数
func (m *ReqDecompression) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		encoding := strings.ToLower(c.GetHeader("Content-Encoding"))
		if encoding == "" {
			// 没有压缩，直接继续
			c.Next()
			return
		}

		provider, ok := m.ReqDecompressor.Decompression(encoding)
		if !ok {
			m.logger.Warn("unsupported content-encoding", zap.String("encoding", encoding))
			c.AbortWithStatusJSON(http.StatusUnsupportedMediaType, gin.H{
				"error": "unsupported content-encoding",
			})
			return
		}

		reader, err := provider.Decompression(c.Request.Body)
		if err != nil {
			m.logger.Error("failed to create decompression stream",
				zap.String("encoding", encoding),
				zap.Error(err),
			)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "invalid compressed body",
			})
			return
		}

		// 替换请求体为解压流
		c.Request.Body = reader

		// 删除 Content-Encoding 头，避免后续 handler 误判
		c.Request.Header.Del("Content-Encoding")

		// 继续执行
		c.Next()
	}
}

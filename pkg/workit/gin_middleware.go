package workit

import "github.com/gin-gonic/gin"

type GinMiddleware interface {
	Handle() gin.HandlerFunc
	ShouldSkip(path string, method string) bool
}

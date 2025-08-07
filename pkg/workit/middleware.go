package workit

import "github.com/gin-gonic/gin"

type Middleware interface {
	Handle() gin.HandlerFunc
	ShouldSkip(path string) bool
}

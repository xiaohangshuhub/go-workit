package gin

import "github.com/gin-gonic/gin"

// Middleware  gin middleware interface
type Middleware interface {
	Handle() gin.HandlerFunc
}

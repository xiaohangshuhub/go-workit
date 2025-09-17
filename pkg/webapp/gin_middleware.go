package webapp

import "github.com/gin-gonic/gin"

// GinMiddleware  gin middleware interface
type GinMiddleware interface {
	Handle() gin.HandlerFunc
}

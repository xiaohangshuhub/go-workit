package webapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ResponseWithData 用来在Swagger里指定Data的具体类型
type Response[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

// Success 返回成功
func Success[T any](c *gin.Context, data T) {
	c.JSON(http.StatusOK, Response[T]{
		Code:    0,
		Message: "OK",
		Data:    data,
	})
}

// Fail 返回失败
func Fail(c *gin.Context, code int, message string) {
	c.JSON(code, Response[any]{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

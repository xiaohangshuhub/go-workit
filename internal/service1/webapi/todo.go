package webapi

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Hello(
	router *gin.Engine, //gin 放在第一位
	log *zap.Logger, // 日志
) {

	// 创建路由组
	group := router.Group("/hello")

	// 创建路由

	group.GET("", HelloNewb(log))
}

// HelloNewb godoc
// @Summary 你好, Newb
// @Description 返回 "hello newb"
// @Tags Hello
// @Accept json
// @Produce json
// @Success 200 {object} Response[string]
// @Router /hello [get]
func HelloNewb(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		data := Success("hello newb")
		c.JSON(200, data)
	}
}

// ResponseWithData 用来在Swagger里指定Data的具体类型
type Response[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

// Success 返回成功
func Success[T any](data T) Response[T] {
	return Response[T]{
		Code:    0,
		Message: "OK",
		Data:    data,
	}
}

// Fail 返回失败
func Fail(code int, message string) Response[any] {
	return Response[any]{
		Code:    code,
		Message: message,
		Data:    nil,
	}
}

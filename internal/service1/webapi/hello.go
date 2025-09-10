package webapi

import (
	"cli-template/internal/service1/webapi/response"

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

// Hello xiaohangshu godoc
// @Summary hello Newb
// @Description 返回 "hello newb"
// @Tags Hello
// @Accept json
// @Produce json
// @Success 200 {object} api.Response[string]
// @Router /hello [get]
// @Security BearerAuth
func HelloNewb(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		data := response.Success("你好,小航书")
		c.JSON(200, data)
	}
}

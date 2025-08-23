package webapi

import (
	"github.com/gin-gonic/gin"
	"github.com/xiaohangshuhub/go-workit/pkg/api"
	"go.uber.org/zap"
)

func Hello(
	router *gin.Engine, //gin 放在第一位
	log *zap.Logger, // 日志
) {

	// 创建路由组
	group := router.Group("/hello")

	// 创建路由

	group.GET("", helloNewb(log))
}

// HelloNewb godoc
// @Summary hello workit
// @Description 返回 "你好 小航书"
// @Tags Hello
// @Accept json
// @Produce json
// @Success 200 {object} api.Response[string]
// @Router /hello [get]
// @Security BearerAuth
func helloNewb(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		data := api.Success("你好,小航书")
		c.JSON(200, data)
	}
}

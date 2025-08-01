package webapi

import (
	"github.com/labstack/echo/v4"
	"github.com/xiaohangshuhub/go-workit/pkg/api"
	"go.uber.org/zap"
)

func Hello(
	router *echo.Echo, //echo
	log *zap.Logger, // 日志
) {

	// 创建路由组
	group := router.Group("/hello")

	// 创建路由

	group.GET("", HelloNewb(log))
}

// HelloNewb godoc
// @Summary hello workit
// @Description 返回 "你好 小航书"
// @Tags Hello
// @Accept json
// @Produce json
// @Success 200 {object} api.Response[string]
// @Router /hello [get]
func HelloNewb(log *zap.Logger) echo.HandlerFunc {
	return func(c echo.Context) error {
		data := api.Success("你好,小航书")
		return c.JSON(200, data)
	}
}

# webapi

```
func RegisterTodoRoutes(
	router *gin.Engine, //gin 放在第一位
	create *todoapp.CreateTodoCommandHandler) {
	router.POST("/todos", CreateTodoHandler(create))
}

// CreateTodoHandler godoc
// @Summary 创建Todo
// @Description 创建新的待办事项
// @Tags Todos
// @Accept json
// @Produce json
// @Param data body todoapp.CreateTodoCommand true "请求参数"
// @Success 200 {object} Response[todoapp.CreateTodoResult]
// @Failure 400 {object} Response[any]
// @Failure 500 {object} Response[any]
// @Router /todos [post]
func CreateTodoHandler(handler *todoapp.CreateTodoCommandHandler) gin.HandlerFunc {
	return func(c *gin.Context) {

		var cmd todoapp.CreateTodoCommand

		if err := c.ShouldBindJSON(&cmd); err != nil {
			Fail(c, 400, "参数错误: "+err.Error())
			return
		}

		result, err := handler.Handle(cmd)

		if err != nil {
			Fail(c, 500, "创建失败: "+err.Error())
			return
		}
		Success(c, result)
	}
}
```

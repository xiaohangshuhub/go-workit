# WebAPI 层

WebAPI 层负责处理客户端 HTTP 请求，是系统的最外层接口，主要作用是将外部请求转发到应用服务，并统一返回标准格式的响应。

## 职责

- 接收并验证外部请求数据（如参数校验、格式解析等）。
- 调用应用层服务处理业务逻辑。
- 返回标准化的 API 响应。
- 处理异常情况并反馈错误信息。

## 组成

WebAPI 层目前包含以下核心组件：

1. **统一响应结构（Response）**
   封装 HTTP 响应格式，支持泛型返回，使接口定义清晰一致。
   示例：成功返回时统一包含 `code`、`message`、`data` 字段。
2. **Todo 接口路由（Todo API）**
   提供待办事项（Todo）创建的示例接口，示范如何绑定请求、处理响应以及接口注释（Swagger）。

## 设计原则

- **统一标准**：所有返回数据结构保持一致，方便前端处理和接口文档生成。
- **松耦合**：仅依赖应用层（Application Layer）暴露的服务，不直接处理业务逻辑。
- **简洁清晰**：每个接口方法职责单一，易于阅读和维护。
- **可扩展性**：方便后续增加更多模块和接口，支持扩展性开发。

## 示例

### 统一响应结构（Response）

```go
type Response[T any] struct {
    Code    int    `json:"code"`    // 响应码
    Message string `json:"message"` // 响应信息
    Data    T      `json:"data"`    // 响应数据
}

// 返回成功
func Success[T any](c *gin.Context, data T) { ... }

// 返回失败
func Fail(c *gin.Context, code int, message string) { ... }
```

通过泛型支持任意类型的数据返回，提高复用性和类型安全。

### 创建 Todo 示例接口

```go
func RegisterTodoRoutes(router *gin.Engine, create *todoapp.CreateTodoCommandHandler) {
    router.POST("/todos", CreateTodoHandler(create))
}

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

支持通过 Swagger 注解自动生成接口文档。

### Swagger 注释示例

```go
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
```

### 生成接口文档

使用 [swaggo/swag](https://github.com/swaggo/swag) 工具自动生成 API 文档：

```bash
swag init -g main.go
```

或者指定入口文件：

```bash
swag init -g path/to/your/main.go
```

访问接口文档：

```
http://localhost:8080/swagger/index.html
```

## 注意事项

- WebAPI 层只负责请求与响应，不直接嵌入业务逻辑。
- 错误码和错误信息应统一规范，便于前端统一处理。
- 参数校验应放在接收请求时立即完成，确保进入应用层的数据有效。
- Swagger 注释应保持更新，确保文档准确性。

通过标准化的 WebAPI 层设计，可以提升系统接口的一致性、可维护性和对外开放能力。

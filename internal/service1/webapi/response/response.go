package response

const (
	// 成功
	CodeSuccess = 0

	// 客户端错误（4xx）
	CodeBadRequest       = 400
	CodeUnauthorized     = 401
	CodeForbidden        = 403
	CodeNotFound         = 404
	CodeMethodNotAllowed = 405
	CodeRequestTimeout   = 408
	CodeConflict         = 409
	CodeTooManyRequests  = 429

	// 服务端错误（5xx）
	CodeInternalServerError = 500
	CodeNotImplemented      = 501
	CodeBadGateway          = 502
	CodeServiceUnavailable  = 503
	CodeGatewayTimeout      = 504
)

// ResponseWithData 用来在Swagger里指定Data的具体类型
type Response[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

// Success 返回成功
func Success[T any](data T) Response[T] {
	return Response[T]{
		Code:    CodeSuccess,
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

// --- 4xx 错误响应 ---

func BadRequest(msg ...string) Response[any] {
	return Fail(CodeBadRequest, pickMessage("请求参数不合法", msg...))
}

func Unauthorized(msg ...string) Response[any] {
	return Fail(CodeUnauthorized, pickMessage("未授权，请登录", msg...))
}

func Forbidden(msg ...string) Response[any] {
	return Fail(CodeForbidden, pickMessage("没有权限访问", msg...))
}

func NotFound(msg ...string) Response[any] {
	return Fail(CodeNotFound, pickMessage("资源不存在", msg...))
}

func MethodNotAllowed(msg ...string) Response[any] {
	return Fail(CodeMethodNotAllowed, pickMessage("请求方法不被允许", msg...))
}

func RequestTimeout(msg ...string) Response[any] {
	return Fail(CodeRequestTimeout, pickMessage("请求超时", msg...))
}

func Conflict(msg ...string) Response[any] {
	return Fail(CodeConflict, pickMessage("资源冲突或已存在", msg...))
}

func TooManyRequests(msg ...string) Response[any] {
	return Fail(CodeTooManyRequests, pickMessage("请求过于频繁", msg...))
}

// --- 5xx 错误响应 ---

func InternalServerError(msg ...string) Response[any] {
	return Fail(CodeInternalServerError, pickMessage("服务器内部错误", msg...))
}

func NotImplemented(msg ...string) Response[any] {
	return Fail(CodeNotImplemented, pickMessage("接口尚未实现", msg...))
}

func BadGateway(msg ...string) Response[any] {
	return Fail(CodeBadGateway, pickMessage("网关错误", msg...))
}

func ServiceUnavailable(msg ...string) Response[any] {
	return Fail(CodeServiceUnavailable, pickMessage("服务暂不可用", msg...))
}

func GatewayTimeout(msg ...string) Response[any] {
	return Fail(CodeGatewayTimeout, pickMessage("网关超时", msg...))
}

// pickMessage 默认 message 与自定义 message 处理
func pickMessage(defaultMsg string, custom ...string) string {
	if len(custom) > 0 {
		return custom[0]
	}
	return defaultMsg
}

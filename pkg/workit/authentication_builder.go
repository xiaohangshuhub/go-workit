package workit

// 鉴权构建器
type AuthenticationBuilder struct {
	schemes map[string]AuthenticationHandler
}

// 摘要:
//  - NewAuthentication:创建一个新的鉴权构建器
//
// 返回值:
// 	- *AuthenticationBuilder:返回新的鉴权构建器
func newAuthenticationBuilder() *AuthenticationBuilder {
	return &AuthenticationBuilder{
		schemes: make(map[string]AuthenticationHandler),
	}
}

// 摘要:
// 	- AddSchema:添加鉴权方案
//
//参数:
// 	- handler:鉴权方案
//
// 返回值:
// 	- *AuthenticationBuilder:返回当前构建器
func (b *AuthenticationBuilder) AddScheme(handler AuthenticationHandler) *AuthenticationBuilder {
	b.schemes[handler.Scheme()] = handler
	return b
}

// 摘要:
// 	- Schemes:获取所有鉴权方案
//
// 返回值:
// 	- map[string]AuthenticationHandler:返回所有鉴权方案
func (b *AuthenticationBuilder) Schemes() map[string]AuthenticationHandler {
	return b.schemes
}

func (b *AuthenticationBuilder) AddJwtBearer(fn func(*JwtBearerOptions)) *AuthenticationBuilder {

	options := NewJwtBearerOptions()

	fn(options)

	b.AddScheme(NewJWTBearerHandler(options))

	return b
}

func (b *AuthenticationBuilder) Build() *AuthenticateProvider {
	return newAuthenticateProvider(b.schemes)
}

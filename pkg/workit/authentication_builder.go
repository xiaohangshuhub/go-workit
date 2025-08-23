package workit

// 鉴权构建器
type AuthenticationBuilder struct {
	schemes map[string]AuthenticationHandler
}

// 摘要:
//   - NewAuthentication:创建一个新的鉴权构建器
//
// 返回值:
//   - *AuthenticationBuilder:返回新的鉴权构建器
func newAuthenticationBuilder() *AuthenticationBuilder {
	return &AuthenticationBuilder{
		schemes: make(map[string]AuthenticationHandler),
	}
}

// 摘要:
//   - AddSchema:添加鉴权方案
//
// 参数:
//   - handler:鉴权方案
//
// 返回值:
//   - *AuthenticationBuilder:返回当前构建器
func (b *AuthenticationBuilder) AddScheme(schemeName string, handler AuthenticationHandler) *AuthenticationBuilder {

	if _, ok := b.schemes[schemeName]; ok {
		panic("scheme already exists:" + schemeName)
	}

	b.schemes[schemeName] = handler
	return b
}

// Schemes  返回所有注册的鉴权方案
func (b *AuthenticationBuilder) Schemes() map[string]AuthenticationHandler {
	return b.schemes
}

// AddJwtBearer  注册新的 schemename JWT Bearer鉴权方案
func (b *AuthenticationBuilder) AddJwtBearer(schemeName string, fn func(*JwtBearerOptions)) *AuthenticationBuilder {

	options := newJwtBearerOptions()

	fn(options)

	b.AddScheme(schemeName, newJWTBearerHandler(options))

	return b
}

// AddCookie  注册新的 schemename Cookie鉴权方案
func (b *AuthenticationBuilder) AddCookie(schemeName string, fn func(*CookieOptions)) *AuthenticationBuilder {

	options := newCookieOptions()

	fn(options)

	b.AddScheme(schemeName, newCookieHandler(options))

	return b
}

// Build  构建鉴权提供者
func (b *AuthenticationBuilder) Build() *AuthenticateProvider {
	return newAuthenticateProvider(b.schemes)
}

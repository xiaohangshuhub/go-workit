package workit

// AuthenticationBuilder 鉴权构建器
type AuthenticationBuilder struct {
	schemes map[string]AuthenticationHandler
}

// NewAuthenticationBuilder 新建鉴权构建器
func newAuthenticationBuilder() *AuthenticationBuilder {
	return &AuthenticationBuilder{
		schemes: make(map[string]AuthenticationHandler),
	}
}

// AddScheme 注册新的鉴权方案
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

	b.AddScheme(schemeName, newCookieAuthentication(options))

	return b
}

// Build  构建鉴权提供者
func (b *AuthenticationBuilder) Build() *AuthenticateProvider {
	return newAuthenticateProvider(b.schemes)
}

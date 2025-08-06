package workit

import (
	"net/http"
	"time"
)

// TokenHandler 接口，用于验证和解析 JWT
type TokenHandler interface {
	ValidateToken(tokenString string) (map[string]interface{}, error)
}

// TokenValidationParameters 对应 .NET 的 TokenValidationParameters
type TokenValidationParameters struct {
	ValidAudience     string
	ValidIssuer       string
	RequireExpiration bool
	ClockSkew         time.Duration
	SigningKey        interface{}
}

// OpenIdConnectConfiguration 占位类型，对应 .NET OpenIdConnectConfiguration
type OpenIdConnectConfiguration struct {
	Issuer  string
	JWKSUri string
}

// ConfigurationManager 接口，对应 .NET IConfigurationManager
type ConfigurationManager interface {
	GetConfiguration() (*OpenIdConnectConfiguration, error)
	Refresh() error
}

// JwtBearerEvents 事件处理钩子
type JwtBearerEvents struct {
	OnTokenValidated       func(claims map[string]interface{}) error
	OnAuthenticationFailed func(err error) error
}

type JwtBearerOptions struct {
	RequireHttpsMetadata       bool
	MetadataAddress            string
	Authority                  string
	Audience                   string
	Challenge                  string
	Events                     *JwtBearerEvents
	BackchannelHttpHandler     http.Handler
	Backchannel                *http.Client
	BackchannelTimeout         time.Duration
	Configuration              *OpenIdConnectConfiguration
	ConfigurationManager       ConfigurationManager
	RefreshOnIssuerKeyNotFound bool
	TokenHandlers              []TokenHandler
	TokenValidationParameters  TokenValidationParameters
	SaveToken                  bool
	IncludeErrorDetails        bool
	MapInboundClaims           bool
	AutomaticRefreshInterval   time.Duration
	RefreshInterval            time.Duration
	UseSecurityTokenValidators bool
}

// NewJwtBearerOptions 初始化默认值
func NewJwtBearerOptions() *JwtBearerOptions {
	return &JwtBearerOptions{
		RequireHttpsMetadata:       true,
		Challenge:                  "Bearer",
		Backchannel:                http.DefaultClient,
		BackchannelTimeout:         time.Minute,
		RefreshOnIssuerKeyNotFound: true,
		SaveToken:                  true,
		IncludeErrorDetails:        true,
		MapInboundClaims:           true,
		AutomaticRefreshInterval:   24 * time.Hour,
		RefreshInterval:            5 * time.Minute,
	}
}

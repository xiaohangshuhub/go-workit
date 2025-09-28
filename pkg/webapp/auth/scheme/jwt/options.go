package jwt

import (
	"net/http"
	"sync"
	"time"

	"github.com/xiaohangshuhub/go-workit/pkg/webapp/web"
)

// TokenValidationParameters token验证参数
type TokenValidationParameters struct {
	ValidAudience            string
	ValidIssuer              string
	RequireExpiration        bool
	ClockSkew                time.Duration
	SigningKey               []byte         // 单个签名密钥（对称或非对称）
	signingKeys              map[string]any // kid -> key 多个密钥
	ValidateIssuer           bool
	ValidateAudience         bool
	ValidateLifetime         bool
	ValidateIssuerSigningKey bool
	RequireExpirationTime    bool
}

// JwtBearerEvents jwt事件
type JwtBearerEvents struct {
	OnMessageReceived      func(r *http.Request) (string, error)
	OnTokenValidated       func(principal *web.ClaimsPrincipal) error
	OnAuthenticationFailed func(err error) error
	OnChallenge            func(w http.ResponseWriter, r *http.Request, err error)
}

// Options jwt 选项
type Options struct {
	RequireHttpsMetadata       bool
	MetadataAddress            string
	Authority                  string
	Audience                   string
	Challenge                  string
	Events                     *JwtBearerEvents
	BackchannelHttpClient      *http.Client
	BackchannelTimeout         time.Duration
	RefreshOnIssuerKeyNotFound bool
	TokenValidationParameters  TokenValidationParameters
	SaveToken                  bool
	IncludeErrorDetails        bool
	MapInboundClaims           bool
	AutomaticRefreshInterval   time.Duration
	RefreshInterval            time.Duration

	configMu     sync.RWMutex
	openIDConfig *OpenIDConfig
	jwksCache    map[string]interface{}
	jwksMu       sync.RWMutex
}

// NewOptions 创建一个新的Options 实例
func NewOptions() *Options {
	return &Options{
		RequireHttpsMetadata:       true,
		Challenge:                  "Bearer",
		BackchannelHttpClient:      http.DefaultClient,
		BackchannelTimeout:         time.Minute,
		RefreshOnIssuerKeyNotFound: true,
		SaveToken:                  true,
		IncludeErrorDetails:        true,
		MapInboundClaims:           true,
		AutomaticRefreshInterval:   24 * time.Hour,
		RefreshInterval:            5 * time.Minute,
		jwksCache:                  make(map[string]interface{}),
		TokenValidationParameters:  TokenValidationParameters{},
	}
}

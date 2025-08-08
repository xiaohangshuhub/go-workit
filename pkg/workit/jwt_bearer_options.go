package workit

import (
	"net/http"
	"sync"
	"time"
)

type TokenValidationParameters struct {
	ValidAudience            string
	ValidIssuer              string
	RequireExpiration        bool
	ClockSkew                time.Duration
	SigningKey               []byte                 // 单个签名密钥（对称或非对称）
	signingKeys              map[string]interface{} // kid -> key 多个密钥
	ValidateIssuer           bool
	ValidateAudience         bool
	ValidateLifetime         bool
	ValidateIssuerSigningKey bool
	RequireExpirationTime    bool
}

type JwtBearerEvents struct {
	OnMessageReceived      func(r *http.Request) (string, error)
	OnTokenValidated       func(principal *ClaimsPrincipal) error
	OnAuthenticationFailed func(err error) error
	OnChallenge            func(w http.ResponseWriter, r *http.Request, err error)
}

type JwtBearerOptions struct {
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

func NewJwtBearerOptions() *JwtBearerOptions {
	return &JwtBearerOptions{
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

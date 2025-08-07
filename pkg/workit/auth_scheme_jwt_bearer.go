package workit

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const SchemeJwtBearer = "JwtBearer"

type JWTBearerHandler struct {
	Options *JwtBearerOptions
}

func NewJWTBearerHandler(options *JwtBearerOptions) *JWTBearerHandler {
	return &JWTBearerHandler{Options: options}
}
func (h *JWTBearerHandler) Scheme() string {
	return SchemeJwtBearer
}

func (h *JWTBearerHandler) Authenticate(r *http.Request) (*ClaimsPrincipal, error) {
	// 先拿 token
	tokenString, err := h.extractToken(r)
	if err != nil {
		h.invokeAuthenticationFailed(err)
		return nil, err
	}
	if tokenString == "" {
		err = errors.New("token not found")
		h.invokeAuthenticationFailed(err)
		return nil, err
	}

	// 确保 OpenID Config 和 JWKS
	err = h.ensureConfigAndKeys()
	if err != nil {
		h.invokeAuthenticationFailed(err)
		return nil, err
	}

	// 验证 token
	principal, err := h.validateToken(tokenString)
	if err != nil {
		// key not found 时尝试刷新一次
		if h.Options.RefreshOnIssuerKeyNotFound && strings.Contains(err.Error(), "key is invalid") {
			_ = h.refreshConfig()
			principal, err = h.validateToken(tokenString)
		}
		if err != nil {
			h.invokeAuthenticationFailed(err)
			return nil, err
		}
	}

	// 触发 OnTokenValidated 事件
	if h.Options.Events != nil && h.Options.Events.OnTokenValidated != nil {
		if err := h.Options.Events.OnTokenValidated(principal); err != nil {
			return nil, err
		}
	}

	return principal, nil
}

func (h *JWTBearerHandler) extractToken(r *http.Request) (string, error) {
	// 先通过自定义事件取 token（支持特殊场景）
	if h.Options.Events != nil && h.Options.Events.OnMessageReceived != nil {
		token, err := h.Options.Events.OnMessageReceived(r)
		if err != nil {
			return "", err
		}
		if token != "" {
			return token, nil
		}
	}

	auth := r.Header.Get("Authorization")
	if auth == "" {
		return "", nil
	}

	// 不区分大小写判断 Bearer 前缀
	if !strings.HasPrefix(strings.ToLower(auth), "bearer ") {
		return "", errors.New("authorization header missing Bearer prefix")
	}

	token := strings.TrimSpace(auth[len("Bearer "):])
	return token, nil
}

func (h *JWTBearerHandler) ensureConfigAndKeys() error {
	h.Options.configMu.RLock()
	cfg := h.Options.openIDConfig
	h.Options.configMu.RUnlock()

	if cfg == nil || time.Now().After(cfg.Expires) {
		if err := h.Options.FetchOpenIDConfig(); err != nil {
			return err
		}
		if err := h.Options.FetchJWKS(); err != nil {
			return err
		}
	}
	return nil
}

func (h *JWTBearerHandler) validateToken(tokenString string) (*ClaimsPrincipal, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		// 先用单个SigningKey
		if h.Options.TokenValidationParameters.SigningKey != nil {
			return h.Options.TokenValidationParameters.SigningKey, nil
		}

		// 需要 kid 来从 SigningKeys 中找
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, errors.New("kid header not found")
		}

		h.Options.jwksMu.RLock()
		key, exists := h.Options.TokenValidationParameters.SigningKeys[kid]
		h.Options.jwksMu.RUnlock()
		if !exists {
			return nil, errors.New("key not found for kid")
		}
		return key, nil
	}

	parser := jwt.NewParser(jwt.WithValidMethods([]string{"RS256", "HS256"})) // 允许RS256和HS256

	claims := jwt.MapClaims{}
	token, err := parser.ParseWithClaims(tokenString, claims, keyFunc)
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("token is invalid")
	}

	// 验证参数
	params := h.Options.TokenValidationParameters

	// 验证签名密钥
	if params.ValidateIssuerSigningKey && (params.SigningKey == nil && len(params.SigningKeys) == 0) {
		return nil, errors.New("no signing key configured")
	}

	// 验证过期时间
	if params.RequireExpiration {
		if expRaw, ok := claims["exp"]; ok {
			expFloat, ok := expRaw.(float64)
			if !ok {
				return nil, errors.New("invalid exp claim")
			}
			expTime := time.Unix(int64(expFloat), 0)
			if time.Now().After(expTime.Add(params.ClockSkew)) {
				return nil, errors.New("token expired")
			}
		} else if params.RequireExpirationTime {
			return nil, errors.New("expiration required but not present")
		}
	}

	// 验证 Audience
	if params.ValidateAudience {
		audClaim, audOk := claims["aud"]
		if !audOk {
			return nil, errors.New("audience claim missing")
		}

		audienceValid := false
		switch aud := audClaim.(type) {
		case string:
			audienceValid = aud == params.ValidAudience
		case []interface{}:
			for _, a := range aud {
				if s, ok := a.(string); ok && s == params.ValidAudience {
					audienceValid = true
					break
				}
			}
		default:
			return nil, errors.New("invalid aud claim type")
		}

		if !audienceValid {
			return nil, errors.New("audience invalid")
		}
	}

	// 验证 Issuer
	if params.ValidateIssuer {
		issClaim, issOk := claims["iss"].(string)
		if !issOk || issClaim != params.ValidIssuer {
			return nil, errors.New("issuer invalid")
		}
	}

	// 构造 ClaimsPrincipal
	principal := &ClaimsPrincipal{
		Claims: make(map[string]interface{}),
	}

	if sub, ok := claims["sub"].(string); ok {
		principal.Name = sub
		principal.AddClaim("sub", sub)
	}

	for k, v := range claims {
		principal.AddClaim(k, v)
	}

	return principal, nil
}

func (h *JWTBearerHandler) invokeAuthenticationFailed(err error) {
	if h.Options.Events != nil && h.Options.Events.OnAuthenticationFailed != nil {
		_ = h.Options.Events.OnAuthenticationFailed(err)
	}
}

func (h *JWTBearerHandler) refreshConfig() error {
	if err := h.Options.FetchOpenIDConfig(); err != nil {
		return err
	}
	return h.Options.FetchJWKS()
}

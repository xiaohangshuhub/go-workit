package workit

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const SchemeJwtBearer = "JwtBearer"

// JWTBearerHandler 处理 JWT Bearer 认证
type JWTBearerAuthenticationHandler struct {
	Options *JwtBearerOptions
}

func newJWTBearerHandler(options *JwtBearerOptions) *JWTBearerAuthenticationHandler {
	return &JWTBearerAuthenticationHandler{Options: options}
}
func (h *JWTBearerAuthenticationHandler) Scheme() string {
	return SchemeJwtBearer
}

// Authenticate 实现 AuthenticationHandler 接口
func (h *JWTBearerAuthenticationHandler) Authenticate(r *http.Request) (*ClaimsPrincipal, error) {
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
		signingKey := h.Options.TokenValidationParameters.SigningKey
		if len(signingKey) == 0 {
			h.invokeAuthenticationFailed(err)
			return nil, err
		}
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

// 取 token
func (h *JWTBearerAuthenticationHandler) extractToken(r *http.Request) (string, error) {
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

// 确保配置和密钥
func (h *JWTBearerAuthenticationHandler) ensureConfigAndKeys() error {
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

// 验证token
func (h *JWTBearerAuthenticationHandler) validateToken(tokenString string) (*ClaimsPrincipal, error) {
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
		key, exists := h.Options.TokenValidationParameters.signingKeys[kid]
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
	if params.ValidateIssuerSigningKey && (params.SigningKey == nil && len(params.signingKeys) == 0) {
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

	// ✅ 构建 ClaimsPrincipal
	principal := &ClaimsPrincipal{
		Claims: make([]Claim, 0, len(claims)),
	}

	// sub
	if sub, ok := claims["sub"].(string); ok {
		principal.Subject = sub
		principal.Name = sub
	}

	// iat
	if iatRaw, ok := claims["iat"].(float64); ok {
		principal.AuthenticatedAt = time.Unix(int64(iatRaw), 0)
	}

	// IdentityProvider 从 "idp" 或 "iss" 提取
	if idp, ok := claims["idp"].(string); ok {
		principal.IdentityProvider = idp
	} else if iss, ok := claims["iss"].(string); ok {
		principal.IdentityProvider = iss
	}

	// 解析 role(s)
	principal.Roles = extractRolesFromClaims(claims)

	// 添加其余 claim
	for k, v := range claims {
		principal.AddClaim(k, v)
	}

	return principal, nil
}

// 触发 OnAuthenticationFailed 事件
func (h *JWTBearerAuthenticationHandler) invokeAuthenticationFailed(err error) {
	if h.Options.Events != nil && h.Options.Events.OnAuthenticationFailed != nil {
		_ = h.Options.Events.OnAuthenticationFailed(err)
	}
}

// 刷新 OpenID Config 和 JWKS
func (h *JWTBearerAuthenticationHandler) refreshConfig() error {
	if err := h.Options.FetchOpenIDConfig(); err != nil {
		return err
	}
	return h.Options.FetchJWKS()
}

// 从 claims 中提取角色，支持多种格式
func extractRolesFromClaims(claims jwt.MapClaims) []string {
	var result []string
	roleKeys := []string{
		"role",
		"roles",
		Role,
	}

	for _, key := range roleKeys {
		if raw, ok := claims[key]; ok {
			switch val := raw.(type) {
			case string:
				result = append(result, val)
			case []interface{}:
				for _, item := range val {
					if s, ok := item.(string); ok {
						result = append(result, s)
					}
				}
			}
		}
	}

	// 可选：兼容 Keycloak
	if realmAccess, ok := claims["realm_access"].(map[string]interface{}); ok {
		if roles, ok := realmAccess["roles"].([]interface{}); ok {
			for _, role := range roles {
				if s, ok := role.(string); ok {
					result = append(result, s)
				}
			}
		}
	}

	return result
}

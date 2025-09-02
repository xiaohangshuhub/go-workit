package workit

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// OpenIDConfig represents the OpenID Connect configuration for an issuer.
type OpenIDConfig struct {
	Issuer  string `json:"issuer"`
	JwksURI string `json:"jwks_uri"`
	Expires time.Time
}

// JSONWebKey represents a JSON Web Key (JWK) as defined in RFC 7517.
type JSONWebKey struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
	Alg string `json:"alg"`
}

// JSONWebKeySet represents a JSON Web Key Set (JWKS) as defined in RFC 7517.
type JSONWebKeySet struct {
	Keys []JSONWebKey `json:"keys"`
}

// FetchOpenIDConfig fetches the OpenID Connect configuration for an issuer.
func (j *JwtBearerOptions) FetchOpenIDConfig() error {
	metaUrl := j.MetadataAddress
	if metaUrl == "" && j.Authority != "" {
		metaUrl = strings.TrimRight(j.Authority, "/") + "/.well-known/openid-configuration"
	}
	if metaUrl == "" {
		return errors.New("no MetadataAddress or Authority configured")
	}

	u, err := url.Parse(metaUrl)
	if err != nil {
		return err
	}
	if j.RequireHttpsMetadata && u.Scheme != "https" {
		return errors.New("RequireHttpsMetadata is true but metadata address is not HTTPS")
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, metaUrl, nil)
	if err != nil {
		return err
	}
	resp, err := j.BackchannelHttpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return errors.New("failed to get metadata")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var config OpenIDConfig
	if err := json.Unmarshal(body, &config); err != nil {
		return err
	}
	config.Expires = time.Now().Add(j.AutomaticRefreshInterval)

	j.configMu.Lock()
	j.openIDConfig = &config
	j.configMu.Unlock()

	return nil
}

// FeatchJWKS fetches the JSON Web Key Set (JWKS) for an issuer.
func (j *JwtBearerOptions) FetchJWKS() error {
	j.configMu.RLock()
	jwksUri := ""
	if j.openIDConfig != nil {
		jwksUri = j.openIDConfig.JwksURI
	}
	j.configMu.RUnlock()
	if jwksUri == "" {
		return errors.New("jwks_uri is empty")
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, jwksUri, nil)
	if err != nil {
		return err
	}
	resp, err := j.BackchannelHttpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return errors.New("failed to get jwks")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var jwks JSONWebKeySet
	if err := json.Unmarshal(body, &jwks); err != nil {
		return err
	}

	tmpCache := make(map[string]interface{})
	for _, key := range jwks.Keys {
		// 这里只演示RSA, 你可以用你喜欢的方式解析RSA公钥
		if key.Kty == "RSA" && key.Use == "sig" {
			pubKey, err := parseRSAPublicKey(key.N, key.E)
			if err == nil {
				tmpCache[key.Kid] = pubKey
			}
		}
	}

	j.jwksMu.Lock()
	j.jwksCache = tmpCache
	j.TokenValidationParameters.signingKeys = tmpCache
	j.jwksMu.Unlock()
	return nil
}

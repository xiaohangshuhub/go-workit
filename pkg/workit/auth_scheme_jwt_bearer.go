package workit

import (
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

const JWTBearerSchemeName = "JWTBearer"

type JWTBearerScheme struct {
	Secret []byte
}

func (j *JWTBearerScheme) Scheme() string {
	return JWTBearerSchemeName
}

func (j *JWTBearerScheme) Authenticate(r *http.Request) (*ClaimsPrincipal, error) {

	token := r.Header.Get("Authorization")
	token = strings.TrimPrefix(token, "Bearer ")

	parsed, err := jwt.ParseWithClaims(token, &jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return j.Secret, nil
	})
	if err != nil || !parsed.Valid {
		return nil, errors.New("invalid token")
	}

	claims := parsed.Claims.(*jwt.MapClaims)
	name := (*claims)["name"].(string)
	roles := []string{}
	if r, ok := (*claims)["roles"].([]interface{}); ok {
		for _, role := range r {
			roles = append(roles, role.(string))
		}
	}

	return &ClaimsPrincipal{Name: name, Roles: roles, Claims: *claims}, nil
}

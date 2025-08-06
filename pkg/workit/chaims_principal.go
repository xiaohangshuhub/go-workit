package workit

// ClaimsPrincipal 表示用户身份和声明
type ClaimsPrincipal struct {
	Name   string
	Roles  []string
	Claims map[string]interface{}
}

func (cp *ClaimsPrincipal) AddRole(role string) {
	for _, r := range cp.Roles {
		if r == role {
			return
		}
	}
	cp.Roles = append(cp.Roles, role)
}

func (cp *ClaimsPrincipal) IsInRole(role string) bool {
	for _, r := range cp.Roles {
		if r == role {
			return true
		}
	}
	return false
}

func (cp *ClaimsPrincipal) AddClaim(key string, value interface{}) {
	if cp.Claims == nil {
		cp.Claims = make(map[string]interface{})
	}
	cp.Claims[key] = value
}

func (cp *ClaimsPrincipal) FindFirst(key string) (interface{}, bool) {
	if cp.Claims == nil {
		return nil, false
	}
	val, ok := cp.Claims[key]
	return val, ok
}

func (cp *ClaimsPrincipal) HasClaim(key string, value interface{}) bool {
	if cp.Claims == nil {
		return false
	}
	if v, ok := cp.Claims[key]; ok {
		return v == value
	}
	return false
}

func (cp *ClaimsPrincipal) Clone() *ClaimsPrincipal {
	newClaims := make(map[string]interface{}, len(cp.Claims))
	for k, v := range cp.Claims {
		newClaims[k] = v
	}
	newRoles := make([]string, len(cp.Roles))
	copy(newRoles, cp.Roles)
	return &ClaimsPrincipal{
		Name:   cp.Name,
		Roles:  newRoles,
		Claims: newClaims,
	}
}

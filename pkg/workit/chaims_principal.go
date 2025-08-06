package workit

// 用户身份
type ClaimsPrincipal struct {
	Name   string                 // 用户名
	Roles  []string               // 角色列表
	Claims map[string]interface{} // 其他声明
}

// 添加角色
func (cp *ClaimsPrincipal) AddRole(role string) {
	for _, r := range cp.Roles {
		if r == role {
			return
		}
	}
	cp.Roles = append(cp.Roles, role)
}

// 检查是否在角色中
func (cp *ClaimsPrincipal) IsInRole(role string) bool {
	for _, r := range cp.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// 添加声明
func (cp *ClaimsPrincipal) AddClaim(key string, value interface{}) {
	if cp.Claims == nil {
		cp.Claims = make(map[string]interface{})
	}
	cp.Claims[key] = value
}

// 获取声明
func (cp *ClaimsPrincipal) FindFirst(key string) (interface{}, bool) {
	if cp.Claims == nil {
		return nil, false
	}
	val, ok := cp.Claims[key]
	return val, ok
}

// 检查是否有指定声明
func (cp *ClaimsPrincipal) HasClaim(key string, value interface{}) bool {
	if cp.Claims == nil {
		return false
	}
	if v, ok := cp.Claims[key]; ok {
		return v == value
	}
	return false
}

// 克隆用户身份
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

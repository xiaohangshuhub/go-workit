package webapp

import "time"

const (
	Actor                      = "http://schemas.xmlsoap.org/ws/2009/09/identity/claims/actor"
	PostalCode                 = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/postalcode"
	PrimaryGroupSid            = "http://schemas.microsoft.com/ws/2008/06/identity/claims/primarygroupsid"
	PrimarySid                 = "http://schemas.microsoft.com/ws/2008/06/identity/claims/primarysid"
	Role                       = "http://schemas.microsoft.com/ws/2008/06/identity/claims/role"
	Rsa                        = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/rsa"
	SerialNumber               = "http://schemas.microsoft.com/ws/2008/06/identity/claims/serialnumber"
	Sid                        = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/sid"
	Spn                        = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/spn"
	StateOrProvince            = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/stateorprovince"
	StreetAddress              = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/streetaddress"
	Surname                    = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname"
	System                     = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/system"
	Thumbprint                 = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/thumbprint"
	Upn                        = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/upn"
	Uri                        = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/uri"
	UserData                   = "http://schemas.microsoft.com/ws/2008/06/identity/claims/userdata"
	Version                    = "http://schemas.microsoft.com/ws/2008/06/identity/claims/version"
	Webpage                    = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/webpage"
	WindowsAccountName         = "http://schemas.microsoft.com/ws/2008/06/identity/claims/windowsaccountname"
	WindowsDeviceClaim         = "http://schemas.microsoft.com/ws/2008/06/identity/claims/windowsdeviceclaim"
	WindowsDeviceGroup         = "http://schemas.microsoft.com/ws/2008/06/identity/claims/windowsdevicegroup"
	WindowsFqbnVersion         = "http://schemas.microsoft.com/ws/2008/06/identity/claims/windowsfqbnversion"
	WindowsSubAuthority        = "http://schemas.microsoft.com/ws/2008/06/identity/claims/windowssubauthority"
	OtherPhone                 = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/otherphone"
	NameIdentifier             = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/nameidentifier"
	Name                       = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name"
	MobilePhone                = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/mobilephone"
	Anonymous                  = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/anonymous"
	Authentication             = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/authenticated"
	AuthenticationInstant      = "http://schemas.microsoft.com/ws/2008/06/identity/claims/authenticationinstant"
	AuthenticationMethod       = "http://schemas.microsoft.com/ws/2008/06/identity/claims/authenticationmethod"
	AuthorizationDecision      = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/authorizationdecision"
	CookiePath                 = "http://schemas.microsoft.com/ws/2008/06/identity/claims/cookiepath"
	Country                    = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/country"
	DateOfBirth                = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/dateofbirth"
	DenyOnlyPrimaryGroupSid    = "http://schemas.microsoft.com/ws/2008/06/identity/claims/denyonlyprimarygroupsid"
	DenyOnlyPrimarySid         = "http://schemas.microsoft.com/ws/2008/06/identity/claims/denyonlyprimarysid"
	DenyOnlySid                = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/denyonlysid"
	WindowsUserClaim           = "http://schemas.microsoft.com/ws/2008/06/identity/claims/windowsuserclaim"
	DenyOnlyWindowsDeviceGroup = "http://schemas.microsoft.com/ws/2008/06/identity/claims/denyonlywindowsdevicegroup"
	Dsa                        = "http://schemas.microsoft.com/ws/2008/06/identity/claims/dsa"
	Email                      = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress"
	Expiration                 = "http://schemas.microsoft.com/ws/2008/06/identity/claims/expiration"
	Expired                    = "http://schemas.microsoft.com/ws/2008/06/identity/claims/expired"
	Gender                     = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/gender"
	GivenName                  = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname"
	GroupSid                   = "http://schemas.microsoft.com/ws/2008/06/identity/claims/groupsid"
	Hash                       = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/hash"
	HomePhone                  = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/homephone"
	IsPersistent               = "http://schemas.microsoft.com/ws/2008/06/identity/claims/ispersistent"
	Locality                   = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/locality"
	Dns                        = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/dns"
	X500DistinguishedName      = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/x500distinguishedname"
)

// Claim 表示一个 Claim
type Claim struct {
	Type  string
	Value any
}

// ClaimsPrincipal 表示用户身份和声明
type ClaimsPrincipal struct {
	Subject              string    // 一般对应 nameidentifier 或 sub（JWT）
	Name                 string    // Display name
	Roles                []string  // 权限角色列表
	Claims               []Claim   // 所有的 Claim（通用扩展）
	IdentityProvider     string    // 身份提供者（如 AzureAD、Google、Internal 等）
	AuthenticationMethod string    // 认证方式（如 password、oauth、saml）
	AuthenticatedAt      time.Time // 认证时间
}

// AddRole 添加角色
func (cp *ClaimsPrincipal) AddRole(role string) {
	for _, r := range cp.Roles {
		if r == role {
			return
		}
	}
	cp.Roles = append(cp.Roles, role)
}

// IsInRole 判断是否有指定角色, 返回 true 表示有指定角色
func (cp *ClaimsPrincipal) IsInRole(role string) bool {
	for _, r := range cp.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// AddClaim 添加 Claim
func (cp *ClaimsPrincipal) AddClaim(key string, value interface{}) {
	cp.Claims = append(cp.Claims, Claim{Type: key, Value: value})
}

// FindFirst 查找第一个 Claim, 返回 Claim 的值和是否存在
func (cp *ClaimsPrincipal) FindFirst(key string) (interface{}, bool) {
	for _, c := range cp.Claims {
		if c.Type == key {
			return c.Value, true
		}
	}
	return nil, false
}

// HasClaim 判断是否有指定 Claim, 返回 true 表示有指定 Claim
func (cp *ClaimsPrincipal) HasClaim(key string, value interface{}) bool {
	for _, c := range cp.Claims {
		if c.Type == key && c.Value == value {
			return true
		}
	}
	return false
}

// Clone 克隆 ClaimsPrincipal 对象
func (cp *ClaimsPrincipal) Clone() *ClaimsPrincipal {
	newClaims := make([]Claim, len(cp.Claims))
	copy(newClaims, cp.Claims)

	newRoles := make([]string, len(cp.Roles))
	copy(newRoles, cp.Roles)

	return &ClaimsPrincipal{
		Subject:              cp.Subject,
		Name:                 cp.Name,
		Roles:                newRoles,
		Claims:               newClaims,
		IdentityProvider:     cp.IdentityProvider,
		AuthenticationMethod: cp.AuthenticationMethod,
		AuthenticatedAt:      cp.AuthenticatedAt,
	}
}

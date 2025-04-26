package validate

import "regexp"

// IsEmail 验证是否为电子邮件地址
func IsEmail(email string) bool {
	pattern := `^[0-9a-z][_.0-9a-z-]{0,31}@([0-9a-z][0-9a-z-]{0,30}[0-9a-z]\.){1,4}[a-z]{2,4}$`
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(email)
}

// IsPhone 验证是否为手机号码
func IsPhone(phone string) bool {
	pattern := `^1[3-9]\d{9}$`
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(phone)
}

// IsIDCard 验证是否为身份证号
func IsIDCard(id string) bool {
	pattern := `(^\d{15}$)|(^\d{18}$)|(^\d{17}(\d|X|x)$)`
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(id)
}

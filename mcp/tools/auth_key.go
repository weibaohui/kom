package tools

var authKey string

// SetAuthKey 设置包级认证密钥为指定的字符串。
func SetAuthKey(key string) {
	authKey = key
}
// AuthKey 返回当前存储的认证密钥字符串。
func AuthKey() string {
	return authKey
}

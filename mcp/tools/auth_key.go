package tools

var authKey string

func SetAuthKey(key string) {
	authKey = key
}
func AuthKey() string {
	return authKey
}

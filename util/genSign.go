package util

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
)

//GenSign 產生 Sing
func GenSign(appKey string, data string) string {
	key := []byte(appKey)
	mac := hmac.New(sha1.New, key)
	mac.Write([]byte(data))
	sign := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return sign
}

//CompareSign 比對 sign 是否符合
func CompareSign(appKey string, signStr string, srcData string) bool {
	s := GenSign(appKey, srcData)
	return s == signStr
}

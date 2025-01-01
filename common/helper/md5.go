package helper

import (
	"crypto/md5"
	"encoding/hex"
)

func GetMD5FromString(content string) string {
	b := md5.Sum([]byte(content))
	return hex.EncodeToString(b[:])
}

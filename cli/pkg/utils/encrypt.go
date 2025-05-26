package utils

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
)

func MD5(str string) string {
	hasher := md5.New()

	hasher.Write([]byte(str))
	hashBytes := hasher.Sum(nil)
	return hex.EncodeToString(hashBytes)
}

func Base64decode(s string) (string, error) {
	r, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	return string(r), nil
}

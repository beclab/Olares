package utils

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"os"
	"strings"

	"k8s.io/klog/v2"
)

// PrettyJSON print pretty json.
func PrettyJSON(v any) string {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		klog.Error("cannot encode json", err)
	}
	return buf.String()
}

// ListContains returns true if a value is present in items slice, false otherwise.
func ListContains[T comparable](items []T, v T) bool {
	for _, item := range items {
		if v == item {
			return true
		}
	}
	return false
}

func GetEnvOrDefault(env string, defaultValue string) string {
	if value, exists := os.LookupEnv(env); exists {
		return value
	}
	return defaultValue
}

func IsServiceAccount(name string) (isSA bool, namespace string, saName string) {
	isSA = strings.HasPrefix(name, "system:serviceaccount:")
	if isSA {
		parts := strings.SplitN(name, ":", 4)
		if len(parts) == 4 {
			return true, parts[2], parts[3]
		}
	}
	return false, "", ""
}

func IsUserSystemNamespace(namespace string) (bool, string) {
	usersystemPrefix := "user-system-"
	return strings.HasPrefix(namespace, usersystemPrefix), strings.TrimPrefix(namespace, usersystemPrefix)
}

func IsUserSpaceNamespace(namespace string) (bool, string) {
	userspacerefix := "user-space-"
	return strings.HasPrefix(namespace, userspacerefix), strings.TrimPrefix(namespace, userspacerefix)
}

func IsUserNamespace(namespace string) (bool, string) {
	if isUserSystem, userName := IsUserSystemNamespace(namespace); isUserSystem {
		return true, userName
	}
	if isUserSpace, userName := IsUserSpaceNamespace(namespace); isUserSpace {
		return true, userName
	}
	return false, ""
}

func Md5String(s string) string {
	hash := md5.Sum([]byte(s))
	hashString := hex.EncodeToString(hash[:])
	return hashString
}

func AppId(s string) string {
	// AppId is the md5 hash of the app name
	if s == "" {
		return ""
	}
	return Md5String(s)[:8]
}

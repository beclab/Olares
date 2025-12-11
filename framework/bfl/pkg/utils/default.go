package utils

import "os"

func EnvOrDefault(name, def string) string {
	v := os.Getenv(name)

	if v == "" && def != "" {
		return def
	}
	return v
}

func StringOrDefault(val, def string) string {
	if val != "" {
		return val
	}

	if def != "" {
		return def
	}

	return ""
}

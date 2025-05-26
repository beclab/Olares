package utils

import (
	"regexp"
)

func CheckUrl(val string) bool {
	regex := `^(https://)([a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}$`
	re := regexp.MustCompile(regex)
	if ok := re.MatchString(val); !ok {
		return false
	}
	return true
}

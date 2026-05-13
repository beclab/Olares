package users

import (
	"crypto/rand"
	"fmt"
	"io"
	"regexp"
)

const passwordCharsetSPA = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

const spaPasswordALLRuleRaw = `^(.*[a-z].*[A-Z].*[0-9].*)$|^(.*[a-z].*[0-9].*[A-Z].*)$|^(.*[A-Z].*[a-z].*[0-9].*)$|^(.*[A-Z].*[0-9].*[a-z].*)$|^(.*[0-9].*[a-z].*[A-Z].*)$|^(.*[0-9].*[A-Z].*[a-z].*)$|^(\$2[ayb]\$.{56})$`

var spaPasswordAllRuleRegexp = regexp.MustCompile(spaPasswordALLRuleRaw)

func generatePasswordSPA() (string, error) {
	return generatePasswordSPAReader(rand.Reader)
}

func generatePasswordSPAReader(r io.Reader) (string, error) {
	if r == nil {
		return "", fmt.Errorf("nil reader")
	}
	const n = 16
	for attempt := 0; attempt < 10000; attempt++ {
		b := make([]byte, n)
		if _, err := io.ReadFull(r, b); err != nil {
			return "", err
		}
		sb := make([]byte, n)
		for i := range b {
			sb[i] = passwordCharsetSPA[int(b[i])%len(passwordCharsetSPA)]
		}
		s := string(sb)
		if spaPasswordAllRuleRegexp.MatchString(s) {
			return s, nil
		}
	}
	return "", fmt.Errorf("give up generating password matching SPA ALL_RULE")
}

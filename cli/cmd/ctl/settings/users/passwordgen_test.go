package users

import (
	"crypto/rand"
	"strings"
	"testing"
)

func TestGeneratePasswordSPA(t *testing.T) {
	t.Parallel()
	const nSample = 200
	for i := 0; i < nSample; i++ {
		p, err := generatePasswordSPAReader(rand.Reader)
		if err != nil {
			t.Fatalf("iteration %d: %v", i, err)
		}
		if len(p) != 16 {
			t.Fatalf("len want 16 got %d %q", len(p), p)
		}
		for _, r := range p {
			if !strings.ContainsRune(passwordCharsetSPA, r) {
				t.Fatalf("char %q not in SPA charset %q", r, passwordCharsetSPA)
			}
		}
		if !spaPasswordAllRuleRegexp.MatchString(p) {
			t.Fatalf("does not match ALL_RULE: %q", p)
		}
	}
}

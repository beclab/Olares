package crypto

import (
	"testing"
)

func TestBox(t *testing.T) {
	encryptedKey := "wVFyiOF4vwKlSD2DJBuuiZ0HiZK9x+8Mgyg2mYvpEvIprx9kKgT3kP75JD60jjLW"
	nonce := "u71mw28ZK8LsgvTEC5UMcNREjXfGtDMh"
	publicKey := "cf44BhkybbBfsE0fZHe2jvqtCj6KLXvSq4hVjV0svzk="
	privateKey := "0WP0C2tB59Sij6VGO/VVsTuIfylBOR7QpcdYEAtFhPU="
	// encryptedKey := "HRAmxqb8rSRfmz+p6SX5JpMEVvirsJZoD5UdQJPC/4s="
	// nonce := "PmIqotVP1pQJyNTMxMhuUzMGXHoqhP2e"
	// publicKey := "cf44BhkybbBfsE0fZHe2jvqtCj6KLXvSq4hVjV0svzk="

	plainText, err := DecryptAsymmetric(encryptedKey, nonce, publicKey, privateKey)
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	t.Log(plainText)
}

package infisical

import (
	"errors"
	"strings"

	infisical_crypto "bytetrade.io/web3os/tapr/pkg/vault/infisical/crypto"
	"k8s.io/klog/v2"
)

func DecryptUserPrivateKeyHelper(user *UserEncryptionKeysPG, password string) (string, error) {
	switch user.EncryptionVersion {
	case 1:

		privateKey, err := infisical_crypto.Decrypt(
			user.EncryptedPrivateKey,
			user.IV,
			user.Tag,
			secretOfPassword(password),
		)

		if err != nil {
			klog.Error("decrypt user private key error, ", err)
			return "", err
		}

		return privateKey, nil

	default:
		return "", errors.New("unimplement")
	}
}
func secretOfPassword(password string) string {
	if len(password) >= 32 {
		return password[:32]
	}

	return strings.Repeat("0", 32-len(password)) + password
}

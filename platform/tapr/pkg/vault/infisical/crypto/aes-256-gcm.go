package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
)

const BLOCK_SIZE_BYTES = 16

func Encrypt(text, secret string) (ciphertext, iv, tag string, err error) {
	plainText := []byte(text)
	nonce := make([]byte, BLOCK_SIZE_BYTES)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", "", "", err
	}

	block, err := aes.NewCipher([]byte(secret))
	if err != nil {
		return "", "", "", err
	}

	aesgcm, err := cipher.NewGCMWithNonceSize(block, BLOCK_SIZE_BYTES)
	if err != nil {
		return "", "", "", err
	}

	ret := aesgcm.Seal(nil, nonce, plainText, nil)
	ciphertext = base64.StdEncoding.EncodeToString(ret[:len(plainText)])
	iv = base64.StdEncoding.EncodeToString(nonce)
	tag = base64.StdEncoding.EncodeToString(ret[len(plainText):])

	return
}

// Shared secret key, must be 32 bytes.
func Decrypt(ciphertext, iv, tag, secret string) (string, error) {

	nonce, err := base64.StdEncoding.DecodeString(iv)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher([]byte(secret))
	if err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCMWithNonceSize(block, BLOCK_SIZE_BYTES)
	if err != nil {
		return "", err
	}

	cipherIn, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	part2, err := base64.StdEncoding.DecodeString(tag)
	if err != nil {
		return "", err
	}

	cipherIn = append(cipherIn, part2...)

	plaintext, err := aesgcm.Open(nil, nonce, cipherIn, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

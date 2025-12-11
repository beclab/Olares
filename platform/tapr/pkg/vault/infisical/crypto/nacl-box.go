package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"

	"golang.org/x/crypto/nacl/box"
	"k8s.io/klog/v2"
)

const NACL_NONCE_SIZE_BYTES = 24

/*
plaintext hex string,
publicKey, privateKey base64 string
return: ciphertext, nonce base64 string
*/
func EncryptAssymmetric(plaintext, publicKey, privateKey string) (ciphertext, nonce string, err error) {
	var nonceBytes [NACL_NONCE_SIZE_BYTES]byte
	if _, err := io.ReadFull(rand.Reader, nonceBytes[:]); err != nil {
		return "", "", err
	}

	publicKeyData, err := base64.StdEncoding.DecodeString(publicKey)
	if err != nil {
		klog.Error("decode publicKey error, ", err)
		return "", "", err
	}

	privateKeyData, err := base64.StdEncoding.DecodeString(privateKey)
	if err != nil {
		klog.Error("decode privateKey error, ", err)
		return "", "", err
	}

	var (
		pk, sk [32]byte
	)
	copy(pk[:], publicKeyData)
	copy(sk[:], privateKeyData)

	encrypted := box.Seal(nil,
		[]byte(plaintext),
		&nonceBytes,
		&pk,
		&sk,
	)

	return base64.StdEncoding.EncodeToString(encrypted), base64.StdEncoding.EncodeToString(nonceBytes[:]), nil
}

/*
ciphertext, publicKey, privateKey base64 string
return: plaintext hex string
*/
func DecryptAsymmetric(ciphertext, nonce, publicKey, privateKey string) (plainText string, err error) {
	cipherData, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		klog.Error("decode ciphertext error, ", err)
		return "", err
	}

	var (
		n      [24]byte
		pk, sk [32]byte
	)

	nonceData, err := base64.StdEncoding.DecodeString(nonce)
	if err != nil {
		klog.Error("decode nonce error, ", err)
		return "", err
	}

	pkData, err := base64.StdEncoding.DecodeString(publicKey)
	if err != nil {
		klog.Error("decode publicKey error, ", err)
		return "", err
	}

	skData, err := base64.StdEncoding.DecodeString(privateKey)
	if err != nil {
		klog.Error("decode privateKey error, ", err)
		return "", err
	}

	copy(n[:], nonceData)
	copy(pk[:], pkData)
	copy(sk[:], skData)

	plainData, ok := box.Open(nil, cipherData, &n, &pk, &sk)
	if !ok {
		return "", errors.New("decrypt key error")
	}

	return string(plainData), nil
}

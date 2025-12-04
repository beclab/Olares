package utils

import (
	"bytes"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"strconv"
	"time"
)

var ErrUnPadding = errors.New("UnPadding error")

func pKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func pKCS7UnPadding(origin []byte) ([]byte, error) {
	length := len(origin)
	if length == 0 {
		return origin, ErrUnPadding
	}
	unpadding := int(origin[length-1])
	if length < unpadding {
		return origin, ErrUnPadding
	}
	return origin[:(length - unpadding)], nil
}

// AesEncrypt encrypts the given plaintext using AES encryption with the specified key.
func AesEncrypt(origin, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	origin = pKCS7Padding(origin, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	crypted := make([]byte, len(origin))
	blockMode.CryptBlocks(crypted, origin)
	return crypted, nil
}

// AesDecrypt decrypts the given ciphertext using AES encryption with the specified key.
func AesDecrypt(crypted, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	origin := make([]byte, len(crypted))
	blockMode.CryptBlocks(origin, crypted)
	origin, err = pKCS7UnPadding(origin)
	if err != nil {
		return origin, err
	}
	return origin, nil
}

func getTimestamp() string {
	t := time.Now().Unix()
	return strconv.Itoa(int(t))
}

// CheckSSLCertificate checks the validity of an SSL certificate and private key for a given hostname.
func CheckSSLCertificate(cert, key []byte, hostname string) error {
	block, _ := pem.Decode(cert)
	if block == nil {
		return errors.New("certificate is invalid")
	}
	pub, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return errors.New("certificate is invalid")
	}
	// verify hostname
	err = pub.VerifyHostname(hostname)
	if err != nil {
		return err
	}

	// verify certificate whether valid or expired
	currentTime := time.Now()
	if currentTime.Before(pub.NotBefore) {
		return errors.New("certificate is not yet valid")
	}
	if currentTime.After(pub.NotAfter) {
		return errors.New("certificate has expired")
	}

	block, _ = pem.Decode(key)
	if block == nil {
		return errors.New("error decoding private key PEM block")
	}

	hash := sha256.Sum256([]byte("hello"))

	switch block.Type {
	case "PRIVATE KEY":
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return fmt.Errorf("error parsing pkcs#8 private key: %v", err)
		}
		signature, err := rsa.SignPKCS1v15(rand.Reader, key.(*rsa.PrivateKey), crypto.SHA256, hash[:])
		if err != nil {
			return errors.New("failed to sign message")
		}
		rsaPub, ok := pub.PublicKey.(*rsa.PublicKey)
		if !ok {
			return errors.New("not RSA public key")
		}
		err = rsa.VerifyPKCS1v15(rsaPub, crypto.SHA256, hash[:], signature)
		if err != nil {
			return errors.New("certificate and private key not match")
		}
	case "EC PRIVATE KEY":
		key, err := x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			return fmt.Errorf("error parsing ecdsa private key: %v", err)
		}

		r, s, err := ecdsa.Sign(rand.Reader, key, hash[:])
		if err != nil {
			return fmt.Errorf("ecdsa sign err: %v", err)
		}
		verified := ecdsa.Verify(pub.PublicKey.(*ecdsa.PublicKey), hash[:], r, s)
		if !verified {
			return errors.New("certificate and private key not match")
		}
	case "RSA PRIVATE KEY":
		key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return fmt.Errorf("error parsing rsa private key: %v", err)
		}
		err = key.Validate()
		if err != nil {
			return fmt.Errorf("rsa private key failed validation: %v", err)
		}
		signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, hash[:])
		if err != nil {
			return errors.New("failed to sign message")
		}
		rsaPub, ok := pub.PublicKey.(*rsa.PublicKey)
		if !ok {
			return errors.New("not RSA public key")
		}
		err = rsa.VerifyPKCS1v15(rsaPub, crypto.SHA256, hash[:], signature)
		if err != nil {
			return errors.New("certificate and private key not match")
		}
	default:
		return fmt.Errorf("unknown private key type: %s", block.Type)
	}

	return nil
}

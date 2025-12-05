package utils

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"os"
	"time"
	"unicode"

	"bytetrade.io/web3os/app-service/pkg/constants"
	"github.com/go-crypt/crypt/algorithm/pbkdf2"

	"k8s.io/klog/v2"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// FmtAppName returns application name.
func FmtAppName(name, namespace string) string {
	return fmt.Sprintf("%s-%s", namespace, name)
}

// UserspaceName returns user-space namespace for a user.
func UserspaceName(user string) string {
	return fmt.Sprintf(constants.OwnerNamespaceTempl, constants.OwnerNamespacePrefix, user)
}

// PrettyJSON print pretty json.
func PrettyJSON(v any) string {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		klog.Errorf("Failed to encode json err=%v", err)
	}
	return buf.String()
}

// RandString returns random string with 64 char length.
func RandString() string {
	rand.Seed(time.Now().UnixNano())

	password := make([]byte, 64)
	for i := range password {
		password[i] = charset[rand.Intn(len(charset))]
	}
	// nats env variable used in .conf does not support a string begin with 0-9 or $
	if unicode.IsDigit(rune(password[0])) || password[0] == '$' {
		password[0] = 'T'
	}
	return string(password)
}

// Md5String returns md5 for a string.
func Md5String(s string) string {
	hash := md5.Sum([]byte(s))
	hashString := hex.EncodeToString(hash[:])
	return hashString
}

// default config
// PBKDF2Password{
// 	Variant:    sha512,
// 	Iterations: 310000,
// 	SaltLength: 16,
// },

func Pbkdf2Crypto(password string) (hash string, err error) {
	hasher, err := pbkdf2.New(
		pbkdf2.WithVariantName("sha512"),
		pbkdf2.WithIterations(310000),
		pbkdf2.WithSaltLength(16),
	)

	if err != nil {
		err = fmt.Errorf("failed to initialize hash settings: %w", err)

		return
	}

	digest, err := hasher.Hash(password)
	if err != nil {
		return
	}

	hash = digest.Encode()
	return
}

func GetRandomCharacters() (r string) {
	var (
		n       int
		charset string
	)

	n, charset = DefaultN, CharSetRFC3986Unreserved

	rand := &Cryptographical{}

	return rand.StringCustom(n, charset)
}

func isUDPPortAvailable(port int) bool {
	return true
}

func isTCPPortAvailable(port int) bool {
	address := fmt.Sprintf("%s:%d", os.Getenv("HOSTIP"), port)

	_, err := net.Dial("tcp", address)
	if err == nil {
		return false
	}

	return true
}

func IsPortAvailable(protocol string, port int) bool {
	if protocol == "tcp" {
		return isTCPPortAvailable(port)
	}
	if protocol == "udp" {
		return isUDPPortAvailable(port)
	}
	return false
}

func ListToMap[K comparable, T any](list []T, keyFunc func(T) K) map[K]T {
	m := make(map[K]T)
	for _, item := range list {
		m[keyFunc(item)] = item
	}
	return m
}

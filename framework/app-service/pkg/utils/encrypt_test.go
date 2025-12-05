package utils

import (
	"testing"
)

func TestAesEncryptDecrypt(t *testing.T) {
	key := []byte("ZyCUs5znsmk7GTAiNXdE7RpkSRz1zsIm")

	plaintext := []byte("Hello, bytetrade!")

	ciphertext, err := AesEncrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encryption error: %v", err)
	}

	decrypted, err := AesDecrypt(ciphertext, key)
	if err != nil {
		t.Fatalf("Decryption error: %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Errorf("Decrypted data does not match original data")
	}
}

func TestMatchVersion(t *testing.T) {
	testCases := []struct {
		version    string
		constraint string
		expected   bool
	}{
		{"latest", ">=0.3.2-beta.2", true},
		{"1.0.0", "1.0.0", true},
		{"1.0.0", ">=1.0.0", true},
		{"1.0.0", ">1.0.0", false},
		{"1.0.0", "<1.0.0", false},
		{"1.0.0", "<=1.0.0", true},
		{"1.0.0", ">=1.0.0,<2.0.0", true},
		{"1.0.0", ">=1.0.0,<=2.0.0", true},
		{"1.0.0", ">1.0.0,<2.0.0", false},
		{"1.0.0", ">1.0.0,<=2.0.0", false},
		{"1.0.0", ">=1.0.0,<=2.0.0", true},
		{"0.3.3-beta.2", ">=0.3.2-beta.2", true},
		{"0.4.1-20220124", ">=0.3.0-0", true},
	}
	for _, testCase := range testCases {
		matched := MatchVersion(testCase.version, testCase.constraint)
		if matched != testCase.expected {
			t.Errorf("Version: %s,RangeVersion: %s, Expected: %v, Got: %v",
				testCase.version, testCase.constraint, testCase.expected, matched)
		}
	}
}

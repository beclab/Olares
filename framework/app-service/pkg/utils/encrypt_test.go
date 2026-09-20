package utils

import (
	"testing"
)

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

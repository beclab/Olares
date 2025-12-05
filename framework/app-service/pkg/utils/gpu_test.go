package utils

import "testing"

func TestExtractGPUVersion(t *testing.T) {
	testCases := []struct {
		name     string
		expected string
	}{
		{
			name:     "NVIDIA RTX A6000",
			expected: "A6000",
		},
		{
			name:     "QUADRO RTX 6000",
			expected: "6000",
		},
		{
			name:     "QUADRO T2000",
			expected: "T2000",
		},
		{
			name:     "NVIDIA T600",
			expected: "T600",
		},
		{
			name:     "GeForce RTX 3090 Ti",
			expected: "3090",
		},
		{
			name:     "GeForce RTX 3090",
			expected: "3090",
		},
		{
			name:     "NVIDIA A100",
			expected: "A100",
		},
		{
			name:     "NVIDIA RTX 2000 Ada",
			expected: "2000",
		},
		{
			name:     "NVIDIA GH200",
			expected: "GH200",
		},
		{
			name:     "NVIDIA H200",
			expected: "H200",
		},
		{
			name:     "NVIDIA GB200",
			expected: "GB200",
		},
		{
			name:     "NVIDIA B200",
			expected: "B200",
		},
		{
			name:     "NVIDIA RTX PRO 5000 Blackwell",
			expected: "5000",
		},
	}
	for _, testCase := range testCases {
		expected := ExtractGPUVersion(testCase.name)
		if expected != testCase.expected {
			t.Errorf("expeected: %s, got: %s", testCase.expected, expected)
		}
	}
}

package utils

import (
	"regexp"
)

func ExtractGPUVersion(gpuName string) string {
	pattern := regexp.MustCompile(`[A-Za-z]*\d+[A-Za-z]*`)
	return pattern.FindString(gpuName)
}

package images

import "strings"

func IsImage(ext string) bool {
	switch strings.ToLower(ext) {
	case "png", "jpeg", "jpg", "gif":
		return true
	}

	return false
}

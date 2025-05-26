package manifest

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"bytetrade.io/web3os/installer/pkg/core/logger"
)

func ReadItem(line string) (*ManifestItem, error) {
	token := strings.Split(line, ",")
	if len(token) < 8 {
		return nil, errors.New("invalid format")
	}

	item := &ManifestItem{
		Filename: token[0],
		Path:     token[1],
		Type:     token[2],
		URL: itemUrl{
			AMD64: fileUrl{
				Url:      token[3],
				Checksum: token[4],
			},
			ARM64: fileUrl{
				Url:      token[5],
				Checksum: token[6],
			},
		},
	}
	if strings.HasPrefix(token[2], "images.") && len(token) > 7 {
		item.ImageName = token[7]
	}
	item.FileID = token[7]

	return item, nil
}

func ReadAll(path string) (InstallationManifest, error) {
	m := New()
	if data, err := os.ReadFile(path); err != nil {
		logger.Error("unable to read manifest, ", err)
		return nil, err
	} else {
		scanner := bufio.NewScanner(bytes.NewReader(data))

		for scanner.Scan() {
			line := scanner.Text()
			if line == "" && strings.HasPrefix(line, "#") {
				continue
			}

			item, err := ReadItem(line)
			if err != nil {
				return nil, err
			}

			m[item.FileID] = item
		}
	}

	return m, nil
}

func New() InstallationManifest {
	return make(InstallationManifest)
}

func (item *ManifestItem) GetItemUrlForHost(osArch string) *fileUrl {
	switch osArch {
	case "arm64":
		return &item.URL.ARM64
	}

	return &item.URL.AMD64
}

func (item *ManifestItem) FilePath(basePath string) string {
	return filepath.Join(basePath, item.Path, item.Filename)
}

func (i InstallationManifest) Get(fileID string) (*ManifestItem, error) {
	item, ok := i[fileID]
	if !ok {
		return nil, fmt.Errorf("manifest item not found, %s", fileID)
	}

	return item, nil
}

func (i InstallationManifest) GetImageList() (list []*ManifestItem, keys []string) {
	for _, item := range i {
		if item.ImageName != "" {
			list = append(list, item)
			keys = append(keys, item.FileID)
		}
	}

	return
}

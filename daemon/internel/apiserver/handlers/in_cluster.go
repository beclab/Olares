package handlers

import (
	"path/filepath"
	"strings"

	"github.com/beclab/Olares/daemon/pkg/commands"
)

func clusterPathToNodePath(path string) string {
	return filepath.Join(commands.MOUNT_BASE_DIR, path)
}

func nodePathToClusterPath(path string) string {
	if strings.HasPrefix(path, commands.MOUNT_BASE_DIR) {
		return strings.TrimPrefix(path, commands.MOUNT_BASE_DIR+"/")
	}

	return path
}

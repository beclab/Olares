package apiserver

import (
	"path/filepath"
	"strings"

	"bytetrade.io/web3os/terminusd/pkg/commands"
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

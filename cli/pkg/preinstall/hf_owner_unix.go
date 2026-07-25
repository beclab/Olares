//go:build darwin || linux

package preinstall

import (
	"io/fs"
	"syscall"
)

func fileOwnerUID(info fs.FileInfo) (uint32, bool) {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, false
	}
	return stat.Uid, true
}

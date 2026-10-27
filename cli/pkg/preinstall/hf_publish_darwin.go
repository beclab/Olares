//go:build darwin

package preinstall

import (
	"os"

	"golang.org/x/sys/unix"
)

func platformRenameNoReplace(parent *os.File, oldName, newName string) error {
	fd := int(parent.Fd())
	return unix.RenameatxNp(fd, oldName, fd, newName, unix.RENAME_EXCL)
}

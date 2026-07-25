package preinstall

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func renameNoReplace(parent *os.File, oldName, newName string) error {
	for _, name := range []string{oldName, newName} {
		if name == "" || name == "." || name == ".." ||
			filepath.Clean(name) != name || filepath.Base(name) != name {
			return fmt.Errorf("publish path %q must be a clean single entry", name)
		}
	}
	if err := platformRenameNoReplace(parent, oldName, newName); err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("target already exists: %w", fs.ErrExist)
		}
		return err
	}
	return nil
}

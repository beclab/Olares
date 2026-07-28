package preinstall

import (
	"fmt"
	"io/fs"
	"os"
)

func renameNoReplace(parent *os.File, oldName, newName string) error {
	for _, name := range []string{oldName, newName} {
		if err := validateSingleEntry(name); err != nil {
			return fmt.Errorf("publish %w", err)
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

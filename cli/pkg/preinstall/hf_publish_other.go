//go:build !darwin && !linux

package preinstall

import (
	"fmt"
	"os"
)

func platformRenameNoReplace(_ *os.File, _, _ string) error {
	return fmt.Errorf("atomic no-replace directory publish is unsupported on this platform")
}

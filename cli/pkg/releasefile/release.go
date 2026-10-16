package releasefile

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/beclab/Olares/cli/internal/lockfile"
	"github.com/joho/godotenv"
)

const Path = "/etc/olares/release"

// Update updates the supplied entries and retains all other values.
func Update(ctx context.Context, entries map[string]string) error {
	filename := Path
	if len(entries) == 0 {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return err
	}
	unlock, err := lockfile.Acquire(ctx, filename+".lock")
	if err != nil {
		return err
	}
	defer unlock()

	values, err := godotenv.Read(filename)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read %s: %w", filename, err)
	}
	if values == nil {
		values = make(map[string]string)
	}
	for key, value := range entries {
		values[key] = value
	}
	content, err := godotenv.Marshal(values)
	if err != nil {
		return err
	}

	mode := os.FileMode(0644)
	if info, err := os.Stat(filename); err == nil {
		mode = info.Mode().Perm()
	} else if !os.IsNotExist(err) {
		return err
	}
	// Readers must see either the old file or the complete updated file.
	tmp, err := os.CreateTemp(filepath.Dir(filename), ".release-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()
	if err := tmp.Chmod(mode); err != nil {
		return err
	}
	if _, err := tmp.WriteString(content + "\n"); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), filename)
}

package terminus

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func validateUserChartSources(configDir string) error {
	checkChart := func(dir string) error {
		info, err := os.Stat(filepath.Join(dir, "Chart.yaml"))
		if err != nil {
			return fmt.Errorf("validate user chart %s: %w", dir, err)
		}
		if !info.Mode().IsRegular() || info.Size() == 0 {
			return fmt.Errorf("invalid Chart.yaml in %s", dir)
		}
		return nil
	}
	if err := checkChart(filepath.Join(configDir, "launcher")); err != nil {
		return err
	}
	appsDir := filepath.Join(configDir, "apps")
	entries, err := os.ReadDir(appsDir)
	if err != nil {
		return fmt.Errorf("read user chart bundle: %w", err)
	}
	chartCount := 0
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		if err := checkChart(filepath.Join(appsDir, entry.Name())); err != nil {
			return err
		}
		chartCount++
	}
	if chartCount == 0 {
		return fmt.Errorf("no user application charts in %s", appsDir)
	}
	return nil
}

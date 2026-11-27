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
	for _, name := range []string{"systemserver", "wizard"} {
		if err := checkChart(filepath.Join(appsDir, name)); err != nil {
			return err
		}
	}
	return nil
}

// Stage both trees on the PVC and publish a single relative symlink by rename.
// Keep previous generations so an in-flight user creation can finish with
// the chart paths it resolved before the publication.
func publishUserChartBundle(configDir, pod string, run func(...string) (string, error)) error {
	exec := func(args ...string) (string, error) {
		return run(append([]string{"exec", pod, "-c", "app-service", "--"}, args...)...)
	}
	stage, err := exec("mktemp", "-d", "/userapps/.bundle.XXXXXXXX")
	if err != nil {
		return fmt.Errorf("stage user chart bundle: %w", err)
	}
	stage = strings.TrimSpace(stage)
	name := filepath.Base(stage)
	if stage != "/userapps/"+name || !strings.HasPrefix(name, ".bundle.") ||
		len(name) != len(".bundle.XXXXXXXX") || strings.ContainsAny(strings.TrimPrefix(name, ".bundle."), "/. \t\n") {
		return fmt.Errorf("invalid user chart staging path")
	}
	for _, tree := range []string{"launcher", "apps"} {
		// Copy into the staging parent: kubectl cp may append the basename
		// when the destination already exists.
		if _, err := run("cp", filepath.Join(configDir, tree), pod+":"+stage, "-c", "app-service"); err != nil {
			_, _ = exec("rm", "-rf", stage)
			return fmt.Errorf("copy user chart tree %s: %w", tree, err)
		}
	}
	const publish = `set -eu
test -s "$1/launcher/Chart.yaml"
test -s "$1/apps/systemserver/Chart.yaml"
test -s "$1/apps/wizard/Chart.yaml"
ln -s "$2" "$1/current"
mv -Tf "$1/current" /userapps/current`
	if _, err := exec("/busybox/sh", "-c", publish, "publish-user-charts", stage, name); err != nil {
		return fmt.Errorf("publish user chart bundle: %w", err)
	}
	return nil
}

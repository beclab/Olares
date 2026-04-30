package oac

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// appDataRef is the template reference the app-data check is looking for.
var appDataRef = regexp.MustCompile(`\.Values\.userspace\.appdata`)

// checkAppDataUsage implements the built-in "if the chart references
// .Values.userspace.appdata its manifest must set permission.appData" rule.
// It is wired up via WithAppDataValidator().
func checkAppDataUsage(oacPath string, m Manifest) error {
	if m.PermissionAppData() {
		return nil
	}
	if !strings.HasSuffix(oacPath, string(filepath.Separator)) {
		oacPath += string(filepath.Separator)
	}
	templates := filepath.Join(oacPath, "templates")
	var firstHit string
	err := filepath.Walk(templates, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil || info.IsDir() {
			return err
		}
		if !strings.HasSuffix(path, ".yaml") {
			return nil
		}
		f, e := os.Open(path)
		if e != nil {
			return e
		}
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			if appDataRef.MatchString(scanner.Text()) {
				if firstHit == "" {
					firstHit = filepath.Base(path)
				}
				return nil
			}
		}
		// Fail closed: a mid-file scanner error (read failure, or a token
		// longer than bufio.MaxScanTokenSize) means we cannot rule out an
		// .Values.userspace.appdata reference past the error point. Surface
		// the error rather than reporting "no reference found".
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("scan %s: %w", path, err)
		}
		return nil
	})
	if err != nil {
		return err
	}
	if firstHit != "" {
		return fmt.Errorf(
			"found .Values.userspace.appdata in %s, but permission.appData is not set in OlaresManifest.yaml",
			firstHit,
		)
	}
	return nil
}

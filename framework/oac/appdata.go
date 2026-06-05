package oac

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/beclab/Olares/framework/oac/internal/manifest"
)

var (
	appDataRef   = regexp.MustCompile(`\.Values\.userspace\.appdata`)
	appCommonRef = regexp.MustCompile(`\.Values\.userspace\.appCommon`)
	sharedLibRef = regexp.MustCompile(`\.Values\.sharedlib`)
)

// findFirstTemplateRef scans templates/*.yaml under oacPath and returns the
// basename of the first file whose content matches re, or "" when none match.
func findFirstTemplateRef(oacPath string, re *regexp.Regexp) (string, error) {
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
			if re.MatchString(scanner.Text()) {
				if firstHit == "" {
					firstHit = filepath.Base(path)
				}
				return nil
			}
		}
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("scan %s: %w", path, err)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return firstHit, nil
}

func checkDeniedTemplateRef(oacPath string, re *regexp.Regexp, allowed bool, msg string) error {
	if allowed {
		return nil
	}
	hit, err := findFirstTemplateRef(oacPath, re)
	if err != nil {
		return err
	}
	if hit != "" {
		return fmt.Errorf(msg, hit)
	}
	return nil
}

// checkAppDataUsage implements the built-in "if the chart references
// .Values.userspace.appdata its manifest must set permission.appData" rule.
func checkAppDataUsage(oacPath string, m Manifest) error {
	return checkDeniedTemplateRef(oacPath, appDataRef, m.PermissionAppData(),
		"found .Values.userspace.appdata in %s, but permission.appData is not set in OlaresManifest.yaml")
}

// checkAppCommonUsage rejects chart templates that reference
// .Values.userspace.appCommon when permission.appCommon is not true.
func checkAppCommonUsage(oacPath string, m Manifest) error {
	return checkDeniedTemplateRef(oacPath, appCommonRef, m.PermissionAppCommon(),
		"found .Values.userspace.appCommon in %s, but permission.appCommon is not true in OlaresManifest.yaml")
}

// checkSharedLibUsage rejects chart templates that reference .Values.sharedlib
// when permission.externalData is not true. The rule only applies to manifests
// with olaresManifest.version >= 0.12.0, matching when externalData permission
// is meaningful.
func checkSharedLibUsage(oacPath string, m Manifest) error {
	if !manifest.IsModernResourcesManifest(m.ConfigVersion()) {
		return nil
	}
	return checkDeniedTemplateRef(oacPath, sharedLibRef, m.PermissionExternalData(),
		"found .Values.sharedlib in %s, but permission.externalData is not true in OlaresManifest.yaml")
}

// checkPermissionTemplateUsage runs every manifest-vs-template permission
// cross-check (appdata, appCommon, sharedlib).
func checkPermissionTemplateUsage(oacPath string, m Manifest) error {
	return errors.Join(
		checkAppDataUsage(oacPath, m),
		checkAppCommonUsage(oacPath, m),
		checkSharedLibUsage(oacPath, m),
	)
}

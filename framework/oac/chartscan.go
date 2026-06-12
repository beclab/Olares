package oac

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// olaresUserChartRef matches user-level environment variable names embedded in
// chart files. v3 apps must reference app envs via .Values.olaresEnv.<envName>
// instead of inlining OLARES_USER_* names in templates or values.yaml.
var olaresUserChartRef = regexp.MustCompile(`\bOLARES_USER`)

func isChartYAMLFile(path string) bool {
	lower := strings.ToLower(path)
	return strings.HasSuffix(lower, ".yaml") || strings.HasSuffix(lower, ".yml")
}

func shouldScanChartFile(oacPath, path string) bool {
	rel, err := filepath.Rel(oacPath, path)
	if err != nil {
		return false
	}
	if filepath.Base(path) == "OlaresManifest.yaml" {
		return false
	}
	if strings.HasSuffix(rel, "OlaresManifest.yaml") {
		return false
	}
	base := filepath.Base(path)
	if base == "values.yaml" || base == "values.yml" {
		return true
	}
	if filepath.Base(filepath.Dir(path)) == "templates" && isChartYAMLFile(path) {
		return true
	}
	return false
}

// findFirstInChartFiles scans chart templates and values.yaml files under
// oacPath (including subcharts) and returns the relative path of the first
// file whose content matches re.
func findFirstInChartFiles(oacPath string, re *regexp.Regexp) (string, error) {
	if !strings.HasSuffix(oacPath, string(filepath.Separator)) {
		oacPath += string(filepath.Separator)
	}
	var firstHit string
	err := filepath.Walk(oacPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info == nil || info.IsDir() {
			return nil
		}
		if !shouldScanChartFile(oacPath, path) {
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
					rel, relErr := filepath.Rel(oacPath, path)
					if relErr != nil {
						firstHit = filepath.Base(path)
					} else {
						firstHit = rel
					}
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

// checkV3OLARESUserInChart rejects chart files that reference OLARES_USER*
// environment variable names directly. apiVersion=v3 apps must declare app
// env names under envs[] and consume them as .Values.olaresEnv.<envName>.
func checkV3OLARESUserInChart(oacPath string) error {
	hit, err := findFirstInChartFiles(oacPath, olaresUserChartRef)
	if err != nil {
		return err
	}
	if hit != "" {
		return fmt.Errorf(
			"found OLARES_USER in %s; apiVersion=v3 apps must not use OLARES_USER_* names in chart files — declare envs[].envName in OlaresManifest.yaml and reference .Values.olaresEnv.<envName> instead",
			hit,
		)
	}
	return nil
}

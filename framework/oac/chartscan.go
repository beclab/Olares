package oac

import (
	"bufio"
	"errors"
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

var (
	// workloadReplicaCountDotRef matches the canonical replica-count reference
	// {{ .Values.workloads.<name>.replicaCount }}.
	workloadReplicaCountDotRef = regexp.MustCompile(`\.Values\.workloads\.([A-Za-z0-9_]+)\.replicaCount`)
	// workloadReplicaCountIndexRef matches the index spelling
	// {{ (index .Values.workloads "<name>").replicaCount }}, which is required
	// when <name> contains characters (e.g. a hyphen) the dot accessor cannot
	// express.
	workloadReplicaCountIndexRef = regexp.MustCompile(`index\s+\.Values\.workloads\s+"([^"]+)"\s*\)\s*\.replicaCount`)

	// workloadKindLine matches a `kind: Deployment` / `kind: StatefulSet` line.
	workloadKindLine = regexp.MustCompile(`^\s*kind:\s*["']?(Deployment|StatefulSet)["']?\s*$`)
	// replicasFieldLine matches a `replicas: <value>` line and captures the value.
	replicasFieldLine = regexp.MustCompile(`^\s*replicas:\s*(\S.*?)\s*$`)
	// yamlDocSeparator splits a multi-document YAML file on `---` lines.
	yamlDocSeparator = regexp.MustCompile(`(?m)^---\s*$`)
)

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

// isUnderTemplatesDir reports whether path has a "templates" path segment, i.e.
// the file is a chart template. This intentionally matches nested layouts
// (templates/web/deployment.yaml) and subchart templates, not just files whose
// immediate parent is templates/.
func isUnderTemplatesDir(path string) bool {
	for _, seg := range strings.Split(filepath.ToSlash(path), "/") {
		if seg == "templates" {
			return true
		}
	}
	return false
}

// checkWorkloadReplicaTemplates verifies that every Deployment/StatefulSet
// template under oacPath that declares spec.replicas sources the value from
// .Values.workloads.<name>.replicaCount, where <name> is a key declared in the
// manifest's workloadReplicas map.
//
// It complements checkWorkloadReplicaValues: that function guarantees
// values.yaml carries each workloads.<name>.replicaCount default, while this
// one guarantees the templates actually consume it instead of hardcoding a
// replica count. Documents that omit replicas: are left to helm's default and
// to the rendered-name correspondence check.
func checkWorkloadReplicaTemplates(oacPath string, replicas map[string]int32) error {
	root := oacPath
	if !strings.HasSuffix(root, string(filepath.Separator)) {
		root += string(filepath.Separator)
	}
	var errs []error
	walkErr := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info == nil || info.IsDir() || !isUnderTemplatesDir(path) || !isChartYAMLFile(path) {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return fmt.Errorf("read %s: %w", path, readErr)
		}
		rel, relErr := filepath.Rel(oacPath, path)
		if relErr != nil {
			rel = filepath.Base(path)
		}
		errs = append(errs, scanWorkloadReplicaDoc(rel, string(data), replicas)...)
		return nil
	})
	if walkErr != nil {
		return walkErr
	}
	return errors.Join(errs...)
}

// scanWorkloadReplicaDoc splits content into YAML documents and returns one
// error per Deployment/StatefulSet document whose replicas: field is hardcoded
// (or otherwise not a .Values.workloads.<name>.replicaCount reference) or whose
// reference names a workload absent from replicas. rel is the chart-relative
// path used in error messages.
func scanWorkloadReplicaDoc(rel, content string, replicas map[string]int32) []error {
	var errs []error
	for _, doc := range yamlDocSeparator.Split(content, -1) {
		if !isWorkloadDoc(doc) {
			continue
		}
		value, ok := findReplicasValue(doc)
		if !ok {
			continue
		}
		name, matched := workloadReplicaRefName(value)
		if !matched {
			errs = append(errs, fmt.Errorf(
				"%s: Deployment/StatefulSet spec.replicas is set to %q; it must reference .Values.workloads.<name>.replicaCount so the replica count is sourced from values.yaml",
				rel, value,
			))
			continue
		}
		if _, ok := replicas[name]; !ok {
			errs = append(errs, fmt.Errorf(
				"%s: spec.replicas references .Values.workloads.%s.replicaCount, but %q is not declared in workloadReplicas",
				rel, name, name,
			))
		}
	}
	return errs
}

// isWorkloadDoc reports whether a YAML document declares a Deployment or
// StatefulSet kind.
func isWorkloadDoc(doc string) bool {
	for _, line := range strings.Split(doc, "\n") {
		if workloadKindLine.MatchString(line) {
			return true
		}
	}
	return false
}

// findReplicasValue returns the value of the first non-comment `replicas:` line
// in doc. Deployments and StatefulSets carry exactly one such field (spec.replicas).
func findReplicasValue(doc string) (string, bool) {
	for _, line := range strings.Split(doc, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		if m := replicasFieldLine.FindStringSubmatch(line); m != nil {
			return strings.TrimSpace(m[1]), true
		}
	}
	return "", false
}

// workloadReplicaRefName extracts the workload name from a replicas value that
// references .Values.workloads.<name>.replicaCount in either the dot or index
// spelling. It returns false when value is not such a reference.
func workloadReplicaRefName(value string) (string, bool) {
	if m := workloadReplicaCountDotRef.FindStringSubmatch(value); m != nil {
		return m[1], true
	}
	if m := workloadReplicaCountIndexRef.FindStringSubmatch(value); m != nil {
		return m[1], true
	}
	return "", false
}

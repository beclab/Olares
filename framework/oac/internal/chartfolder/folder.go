// Package chartfolder implements the structural checks that verify a chart
// directory is well-formed (Chart.yaml / values.yaml / templates/ /
// OlaresManifest.yaml present) and that its metadata is consistent with
// the parsed manifest.
package chartfolder

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"

	olm "github.com/beclab/Olares/framework/oac/internal/manifest"
)

// Error-message templates. They are exported here so the top-level oachecker
// package can surface them with identical wording to the legacy code base.
const (
	InvalidFolderName            = "invalid folder name: %q must match ^[a-z0-9]{1,30}$"
	FolderNotExist               = "folder does not exist: %q"
	MissingChartYaml             = "missing Chart.yaml in folder: %q"
	ReadChartYamlFailed          = "failed to read Chart.yaml in folder %q: %v"
	ParseChartYamlFailed         = "failed to parse Chart.yaml in folder %q: %v"
	ApiVersionFieldEmptyInAppCfg = "apiVersion field empty in Chart.yaml for chart %q"
	NameFieldEmptyInAppCfg       = "name field empty in Chart.yaml for chart %q"
	VersionFieldEmptyInAppCfg    = "version field empty in Chart.yaml for chart %q"
	MissingValuesYaml            = "missing values.yaml in folder: %q"
	MissingTemplatesFolder       = "missing templates folder in folder: %q"
	MissingAppCfg                = "missing OlaresManifest.yaml in folder: %q"
	NameMustSame1                = "inconsistent info. name must be the same in chart. name in Chart.yaml:%s, chartFolder:%s, OlaresManifest.yaml:%s"
	NameMustSame2                = "inconsistent info. name must be the same in chart. name in Chart.yaml:%s, chartFolder:%s, folder in title:%s, OlaresManifest.yaml:%s"
	VersionMustSame1             = "inconsistent info. Version must be the same in chart. version in OlaresManifest.yaml:%s, Chart.yaml:%s"
	VersionMustSame2             = "inconsistent info. Version must be the same in chart. version in OlaresManifest.yaml:%s, Chart.yaml:%s, title:%s"
	FolderNameInvalid            = "foldername %q is on the reserved foldername list"
)

// Chart is the minimum structure decoded from Chart.yaml.
type Chart struct {
	APIVersion string `yaml:"apiVersion"`
	Name       string `yaml:"name"`
	Version    string `yaml:"version"`
}

// TitleInfo mirrors the original TitleInfo shape; PR-driven callers pass the
// parsed title fields in so folder consistency can be validated against them.
type TitleInfo struct {
	PrType  string
	Folder  string
	Version string
}

// LoadChart reads and parses Chart.yaml from a chart folder.
func LoadChart(folder string) (*Chart, error) {
	chartFile := filepath.Join(folder, "Chart.yaml")
	if !fileExists(chartFile) {
		return nil, fmt.Errorf(MissingChartYaml, folder)
	}
	data, err := os.ReadFile(chartFile)
	if err != nil {
		return nil, fmt.Errorf(ReadChartYamlFailed, folder, err)
	}
	var c Chart
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf(ParseChartYamlFailed, folder, err)
	}
	if err := validateChartFields(c); err != nil {
		return nil, err
	}
	return &c, nil
}

// CheckLayout validates that folder is a proper chart directory. It does not
// touch manifest-level content.
func CheckLayout(folder string) error {
	folderName := path.Base(folder)
	if !isValidFolderName(folderName) {
		return fmt.Errorf(InvalidFolderName, folder)
	}
	if !dirExists(folder) {
		return fmt.Errorf(FolderNotExist, folder)
	}
	if !fileExists(filepath.Join(folder, "Chart.yaml")) {
		return fmt.Errorf(MissingChartYaml, folder)
	}
	if !fileExists(filepath.Join(folder, "values.yaml")) {
		return fmt.Errorf(MissingValuesYaml, folder)
	}
	if !dirExists(filepath.Join(folder, "templates")) {
		return fmt.Errorf(MissingTemplatesFolder, folder)
	}
	if !fileExists(filepath.Join(folder, "OlaresManifest.yaml")) {
		return fmt.Errorf(MissingAppCfg, folder)
	}
	return nil
}

// CheckConsistency runs CheckLayout then cross-validates the folder name and
// manifest metadata against Chart.yaml.
func CheckConsistency(folder string, chartFile *Chart, m olm.Manifest) error {
	if err := CheckLayout(folder); err != nil {
		return err
	}
	folderName := path.Base(folder)
	if chartFile.Name != folderName || m.AppName() != folderName {
		return fmt.Errorf(NameMustSame1, chartFile.Name, folderName, m.AppName())
	}
	if m.AppVersion() != chartFile.Version {
		return fmt.Errorf(VersionMustSame1, m.AppVersion(), chartFile.Version)
	}
	return nil
}

// CheckWithTitle cross-validates folder + chart + manifest against a PR title.
func CheckWithTitle(folder string, chartFile *Chart, m olm.Manifest, title TitleInfo, categories []string) error {
	if err := CheckLayout(folder); err != nil {
		return err
	}
	folderName := path.Base(folder)
	if chartFile.Name != folderName || title.Folder != folderName || m.AppName() != folderName {
		return fmt.Errorf(NameMustSame2, chartFile.Name, folderName, title.Folder, m.AppName())
	}
	if m.AppVersion() != chartFile.Version || title.Version != chartFile.Version {
		return fmt.Errorf(VersionMustSame2, m.AppVersion(), chartFile.Version, title.Version)
	}
	if !validCategories(categories) {
		return errors.New("metadata.categories contains unsupported category values")
	}
	if isReservedWord(folderName) {
		return fmt.Errorf(FolderNameInvalid, folderName)
	}
	return nil
}

func validateChartFields(c Chart) error {
	if c.APIVersion == "" {
		return fmt.Errorf(ApiVersionFieldEmptyInAppCfg, c.Name)
	}
	if c.Name == "" {
		return fmt.Errorf(NameFieldEmptyInAppCfg, c.Name)
	}
	if c.Version == "" {
		return fmt.Errorf(VersionFieldEmptyInAppCfg, c.Name)
	}
	return nil
}

func isReservedWord(name string) bool {
	reserved := []string{
		"user", "system", "space", "default", "os", "kubesphere", "kube",
		"kubekey", "kubernetes", "gpu", "tapr", "bfl", "bytetrade",
		"project", "pod",
	}
	for _, w := range reserved {
		if strings.EqualFold(name, w) {
			return true
		}
	}
	return false
}

func isValidFolderName(name string) bool {
	match, _ := regexp.MatchString("^[a-z0-9]{1,30}$", name)
	return match
}

// fileExists reports whether p points to a regular (non-directory) file.
// Any stat error - including "not exist" - is treated as "not a file".
// The previous implementation also honoured os.IsExist(err), which both
// rarely matches (os.Stat surfaces "not exist" errors, not "already
// exists" errors) and is unsafe: on a hypothetical non-nil err where
// os.IsExist(err) is true, info is nil and info.IsDir() panics.
func fileExists(p string) bool {
	info, err := os.Stat(p)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// dirExists reports whether p points to a directory. See fileExists for
// why a plain `err == nil` check is the correct gate.
func dirExists(p string) bool {
	info, err := os.Stat(p)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// validCategories is duplicated here (rather than injected) because the
// categories list is a small stable constant.
func validCategories(categories []string) bool {
	if len(categories) == 0 {
		return false
	}
	valid := map[string]bool{
		"AI":             true,
		"Blockchain":     true,
		"Utilities":      true,
		"Social Network": true,
		"Data":           true,
		"Entertainment":  true,
		"Productivity":   true,
		"Lifestyle":      true,
		"Developer":      true,
		"Multimedia":     true,
	}
	for _, c := range categories {
		if !valid[c] {
			return false
		}
	}
	return true
}

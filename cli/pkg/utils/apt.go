package utils

import (
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/core/util"
	"github.com/pkg/errors"
)

// AptSource represents an apt source entry
type AptSource struct {
	Type       string   // deb or deb-src
	Options    []string // options in square brackets (e.g., [arch=amd64,trusted=yes])
	URL        string   // repository URL
	Suite      string   // distribution suite (e.g., trixie, bullseye)
	Components []string // components (e.g., main, contrib, non-free)
}

// AddAptSource adds apt source or components to existing apt sources
// It reads /etc/apt/sources.list and any *.list files under /etc/apt/sources.list.d
// If a matching source exists, it merges the new components with existing ones
// If no matching source exists, it adds a new source line to /etc/apt/sources.list
func AddAptSource(sourceType, repoURL, suite string, components []string) error {
	if sourceType != "deb" && sourceType != "deb-src" {
		return errors.New("source type must be 'deb' or 'deb-src'")
	}

	if repoURL == "" {
		return errors.New("repository URL cannot be empty")
	}

	if _, err := url.ParseRequestURI(repoURL); err != nil {
		return errors.Wrap(err, "invalid repository URL")
	}

	if suite == "" {
		return errors.New("suite cannot be empty")
	}

	existingSources, err := readAptSources()
	if err != nil {
		return errors.Wrap(err, "failed to read existing apt sources")
	}

	var matchingSource *AptSource
	var matchingFile string
	var matchingLineIndex int

	for filePath, sources := range existingSources {
		for i, source := range sources {
			if source.Type == sourceType && source.Suite == suite && source.URL == repoURL {
				matchingSource = &source
				matchingFile = filePath
				matchingLineIndex = i
				break
			}
		}
		if matchingSource != nil {
			break
		}
	}

	if matchingSource != nil {
		mergedComponents := mergeComponents(matchingSource.Components, components)
		if len(mergedComponents) == len(matchingSource.Components) && len(components) > 0 {
			return nil
		}

		matchingSource.Components = mergedComponents
		existingSources[matchingFile][matchingLineIndex] = *matchingSource

		if err := writeAptSources(matchingFile, existingSources[matchingFile]); err != nil {
			return errors.Wrap(err, "failed to update apt sources file")
		}

		logger.Infof("Updated existing apt source with new components: %s %s %s %v\n",
			sourceType, repoURL, suite, mergedComponents)
	} else {
		newSource := AptSource{
			Type:       sourceType,
			URL:        repoURL,
			Suite:      suite,
			Components: components,
		}

		sourcesListPath := "/etc/apt/sources.list"
		if existingSources[sourcesListPath] == nil {
			existingSources[sourcesListPath] = []AptSource{}
		}
		existingSources[sourcesListPath] = append(existingSources[sourcesListPath], newSource)

		if err := writeAptSources(sourcesListPath, existingSources[sourcesListPath]); err != nil {
			return errors.Wrap(err, "failed to add new apt source")
		}

		logger.Infof("Added new apt source: %s %s %s %v\n", sourceType, repoURL, suite, components)
	}

	return nil
}

// readAptSources reads all apt source files and parses them
func readAptSources() (map[string][]AptSource, error) {
	sources := make(map[string][]AptSource)

	sourcesListPath := "/etc/apt/sources.list"
	if _, err := os.Stat(sourcesListPath); err == nil {
		parsedSources, err := parseAptSourceFile(sourcesListPath)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse /etc/apt/sources.list")
		}
		sources[sourcesListPath] = parsedSources
	}

	sourcesListDPath := "/etc/apt/sources.list.d"
	if _, err := os.Stat(sourcesListDPath); err == nil {
		err := filepath.Walk(sourcesListDPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Only process .list files
			if !info.IsDir() && strings.HasSuffix(path, ".list") {
				parsedSources, err := parseAptSourceFile(path)
				if err != nil {
					logger.Warnf("Warning: Failed to parse apt source file %s: %v\n", path, err)
					return nil
				}
				sources[path] = parsedSources
			}

			return nil
		})

		if err != nil {
			return nil, errors.Wrap(err, "failed to walk /etc/apt/sources.list.d")
		}
	}

	return sources, nil
}

// parseAptSourceFile parses a single apt source file
func parseAptSourceFile(filePath string) ([]AptSource, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read file")
	}

	var sources []AptSource
	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		source, err := parseAptSourceLine(line)
		if err != nil {
			logger.Debugf("Debug: Failed to parse apt source line '%s': %v\n", line, err)
			continue
		}

		sources = append(sources, source)
	}

	return sources, nil
}

// parseAptSourceLine parses a single apt source line
func parseAptSourceLine(line string) (AptSource, error) {
	// Remove inline comments (everything after #)
	if commentIndex := strings.Index(line, "#"); commentIndex != -1 {
		line = line[:commentIndex]
	}

	line = strings.TrimSpace(line)

	if line == "" {
		return AptSource{}, errors.New("empty line")
	}

	// Regex to match apt source lines with optional options and components
	// Format: deb [options] uri suite [component1] [component2] [...]
	re := regexp.MustCompile(`^(deb|deb-src)(?:\s+\[([^\]]+)\])?\s+(\S+)\s+(\S+)(?:\s+(.+))?$`)
	matches := re.FindStringSubmatch(line)

	// We expect 6 groups: [0] full match, [1] type, [2] options, [3] url, [4] suite, [5] components
	if len(matches) < 6 {
		return AptSource{}, errors.New("invalid apt source line format")
	}

	sourceType := matches[1]
	optionsStr := matches[2] // may be empty
	url := matches[3]
	suite := matches[4]
	componentsStr := strings.TrimSpace(matches[5]) // may be empty

	// Parse options if present
	var options []string
	if optionsStr != "" {
		// Split options by comma and trim whitespace
		optionParts := strings.Split(optionsStr, ",")
		for _, option := range optionParts {
			option = strings.TrimSpace(option)
			if option != "" {
				options = append(options, option)
			}
		}
	}

	// Split components by whitespace (if present)
	var components []string
	if componentsStr != "" {
		components = strings.Fields(componentsStr)
	}

	return AptSource{
		Type:       sourceType,
		Options:    options,
		URL:        url,
		Suite:      suite,
		Components: components,
	}, nil
}

// mergeComponents merges new components with existing ones, removing duplicates
func mergeComponents(existing, new []string) []string {
	componentMap := make(map[string]bool)
	var result []string

	// Add existing components
	for _, comp := range existing {
		if !componentMap[comp] {
			componentMap[comp] = true
			result = append(result, comp)
		}
	}

	// Add new components
	for _, comp := range new {
		if !componentMap[comp] {
			componentMap[comp] = true
			result = append(result, comp)
		}
	}

	return result
}

// formatAptSourceLine formats an AptSource into a proper apt sources.list line
func formatAptSourceLine(source AptSource) string {
	var parts []string

	// Add source type
	parts = append(parts, source.Type)

	// Add options in square brackets if present
	if len(source.Options) > 0 {
		optionsStr := "[" + strings.Join(source.Options, ",") + "]"
		parts = append(parts, optionsStr)
	}

	// Add URL and suite (required)
	parts = append(parts, source.URL)
	parts = append(parts, source.Suite)

	// Add components if present
	if len(source.Components) > 0 {
		parts = append(parts, strings.Join(source.Components, " "))
	}

	return strings.Join(parts, " ")
}

// writeAptSources writes sources back to a file
func writeAptSources(filePath string, sources []AptSource) error {
	backupPath := filePath + ".bak"
	if err := util.CopyFile(filePath, backupPath); err != nil {
		logger.Warnf("Warning: Failed to create backup of %s: %v\n", filePath, err)
	}

	// Build new file content
	var lines []string

	// Read original file to preserve comments and other content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return errors.Wrap(err, "failed to read original file")
	}

	originalLines := strings.Split(string(content), "\n")

	for _, line := range originalLines {
		trimmedLine := strings.TrimSpace(line)

		// Keep comments and empty lines as-is
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "#") {
			lines = append(lines, line)
			continue
		}

		// Parse the line to see if it's an apt source
		source, err := parseAptSourceLine(line)
		if err != nil {
			// Not an apt source line, keep as-is
			lines = append(lines, line)
			continue
		}

		// Check if this source is in our updated sources
		var found bool
		for _, updatedSource := range sources {
			if updatedSource.Type == source.Type &&
				updatedSource.URL == source.URL &&
				updatedSource.Suite == source.Suite {
				// Replace with updated source
				newLine := formatAptSourceLine(updatedSource)
				lines = append(lines, newLine)
				found = true
				break
			}
		}

		if !found {
			// Keep original line
			lines = append(lines, line)
		}
	}

	// Add any new sources that weren't in the original file
	for _, source := range sources {
		var exists bool
		for _, line := range originalLines {
			trimmedLine := strings.TrimSpace(line)
			if trimmedLine == "" || strings.HasPrefix(trimmedLine, "#") {
				continue
			}

			originalSource, err := parseAptSourceLine(line)
			if err != nil {
				continue
			}

			if originalSource.Type == source.Type &&
				originalSource.URL == source.URL &&
				originalSource.Suite == source.Suite {
				exists = true
				break
			}
		}

		if !exists {
			newLine := formatAptSourceLine(source)
			lines = append(lines, newLine)
		}
	}

	newContent := strings.Join(lines, "\n")

	tempFile := filePath + ".olares.tmp"
	if err := os.WriteFile(tempFile, []byte(newContent), 0644); err != nil {
		return errors.Wrap(err, "failed to write temporary file")
	}

	if err := os.Rename(tempFile, filePath); err != nil {
		return errors.Wrap(err, "failed to move temporary file to final location")
	}

	return nil
}

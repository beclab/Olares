package manifest

import (
	"bufio"
	"bytes"
	"regexp"
	"strings"
)

// Header carries the dispatch identifiers extracted from a manifest.
type Header struct {
	OlaresVersion string
	APIVersion    string
}

var (
	olaresVersionRE = regexp.MustCompile(`^olaresManifest\.version\s*:\s*(.*)$`)
	apiVersionRE    = regexp.MustCompile(`^apiVersion\s*:\s*(.*)$`)
)

// utf8BOM is the canonical UTF-8 byte-order mark sequence (U+FEFF encoded
// as three bytes). YAML files saved by Windows editors and pre-7
// PowerShell pipelines often start with one; the regexes below are
// ^-anchored, and Go's \s class does not include U+FEFF, so an unstripped
// BOM at column 0 would desynchronise the very first line from
// olaresVersionRE / apiVersionRE and send the dispatch through the wrong
// (legacy vs modern) pipeline. We strip it once at the top of Peek so the
// downstream scanner sees the same first line a BOM-less file would.
var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

// Peek extracts OlaresVersion and APIVersion from content using a
// line-oriented regex scan instead of a full YAML decode.
func Peek(content []byte) (Header, error) {
	var h Header
	scanner := bufio.NewScanner(bytes.NewReader(bytes.TrimPrefix(content, utf8BOM)))
	scanner.Buffer(make([]byte, 0, 64*1024), 1<<20)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if line[0] == ' ' || line[0] == '\t' || line[0] == '-' {
			continue
		}
		if h.OlaresVersion == "" {
			if m := olaresVersionRE.FindStringSubmatch(line); m != nil {
				h.OlaresVersion = cleanScalar(m[1])
				continue
			}
		}
		if h.APIVersion == "" {
			if m := apiVersionRE.FindStringSubmatch(line); m != nil {
				h.APIVersion = cleanScalar(m[1])
				continue
			}
		}
		if h.OlaresVersion != "" && h.APIVersion != "" {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return Header{}, err
	}
	return h, nil
}

func cleanScalar(raw string) string {
	s := strings.TrimSpace(raw)
	if !strings.HasPrefix(s, `"`) && !strings.HasPrefix(s, `'`) {
		if i := strings.Index(s, " #"); i >= 0 {
			s = strings.TrimSpace(s[:i])
		}
	}
	if len(s) >= 2 {
		first, last := s[0], s[len(s)-1]
		if (first == '"' && last == '"') || (first == '\'' && last == '\'') {
			s = s[1 : len(s)-1]
		}
	}
	return s
}

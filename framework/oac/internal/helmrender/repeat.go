package helmrender

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// repeatFn matches a sprig repeat call inside a template action.
// "repeated" does not match. A preceding "." or "$" is a field or variable
// (.Values.repeat, .repeat, $repeat) and is not a call.
var repeatFn = regexp.MustCompile(`(?:^|[^.$0-9A-Za-z_])repeat\b`)

// rejectTemplateRepeat reports the first chart template action that calls
// sprig repeat. Helm evaluates that call while rendering, and an unbounded
// count allocates the whole string before any later check can run.
func rejectTemplateRepeat(oacPath string) error {
	var found error
	err := filepath.Walk(oacPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if found != nil || info == nil || info.IsDir() || !isHelmTemplate(path) {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return fmt.Errorf("read %s: %w", path, readErr)
		}
		line, ok := repeatActionLine(string(bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})))
		if !ok {
			return nil
		}
		rel, relErr := filepath.Rel(oacPath, path)
		if relErr != nil {
			rel = filepath.Base(path)
		}
		found = fmt.Errorf("%s:%d: template must not use repeat", rel, line)
		return nil
	})
	if err != nil {
		return err
	}
	return found
}

func isHelmTemplate(path string) bool {
	for dir := filepath.Dir(path); ; dir = filepath.Dir(dir) {
		if filepath.Base(dir) == "templates" {
			return true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return false
		}
	}
}

// repeatActionLine returns the 1-based line of the first {{ }} action whose
// body calls repeat. Comments ({{/* */}}) and quoted strings are ignored.
func repeatActionLine(content string) (int, bool) {
	line := 1
	for i := 0; i < len(content); {
		if content[i] == '\n' {
			line++
			i++
			continue
		}
		if i+1 >= len(content) || content[i] != '{' || content[i+1] != '{' {
			i++
			continue
		}
		startLine := line
		bodyStart := i + 2
		if bodyStart < len(content) && content[bodyStart] == '-' {
			bodyStart++
		}
		rest := content[bodyStart:]
		trimmed := strings.TrimLeft(rest, " \t\r\n")
		skipped := len(rest) - len(trimmed)
		if strings.HasPrefix(trimmed, "/*") {
			end := strings.Index(content[bodyStart+skipped:], "*/")
			if end < 0 {
				return 0, false
			}
			i = bodyStart + skipped + end + 2
			line += strings.Count(content[bodyStart:i], "\n")
			continue
		}
		end := strings.Index(content[bodyStart:], "}}")
		if end < 0 {
			return 0, false
		}
		body := content[bodyStart : bodyStart+end]
		body = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(body), "-"))
		if repeatFn.MatchString(stripTemplateStrings(body)) {
			return startLine, true
		}
		i = bodyStart + end + 2
		line += strings.Count(content[bodyStart:i], "\n")
	}
	return 0, false
}

// stripTemplateStrings blanks Go template string literals so a quoted
// "repeat" is not treated as a function call.
func stripTemplateStrings(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	var quote byte
	for i := 0; i < len(s); i++ {
		c := s[i]
		if quote == '`' {
			if c == '`' {
				quote = 0
			}
			continue
		}
		if quote == '"' {
			if c == '\\' && i+1 < len(s) {
				i++
				continue
			}
			if c == '"' {
				quote = 0
			}
			continue
		}
		if c == '"' || c == '`' {
			quote = c
			b.WriteByte(' ')
			continue
		}
		b.WriteByte(c)
	}
	return b.String()
}

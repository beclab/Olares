// Package aptsource enables Debian archive components without replacing the
// administrator's source configuration. Parsing is separate from host I/O.
package aptsource

import (
	"fmt"
	"net/url"
	"path"
	"regexp"
	"slices"
	"strings"
)

const managedFile = "sources.list.d/olares-debian.sources"
const debianKeyring = "/usr/share/keyrings/debian-archive-keyring.gpg"

type sourceFile struct {
	name    string
	content string
}

type change struct {
	name   string
	after  string
	exists bool
}

// Only the official archive aliases are interchangeable. In particular, do not
// infer that an arbitrary third-party or private mirror is a Debian archive.
func debianArchive(raw string) bool {
	u, err := url.Parse(raw)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") &&
		strings.EqualFold(u.Host, "deb.debian.org") && u.User == nil &&
		strings.TrimRight(u.Path, "/") == "/debian" && u.RawQuery == "" && u.Fragment == ""
}

// A mirror explicitly restricted to the Debian archive keyring is also a
// Debian source. This includes official cloud images' mirror+file: sources and
// private mirrors, without treating every source for this suite as Debian.
func debianSource(raw, signedBy string) bool {
	return debianArchive(raw) || strings.TrimSpace(signedBy) == debianKeyring
}

func missingComponents(existing string) []string {
	var missing []string
	for _, component := range []string{"contrib", "non-free"} {
		if !slices.Contains(strings.Fields(existing), component) {
			missing = append(missing, component)
		}
	}
	return missing
}

// Release support belongs to the prepare precheck, not the source file format.
// A codename must still be a single safe token when inserted into a stanza.
var suiteName = regexp.MustCompile(`^[a-z0-9][a-z0-9.+-]*$`)

func plan(files []sourceFile, suite string) ([]change, error) {
	if !suiteName.MatchString(suite) {
		return nil, fmt.Errorf("invalid Debian suite %q", suite)
	}
	var changes []change
	matched := false
	for _, file := range files {
		var after string
		var active bool
		var err error
		if strings.HasSuffix(file.name, ".sources") {
			after, active, err = updateDeb822(file.content, suite)
		} else {
			after, active, err = updateList(file.content, suite)
		}
		if err != nil {
			return nil, fmt.Errorf("%s: %w", file.name, err)
		}
		matched = matched || active
		if after != file.content {
			changes = append(changes, change{file.name, after, true})
		}
	}
	if matched {
		return changes, nil
	}
	// The fallback supplements the existing base system; it does not enable main
	// or non-free-firmware, neither of which was requested by this operation.
	after := "# Managed by Olares: additional Debian archive components.\nTypes: deb\nURIs: https://deb.debian.org/debian\nSuites: " + suite + "\nComponents: contrib non-free\nSigned-By: /usr/share/keyrings/debian-archive-keyring.gpg\n"
	for _, file := range files {
		if file.name == managedFile {
			// Do not overwrite a stale release or administrator-edited managed file.
			return nil, fmt.Errorf("%s exists but does not contain an enabled Debian %s archive; inspect it before retrying", managedFile, suite)
		}
	}
	return append(changes, change{name: managedFile, after: after}), nil
}

var listLine = regexp.MustCompile(`^(deb|deb-src)\s+(?:\[[^\]]+\]\s+)?(\S+)\s+(\S+)(?:\s+(.+))?$`)

func listSignedBy(line string) string {
	_, tail, ok := strings.Cut(line, "[")
	if !ok {
		return ""
	}
	options, _, ok := strings.Cut(tail, "]")
	if !ok {
		return ""
	}
	for _, option := range strings.Fields(options) {
		if value, ok := strings.CutPrefix(option, "signed-by="); ok {
			return value
		}
	}
	return ""
}

// Named distributions require components; exact-path (flat) repositories must
// omit them. Apply the same APT rule to both source formats before editing.
func checkComponents(suite, components string) error {
	if strings.HasSuffix(suite, "/") {
		if len(strings.Fields(components)) != 0 {
			return fmt.Errorf("components must be omitted for exact-path suite %q", suite)
		}
	} else if len(strings.Fields(components)) == 0 {
		return fmt.Errorf("missing components for suite %q", suite)
	}
	return nil
}

func updateList(content, suite string) (string, bool, error) {
	lines := strings.SplitAfter(content, "\n")
	matched := false
	for i, original := range lines {
		text := strings.TrimSpace(original)
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}
		body, _, _ := strings.Cut(text, "#")
		m := listLine.FindStringSubmatch(strings.TrimSpace(body))
		if m == nil {
			return "", false, fmt.Errorf("invalid source at line %d", i+1)
		}
		if err := checkComponents(m[3], m[4]); err != nil {
			return "", false, fmt.Errorf("line %d: %w", i+1, err)
		}
		if m[1] != "deb" || !debianSource(m[2], listSignedBy(body)) || m[3] != suite {
			continue
		}
		matched = true
		missing := missingComponents(m[4])
		if len(missing) == 0 {
			continue
		}
		// Insert before trailing whitespace/comment; preserve options verbatim.
		end := len(original)
		if n := strings.IndexByte(original, '#'); n >= 0 {
			end = n
		}
		end = len(strings.TrimRight(original[:end], " \t\r\n"))
		lines[i] = original[:end] + " " + strings.Join(missing, " ") + original[end:]
	}
	return strings.Join(lines, ""), matched, nil
}

type field struct {
	value string
	last  int
}

func updateDeb822(content, suite string) (string, bool, error) {
	lines := strings.SplitAfter(content, "\n")
	matched := false
	for start := 0; start < len(lines); {
		end := start
		for end < len(lines) && strings.TrimSpace(lines[end]) != "" {
			end++
		}
		fields := map[string]field{}
		last := ""
		for i := start; i < end; i++ {
			line := strings.TrimRight(lines[i], "\r\n")
			if strings.HasPrefix(line, "#") {
				continue
			}
			if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
				if last == "" {
					return "", false, fmt.Errorf("unexpected continuation at line %d", i+1)
				}
				f := fields[last]
				f.value += " " + strings.TrimSpace(line)
				f.last = i
				fields[last] = f
				continue
			}
			key, value, ok := strings.Cut(line, ":")
			if !ok || key == "" || strings.ContainsAny(key, " \t") {
				return "", false, fmt.Errorf("invalid field at line %d", i+1)
			}
			key = strings.ToLower(key)
			if _, exists := fields[key]; exists {
				return "", false, fmt.Errorf("duplicate field %q at line %d", key, i+1)
			}
			fields[key] = field{strings.TrimSpace(value), i}
			last = key
		}
		// Disabled entries are preserved, not evidence that all Debian sources
		// are disabled: the administrator may have switched to another mirror.
		if len(fields) != 0 && !strings.EqualFold(fields["enabled"].value, "no") {
			for _, key := range []string{"types", "uris", "suites"} {
				if fields[key].value == "" {
					return "", false, fmt.Errorf("missing %s in stanza at line %d", key, start+1)
				}
			}
			for _, distribution := range strings.Fields(fields["suites"].value) {
				if err := checkComponents(distribution, fields["components"].value); err != nil {
					return "", false, fmt.Errorf("stanza at line %d: %w", start+1, err)
				}
			}
			archive := false
			for _, uri := range strings.Fields(fields["uris"].value) {
				archive = archive || debianSource(uri, fields["signed-by"].value)
			}
			if archive && slices.Contains(strings.Fields(fields["types"].value), "deb") && slices.Contains(strings.Fields(fields["suites"].value), suite) {
				if v := fields["enabled"].value; v != "" && !strings.EqualFold(v, "yes") {
					return "", false, fmt.Errorf("invalid Enabled value %q", v)
				}
				matched = true
				f := fields["components"]
				if missing := missingComponents(f.value); len(missing) != 0 {
					line := lines[f.last]
					pos := len(strings.TrimRight(line, " \t\r\n"))
					lines[f.last] = line[:pos] + " " + strings.Join(missing, " ") + line[pos:]
				}
			}
		}
		start = end + 1
	}
	return strings.Join(lines, ""), matched, nil
}

func validName(name string) bool {
	if name == "sources.list" {
		return true
	}
	if path.Dir(name) != "sources.list.d" {
		return false
	}
	base := path.Base(name)
	if !strings.HasSuffix(base, ".list") && !strings.HasSuffix(base, ".sources") {
		return false
	}
	for _, c := range base {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || strings.ContainsRune("_.-", c)) {
			return false
		}
	}
	return true
}

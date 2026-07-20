package manifest

import (
	"bytes"
	"fmt"
	"regexp"
)

// templateSyntaxRE matches the two helm template delimiters `{{` and
// `}}` independently. The check fires per-occurrence rather than per
// pair: even a lone `{{` or `}}` that escapes the YAML string masker
// almost always means "this used to be a helm placeholder before the
// rendering step went away", and we want to surface it the same way
// as the full `{{ ... }}` form. The regex runs against a *masked*
// copy of the manifest in which braces inside YAML string scalars
// have been replaced with spaces, so any match here means the
// delimiter sits outside every key's value scalar.
var templateSyntaxRE = regexp.MustCompile(`\{\{|\}\}`)

// checkNoTemplateSyntax returns an error if the manifest contains a
// `{{` or `}}` delimiter that does not belong to any key's value
// scalar — practically, any occurrence outside a single-quoted scalar,
// a double-quoted scalar, or a `|`/`>` block scalar. Once a delimiter
// is properly contained inside a YAML string, YAML treats its body as
// opaque text, so even the full `{{ ... }}` pair is fine there and
// reads as documentation for the manifest author.
//
// The check is gated by the caller to modern manifests
// (olaresManifest.version >= 0.12.0): those manifests are parsed
// verbatim without a helm render pass, so an unquoted placeholder
// would surface as a confusing low-level YAML error (e.g. "yaml: line
// 78: could not find expected ':'"). Detecting the placeholder up
// front turns that into a clear, actionable manifest authoring error.
func checkNoTemplateSyntax(content []byte) error {
	masked := maskYAMLStringRegions(content)
	loc := templateSyntaxRE.FindIndex(masked)
	if loc == nil {
		return nil
	}
	line := bytes.Count(content[:loc[0]], []byte{'\n'}) + 1
	matched := content[loc[0]:loc[1]]
	return fmt.Errorf(
		"line %d: %q is template syntax outside any YAML string value and is not allowed for olaresManifest.version >= %s; quote the value or move it into a `|`/`>` block scalar (literal `{{` / `}}` / `{{ ... }}` are still allowed *inside* YAML strings)",
		line, string(matched), minResourcesManifestVersion,
	)
}

// maskYAMLStringRegions returns a copy of content in which every `{`
// and `}` that sits inside a YAML string scalar has been replaced with
// a space. Line offsets and the position of every non-brace byte are
// preserved, so callers that find a match in the masked buffer can use
// the same indices to report a line number against the original input.
//
// The masker is intentionally a hand-rolled line-by-line scanner
// instead of a real YAML lexer: the whole point of the gate is that
// the input may not parse as YAML yet (a stray `{{ ... }}` would
// derail an honest decode). The scanner covers the three shapes that
// practically appear in OlaresManifest.yaml:
//
//   - double-quoted scalars `"..."` (with `\` escapes that consume the
//     next byte);
//   - single-quoted scalars `'...'` (with `''` representing a literal
//     single quote);
//   - `|` / `>` block scalars whose content lines are indented strictly
//     more than the header line.
//
// Quoted scalars are assumed to live on a single logical line; if a
// literal newline appears inside one, the scanner exits the string at
// end-of-line, which is a deliberate conservative choice — true
// multi-line quoted strings are rare in OlaresManifest.yaml and the
// fallback ("treat as outside string") only ever causes a false
// positive that the author can fix by switching to a block scalar.
//
// Flow style (`{key: value}` / `[a, b]`) is *not* masked. Those
// constructs use literal single `{` / `}` / `[` / `]`, but the regex
// looks for `{{` and `}}`, so the flow markers do not collide with
// template detection.
func maskYAMLStringRegions(content []byte) []byte {
	out := make([]byte, len(content))
	copy(out, content)

	// blockScalarIndent is the indent of the header line that opened
	// the active `|`/`>` scalar; content lines whose indent is strictly
	// greater than this value belong to the scalar. -1 means "no
	// active block scalar".
	blockScalarIndent := -1

	i := 0
	for i < len(out) {
		lineStart := i
		lineEnd := lineStart
		for lineEnd < len(out) && out[lineEnd] != '\n' {
			lineEnd++
		}

		// CRLF-authored manifests (common when files are produced or
		// round-tripped on Windows) leave a trailing `\r` before each
		// `\n`. Treat it as part of the line terminator: a `\r` sitting
		// after a `|` / `>` block-scalar header would otherwise look
		// like garbage to isBlockScalarHeaderTail, suppress the header,
		// and leak braces inside the body. contentEnd is the slice
		// bound the per-line scanners see; lineEnd is still the
		// next-line cursor so byte offsets remain stable.
		contentEnd := lineEnd
		if contentEnd > lineStart && out[contentEnd-1] == '\r' {
			contentEnd--
		}

		indent := 0
		for lineStart+indent < contentEnd && out[lineStart+indent] == ' ' {
			indent++
		}
		blank := lineStart+indent == contentEnd

		switch {
		case blockScalarIndent >= 0 && (blank || indent > blockScalarIndent):
			maskBracesInRange(out, lineStart, contentEnd)
		default:
			if blockScalarIndent >= 0 {
				blockScalarIndent = -1
			}
			if !blank {
				if openedBlockScalar := maskQuotedRegionsOnLine(out, lineStart, contentEnd); openedBlockScalar {
					blockScalarIndent = indent
				}
			}
		}

		i = lineEnd
		if i < len(out) && out[i] == '\n' {
			i++
		}
	}
	return out
}

// maskBracesInRange replaces every `{` and `}` between buf[start:end]
// with a space. Used for lines that sit entirely inside a `|`/`>`
// block scalar so any helm-style placeholders inside read as
// documentation text rather than template syntax.
func maskBracesInRange(buf []byte, start, end int) {
	for k := start; k < end; k++ {
		if buf[k] == '{' || buf[k] == '}' {
			buf[k] = ' '
		}
	}
}

// maskQuotedRegionsOnLine walks one logical line and (a) masks braces
// that fall inside `"..."` / `'...'` scalars and (b) reports whether
// the line opens a `|`/`>` block scalar that the caller needs to
// honour on subsequent lines. Comments (`# ...`) are skipped so a
// `#`-prefixed pipe in a comment does not get mistaken for a block
// scalar header.
func maskQuotedRegionsOnLine(buf []byte, start, end int) bool {
	inDQ := false
	inSQ := false
	openedBlockScalar := false
	for j := start; j < end; j++ {
		c := buf[j]
		switch {
		case inDQ:
			if c == '\\' && j+1 < end {
				if buf[j+1] == '{' || buf[j+1] == '}' {
					buf[j+1] = ' '
				}
				j++
				continue
			}
			if c == '"' {
				inDQ = false
			} else if c == '{' || c == '}' {
				buf[j] = ' '
			}
		case inSQ:
			if c == '\'' {
				if j+1 < end && buf[j+1] == '\'' {
					j++
					continue
				}
				inSQ = false
			} else if c == '{' || c == '}' {
				buf[j] = ' '
			}
		default:
			switch c {
			case '"':
				inDQ = true
			case '\'':
				inSQ = true
			case '#':
				return openedBlockScalar
			case '|', '>':
				if isBlockScalarHeaderTail(buf, j+1, end) {
					openedBlockScalar = true
				}
			}
		}
	}
	return openedBlockScalar
}

// isBlockScalarHeaderTail reports whether buf[start:end] is a valid
// trailer for a `|` or `>` block scalar header: only chomping
// indicators (`+` / `-`), optional explicit indent digits, surrounding
// whitespace, and an optional trailing `# ...` comment are permitted
// before end-of-line. Anything else (e.g. a letter, a `:`, another
// value) means the `|` / `>` we just saw is an ordinary scalar
// character, not a header.
func isBlockScalarHeaderTail(buf []byte, start, end int) bool {
	for k := start; k < end; k++ {
		c := buf[k]
		switch {
		case c == ' ', c == '\t', c == '+', c == '-', c >= '0' && c <= '9':
			continue
		case c == '#':
			return true
		default:
			return false
		}
	}
	return true
}

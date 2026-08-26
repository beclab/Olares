package download

// Naming rules shared by create (which file_name to send) and by the
// renderers (what to show while the server has not written one yet).
//
// Both halves mirror code that already exists on the other side of the
// wire, so one task reads the same in the CLI, in LarePass and in the
// Olares Download list: the byte budget comes from the ytdlp daemon's
// outtmpl sanitiser, the display fallback from download-server's
// frontend fallbackName / LarePass downloadDisplayName.

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"
)

const (
	// fileNameMaxChars / fileNameMaxBytes are the ytdlp daemon's
	// file_name budget (ytdlp/src/core/outtmpl.py): a 255-byte path
	// component, minus 48 bytes reserved for the transient suffixes
	// yt-dlp opens while downloading (.f399, .part, .part-Frag12),
	// minus the 8-byte ".%(ext)s" the daemon splices on. 66 is that
	// byte budget in CJK characters — the unit a title is counted in;
	// the byte cap still applies because 4-byte runes fit under 66
	// yet overrun the daemon, whose open() then fails ENAMETOOLONG.
	fileNameMaxChars = 66
	fileNameMaxBytes = 255 - 48 - 8

	extTemplate = ".%(ext)s"
)

// titleToFileName turns an inspect title into a file_name safe to send
// on create, or "" when nothing usable is left.
//
// Path separators become "_" because the daemon splices file_name into
// the outtmpl and both os.path.join and yt-dlp's template engine read a
// slash as a directory boundary — a YouTube title like "Foo / Bar" then
// lands as a nested Foo_/Bar.webm tree on the user's volume.
// "%(" becomes "_(" for the same reason: it opens an outtmpl field, so
// a title like "100%(백퍼센트)" made yt-dlp look up a field of that name
// and fail the task with err_msg "'백퍼센트'" before a byte was written.
// Only that two-character sequence is neutralised — a lone "%" is a
// literal to yt-dlp ("100% Real" lands verbatim) and rewriting it would
// mangle a name that works. Doubling to "%%(" would also parse, but the
// row would then read "%%(" where the disk reads "%(", so replacing is
// what keeps the two in step. A trailing ".%(ext)s" is held back, since
// that one IS the daemon's own extension template.
// Everything else, including ":" and non-ASCII, is legal on Linux and
// kept as is. A bare "." / ".." is dropped: the daemon rejects those as
// traversal and falls back to its own template, so sending one would
// only pin a name to the task row that never appears on disk.
func titleToFileName(title string) string {
	sanitised := strings.Map(func(r rune) rune {
		switch r {
		case '/', '\\':
			return '_'
		case 0:
			return -1
		}
		return r
	}, strings.TrimSpace(title))
	sanitised = defuseOuttmplFields(sanitised)
	if sanitised == "." || sanitised == ".." {
		return ""
	}
	return truncateFileNameToBudget(sanitised)
}

// outtmplFieldOpen is the only percent form yt-dlp reads as syntax: it
// starts a "%(field)s" reference, whose lookup either fails the task
// (unknown field, unterminated key) or silently expands to something
// other than the name the row advertises.
const outtmplFieldOpen = "%("

// defuseOuttmplFields neutralises field references in the stem, leaving
// the extension — literal or the daemon's ".%(ext)s" template — alone.
// Doing this before truncation also means a clipped stem can never end
// mid-"%(field)s", which yt-dlp rejects as an incomplete format key.
func defuseOuttmplFields(name string) string {
	stem, ext := splitOuttmplExt(name)
	return strings.ReplaceAll(stem, outtmplFieldOpen, "_(") + ext
}

// validateOuttmplSafeName refuses a --name that yt-dlp would read as a
// template. A probed title is rewritten silently (the user did not type
// it), but a name the caller chose is rejected the same way --name is
// for a magnet: being told the name cannot stick beats having it
// changed, or having the task fail with a bare field name for an error.
func validateOuttmplSafeName(name string) error {
	stem, _ := splitOuttmplExt(name)
	if !strings.Contains(stem, outtmplFieldOpen) {
		return nil
	}
	return fmt.Errorf(
		"unsupported --name %q: %q opens a yt-dlp output-template field, so the download fails with the field name as its error (or silently lands under a different name); use a literal name without %q",
		name, outtmplFieldOpen, outtmplFieldOpen,
	)
}

// truncateFileNameToBudget clips a name to both caps at once. When an
// extension is recognised only the stem is eaten, so the suffix that
// survives is the one the daemon keeps.
func truncateFileNameToBudget(name string) string {
	if utf8.RuneCountInString(name) <= fileNameMaxChars && len(name) <= fileNameMaxBytes {
		return name
	}
	stem, ext := splitOuttmplExt(name)
	return truncateStem(
		stem,
		fileNameMaxChars-utf8.RuneCountInString(ext),
		fileNameMaxBytes-len(ext),
	) + ext
}

// splitOuttmplExt mirrors the daemon's _split_outtmpl_ext: a literal
// short extension or the ".%(ext)s" template is held back from
// truncation.
func splitOuttmplExt(name string) (stem, ext string) {
	if strings.HasSuffix(name, extTemplate) {
		return strings.TrimSuffix(name, extTemplate), extTemplate
	}
	lastDot := strings.LastIndex(name, ".")
	if lastDot >= 0 {
		candidate := name[lastDot+1:]
		length := utf8.RuneCountInString(candidate)
		if length >= 1 && length <= 10 && !strings.Contains(candidate, "%") {
			return name[:lastDot], name[lastDot:]
		}
	}
	return name, ""
}

func truncateStem(stem string, maxChars, maxBytes int) string {
	if maxChars <= 0 || maxBytes <= 0 {
		return ""
	}
	chars, bytes := 0, 0
	var out strings.Builder
	for _, r := range stem {
		size := utf8.RuneLen(r)
		if size < 0 {
			size = 1
		}
		if chars+1 > maxChars || bytes+size > maxBytes {
			break
		}
		chars++
		bytes += size
		out.WriteRune(r)
	}
	return out.String()
}

// displayName is the task name for table output. It mirrors
// download-server's fallbackName and LarePass's downloadDisplayName:
// file_name, then the torrent metadata name, then the magnet's dn= /
// info-hash, then a URL basename that actually looks like a file.
//
// It stops there rather than falling back to the raw URL. A routing
// segment such as YouTube's /watch is not a name, and list / info print
// the URL in their own column anyway — showing it here would only read
// as a filename the server never wrote.
func displayName(t DownloadTask) string {
	if name := strings.TrimRight(strings.TrimSpace(t.FileName), "/"); name != "" {
		return name
	}
	if meta := torrentMetaName(t.Extra); meta != "" {
		return meta
	}
	raw := strings.TrimSpace(t.URL)
	if raw == "" {
		return "-"
	}
	if isMagnetURL(raw) {
		if dn := magnetDisplayName(raw); dn != "" {
			return dn
		}
		if hash := magnetInfoHash(raw); hash != "" {
			if len(hash) > 12 {
				hash = hash[:12] + "…"
			}
			return "magnet:" + hash
		}
		return "magnet:…"
	}
	if base := urlBasename(raw); base != "" {
		return base
	}
	return "-"
}

// torrentMetaName reads extra.torrent_meta.name, which the manager
// persists once aria2 has parsed a magnet's metadata — the first real
// name a magnet row ever gets.
func torrentMetaName(extra map[string]interface{}) string {
	meta, ok := extra["torrent_meta"].(map[string]interface{})
	if !ok {
		return ""
	}
	name, _ := meta["name"].(string)
	return strings.TrimSpace(name)
}

func isMagnetURL(raw string) bool {
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(raw)), "magnet:")
}

// isHuggingFaceURL matches what the manager routes to `hf download`:
// the huggingface.co / hf.co hosts plus the hf:// and huggingface:
// pseudo-schemes it still accepts.
func isHuggingFaceURL(raw string) bool {
	s := strings.ToLower(strings.TrimSpace(raw))
	if s == "" {
		return false
	}
	if strings.HasPrefix(s, "hf://") || strings.HasPrefix(s, "huggingface:") {
		return true
	}
	u, err := url.Parse(s)
	if err != nil {
		return false
	}
	host := strings.TrimPrefix(u.Hostname(), "www.")
	return host == "huggingface.co" || host == "hf.co"
}

// magnet: has no authority component, so url.Parse cannot be trusted to
// find the query — it has to be split off by hand.
func magnetParams(raw string) url.Values {
	q := strings.Index(raw, "?")
	if q < 0 {
		return nil
	}
	// A malformed escape yields the pairs that did parse alongside the
	// error; those are still the best guess at a display name.
	values, _ := url.ParseQuery(raw[q+1:])
	return values
}

func magnetDisplayName(raw string) string {
	return strings.TrimSpace(magnetParams(raw).Get("dn"))
}

func magnetInfoHash(raw string) string {
	const prefix = "urn:btih:"
	for _, xt := range magnetParams(raw)["xt"] {
		if strings.HasPrefix(strings.ToLower(xt), prefix) {
			return strings.TrimSpace(xt[len(prefix):])
		}
	}
	return ""
}

// fileNameExtRE is the "looks like a downloadable file" rule: a 1-5
// character alphanumeric extension. Requiring it is what keeps routing
// verbs (YouTube /watch, Bilibili /video/BV…) out of a task name.
var fileNameExtRE = regexp.MustCompile(`\.[A-Za-z0-9]{1,5}$`)

func looksLikeFileName(segment string) bool {
	return fileNameExtRE.MatchString(segment)
}

// urlBasename returns the last path segment when it looks like a file,
// else "". url.Parse has already percent-decoded Path, and the query /
// fragment are off it, so a signed CDN link still yields a clean guess.
func urlBasename(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	segments := strings.Split(u.Path, "/")
	for i := len(segments) - 1; i >= 0; i-- {
		if segments[i] == "" {
			continue
		}
		if looksLikeFileName(segments[i]) {
			return segments[i]
		}
		return ""
	}
	return ""
}

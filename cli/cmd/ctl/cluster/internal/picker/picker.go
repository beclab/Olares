// Package picker is a tiny, dependency-free interactive list picker built on
// golang.org/x/term (already used by `cluster exec` for raw mode) and
// github.com/mattn/go-runewidth (already in the module graph via tabwriter).
//
// It powers `cluster {pod,container} exec -it` with no target: a single flat,
// type-to-filter list of namespace/pod/container entries. The pure pieces
// (Sort, Filter, window) are unit-tested; the raw-mode render/key loop in Pick
// is a thin shell over them.
package picker

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/mattn/go-runewidth"
	"golang.org/x/term"
)

// ErrCanceled is returned by Pick when the user aborts with Esc or Ctrl-C.
// Callers treat it as a clean, no-op exit (not an error).
var ErrCanceled = errors.New("selection canceled")

// Entry is one selectable target: a single container within a pod.
type Entry struct {
	Namespace string
	Pod       string
	Container string
	Running   bool   // pod phase == Running (non-running rows are dimmed)
	Status    string // human status shown after the label, e.g. "Running"
}

// label is the primary text shown/filtered for an entry.
func (e Entry) label() string {
	return e.Namespace + "/" + e.Pod + " \u203a " + e.Container
}

// haystack is the lowercase string Filter matches a query against.
func (e Entry) haystack() string {
	return strings.ToLower(e.Namespace + "/" + e.Pod + "/" + e.Container)
}

// Sort orders entries Running-first, then by namespace, pod, container. Stable
// so equal keys keep their fetch order.
func Sort(entries []Entry) {
	sort.SliceStable(entries, func(i, j int) bool {
		a, b := entries[i], entries[j]
		if a.Running != b.Running {
			return a.Running // running (true) sorts before not-running
		}
		if a.Namespace != b.Namespace {
			return a.Namespace < b.Namespace
		}
		if a.Pod != b.Pod {
			return a.Pod < b.Pod
		}
		return a.Container < b.Container
	})
}

// Filter returns the entries matching query (see matchScore for the model:
// whitespace-tokenized, order-independent, per-token substring), sorted by
// relevance score descending. Ties preserve the input order (which callers set
// to Running-first / alpha via Sort), so equally-relevant rows keep a stable,
// predictable arrangement. An empty query returns the input unchanged.
func Filter(entries []Entry, query string) []Entry {
	if strings.TrimSpace(query) == "" {
		return entries
	}
	type scored struct {
		e     Entry
		score int
		idx   int
	}
	matches := make([]scored, 0, len(entries))
	for i, e := range entries {
		if s, ok := matchScore(e.haystack(), query); ok {
			matches = append(matches, scored{e: e, score: s, idx: i})
		}
	}
	sort.SliceStable(matches, func(a, b int) bool {
		if matches[a].score != matches[b].score {
			return matches[a].score > matches[b].score
		}
		return matches[a].idx < matches[b].idx // stable: keep Running-first/alpha
	})
	out := make([]Entry, len(matches))
	for i := range matches {
		out[i] = matches[i].e
	}
	return out
}

// window returns the [start,end) slice of a list of n items to render in a
// viewport of height rows, keeping cursor visible and roughly centered. Pure
// and unit-tested so the render loop's scrolling can't drift.
func window(n, cursor, height int) (start, end int) {
	if height <= 0 || n == 0 {
		return 0, 0
	}
	if height >= n {
		return 0, n
	}
	start = cursor - height/2
	if start < 0 {
		start = 0
	}
	if start+height > n {
		start = n - height
	}
	return start, start + height
}

const (
	minViewport = 3
	maxViewport = 15
)

// Pick runs the interactive picker over entries, rendering to stderr and
// reading keys from stdin in raw mode. Returns the chosen Entry, or ErrCanceled
// if the user aborts. Requires a terminal on stdin.
func Pick(entries []Entry, header string) (Entry, error) {
	inFd := int(os.Stdin.Fd())
	if !term.IsTerminal(inFd) {
		return Entry{}, errors.New("interactive picker requires a terminal")
	}
	oldState, err := term.MakeRaw(inFd)
	if err != nil {
		return Entry{}, fmt.Errorf("enter raw mode: %w", err)
	}
	defer term.Restore(inFd, oldState)

	out := os.Stderr
	fmt.Fprint(out, "\033[?25l")       // hide cursor
	defer fmt.Fprint(out, "\033[?25h") // show cursor

	query := ""
	cursor := 0
	prevLines := 0

	render := func() {
		filtered := Filter(entries, query)
		if cursor > len(filtered)-1 {
			cursor = len(filtered) - 1
		}
		if cursor < 0 {
			cursor = 0
		}

		width, height := 80, 24
		if w, h, gerr := term.GetSize(int(out.Fd())); gerr == nil && w > 0 && h > 0 {
			width, height = w, h
		}
		viewport := height - 3 // header + filter + footer
		if viewport < minViewport {
			viewport = minViewport
		}
		if viewport > maxViewport {
			viewport = maxViewport
		}
		start, end := window(len(filtered), cursor, viewport)

		var b strings.Builder
		lines := 0
		// emit writes one screen line: plain text is truncated to the terminal
		// width FIRST (runewidth-aware), then wrapped in the SGR style so escape
		// bytes never count toward the width and lines never wrap (wrapping
		// would break the line-count bookkeeping used to clear the frame).
		emit := func(plain, sgr string) {
			if lines > 0 {
				b.WriteString("\r\n")
			}
			t := truncate(plain, width)
			if sgr != "" {
				b.WriteString("\033[" + sgr + "m" + t + "\033[0m")
			} else {
				b.WriteString(t)
			}
			lines++
		}

		// Return to the top of the previous frame and clear downward.
		if prevLines > 0 {
			if prevLines > 1 {
				fmt.Fprintf(&b, "\033[%dA", prevLines-1)
			}
			b.WriteString("\r\033[J")
		}

		emit(header, "2")
		// Filter line: bold "❯ query" (leave a column for the caret), then a
		// reverse-video space as a faux caret.
		if lines > 0 {
			b.WriteString("\r\n")
		}
		b.WriteString("\033[1m" + truncate("\u276f "+query, max(width-1, 1)) + "\033[0m\033[7m \033[0m")
		lines++

		if len(filtered) == 0 {
			emit("  (no matches)", "2")
		}
		for i := start; i < end; i++ {
			e := filtered[i]
			row := "  " + e.label() + "  (" + e.Status + ")"
			switch {
			case i == cursor:
				emit("\u276f "+e.label()+"  ("+e.Status+")", "7") // reverse video
			case !e.Running:
				emit(row, "2") // dim
			default:
				emit(row, "")
			}
		}
		emit(fmt.Sprintf("  [%d/%d]  \u2191\u2193 move \u00b7 filter (space = AND, any order) \u00b7 enter select \u00b7 esc cancel",
			min(cursor+1, len(filtered)), len(filtered)), "2")

		fmt.Fprint(out, b.String())
		prevLines = lines
	}

	clear := func() {
		if prevLines > 1 {
			fmt.Fprintf(out, "\033[%dA", prevLines-1)
		}
		fmt.Fprint(out, "\r\033[J")
		prevLines = 0
	}

	render()

	// selectCurrent returns the highlighted entry, or (zero,false) when the
	// filtered list is empty (Enter is a no-op then).
	selectCurrent := func() (Entry, bool) {
		filtered := Filter(entries, query)
		if len(filtered) == 0 || cursor < 0 || cursor >= len(filtered) {
			return Entry{}, false
		}
		return filtered[cursor], true
	}

	buf := make([]byte, 64)
	for {
		n, rerr := os.Stdin.Read(buf)
		if rerr != nil {
			clear()
			return Entry{}, rerr
		}
		if n == 0 {
			continue
		}
		b := buf[:n]

		// Arrow keys arrive as a standalone ESC '[' 'A'|'B' read.
		if n >= 3 && b[0] == 0x1b && b[1] == '[' {
			switch b[2] {
			case 'A':
				cursor--
			case 'B':
				cursor++
			}
			render()
			continue
		}

		// Otherwise walk the buffer byte-by-byte so a chunk that mixes filter
		// text with a control key (e.g. a pasted "name\r", or terminals that
		// batch keystrokes) is handled correctly instead of dropping the
		// trailing Enter. Printable bytes accumulate into a run that is decoded
		// once (UTF-8 safe); control bytes act immediately.
		var text []byte
		flush := func() {
			if len(text) == 0 {
				return
			}
			// Drop any stray control runes that slipped through.
			s := strings.Map(func(r rune) rune {
				if r < 0x20 {
					return -1
				}
				return r
			}, string(text))
			if s != "" {
				query += s
				cursor = 0
			}
			text = text[:0]
		}

		for i := 0; i < n; i++ {
			switch c := b[i]; c {
			case 3, 0x1b: // Ctrl-C / Esc → cancel
				clear()
				return Entry{}, ErrCanceled
			case '\r', '\n': // Enter → select
				flush()
				if sel, ok := selectCurrent(); ok {
					clear()
					return sel, nil
				}
			case 127, 8: // Backspace / DEL
				flush()
				if query != "" {
					r := []rune(query)
					query = string(r[:len(r)-1])
					cursor = 0
				}
			default:
				text = append(text, c)
			}
		}
		flush()
		render()
	}
}

// truncate clamps s to width display columns (runewidth-aware), appending "…"
// when it overflows. Always call this on PLAIN text (no ANSI escapes) so the
// width accounting stays correct.
func truncate(s string, width int) string {
	if width <= 0 {
		return s
	}
	return runewidth.Truncate(s, width, "\u2026")
}

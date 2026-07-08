package picker

import (
	"fmt"
	"os"
	"sync"
	"time"

	"golang.org/x/term"
)

// spinnerFrames is a Braille dot cycle — compact, single-column, and legible
// on both light and dark terminals.
var spinnerFrames = []string{"\u280b", "\u2819", "\u2839", "\u2838", "\u283c", "\u2834", "\u2826", "\u2827", "\u2807", "\u280f"}

// Spinner is a minimal, dependency-free stderr progress indicator used while a
// slow/large fetch runs before the picker can render. message is re-evaluated
// on every frame so callers can surface live progress (e.g. a growing count).
//
// It is a no-op when stderr is not a terminal (scripts / redirected output),
// and Stop is always safe to call exactly once.
type Spinner struct {
	stop chan struct{}
	done chan struct{}
	once sync.Once
}

// StartSpinner begins animating message on stderr every ~100ms. Returns a
// Spinner whose Stop clears the line and halts the animation. A nil message or
// non-TTY stderr yields a no-op spinner.
func StartSpinner(message func() string) *Spinner {
	s := &Spinner{stop: make(chan struct{}), done: make(chan struct{})}
	if message == nil || !term.IsTerminal(int(os.Stderr.Fd())) {
		close(s.done)
		return s
	}
	go s.run(message)
	return s
}

func (s *Spinner) run(message func() string) {
	defer close(s.done)
	out := os.Stderr
	fmt.Fprint(out, "\033[?25l")       // hide cursor
	defer fmt.Fprint(out, "\033[?25h") // restore cursor

	draw := func(i int) {
		line := spinnerFrames[i%len(spinnerFrames)] + " " + message()
		// Truncate to the terminal width so a long message can't wrap — a
		// wrapped line would defeat the single-line \r\033[K redraw and leave
		// residue.
		if w, _, gerr := term.GetSize(int(out.Fd())); gerr == nil && w > 0 {
			line = truncate(line, w)
		}
		fmt.Fprintf(out, "\r\033[K\033[2m%s\033[0m", line)
	}
	draw(0) // paint immediately so there's feedback before the first tick

	t := time.NewTicker(100 * time.Millisecond)
	defer t.Stop()
	for i := 1; ; i++ {
		select {
		case <-s.stop:
			fmt.Fprint(out, "\r\033[K") // erase the spinner line
			return
		case <-t.C:
			draw(i)
		}
	}
}

// Stop halts the spinner, clears its line, and blocks until the goroutine has
// finished (so the terminal is clean before the caller writes anything else).
// Safe to call once; extra calls after the first are harmless.
func (s *Spinner) Stop() {
	s.once.Do(func() { close(s.stop) })
	<-s.done
}

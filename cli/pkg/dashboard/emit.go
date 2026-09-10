package dashboard

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/beclab/Olares/cli/pkg/clierr"
)

// ErrAlreadyReported is the sentinel returned after a dashboard command has
// already emitted its envelope or per-iteration error. The dashboard root's
// leaf-error wrapper skips the redundant Fprintln while still propagating the
// error so Cobra exits non-zero.
//
// It wraps clierr.ErrAlreadyReported so cmd/main.go recognises it and
// exits without printing the sentinel's own text.
var ErrAlreadyReported = fmt.Errorf("dashboard: error %w", clierr.ErrAlreadyReported)

// EmitDefault is a tiny helper for leaf commands that don't have custom
// table columns: emit JSON in JSON mode, fall back to a generic key /
// value dump in table mode (sorted column headers based on the union of
// all items' Display keys). Most leaves prefer their own TableColumn
// slice and don't call this — it's the catch-all for free-form GPU /
// task detail responses where the column set isn't fixed.
//
// Hoisted to the pkg layer so cmd-area subpackages don't each redeclare
// it. Settings precedent allows light duplication, but this one is
// non-trivial enough (~25 lines) to centralize.
func EmitDefault(env Envelope, fmtMode OutputFormat) error {
	if fmtMode == OutputJSON {
		return WriteJSON(os.Stdout, env)
	}
	if len(env.Items) == 0 {
		fmt.Println("(no items)")
		return nil
	}
	keys := map[string]struct{}{}
	for _, it := range env.Items {
		for k := range it.Display {
			keys[k] = struct{}{}
		}
	}
	headers := make([]string, 0, len(keys))
	for k := range keys {
		headers = append(headers, k)
	}
	sort.Strings(headers)
	cols := make([]TableColumn, len(headers))
	for i, h := range headers {
		key := h
		cols[i] = TableColumn{
			Header: strings.ToUpper(key),
			Get:    func(it Item) string { return DisplayString(it, key) },
		}
	}
	return WriteTable(os.Stdout, cols, env.Items)
}

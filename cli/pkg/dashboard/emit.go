package dashboard

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
)

// ErrAlreadyReported is the sentinel cmd subpackages return when their
// RunE has already written a user-visible diagnostic to stderr (for
// example, the `unknownSubcommandRunE` helper that prints a typo
// suggestion before returning). The dashboard root's leaf-error wrapper
// (cmd/ctl/dashboard/root.go::wrapLeafErrors) checks for this with
// errors.Is and skips the redundant Fprintln, while still propagating
// the error up so cobra exits non-zero.
var ErrAlreadyReported = errors.New("dashboard: error already reported")

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

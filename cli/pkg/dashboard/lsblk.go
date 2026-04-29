package dashboard

import (
	"sort"
	"strings"
)

// ----------------------------------------------------------------------------
// lsblk tree (overview disk partitions)
// ----------------------------------------------------------------------------
//
// LsblkRow is the canonical shape extracted from one
// `node_disk_lsblk_info.data.result[].metric` entry. Mirrors the
// `LsblkMetricRow` type in `Overview2/Disk/config.ts:13`.

// LsblkRow is one row pulled from the per-node lsblk metric.
type LsblkRow struct {
	Name         string
	Node         string
	Pkname       string
	Size         string
	Fstype       string
	Mountpoint   string
	Fsused       string
	FsusePercent string
}

// LsblkFlatRow is one rendered row in the partitions table. `Depth` is 0
// for the root and increments per nesting level; `TreePrefix` is the
// ASCII-art prefix to prepend to `Name` for human display. `Parent`
// carries the resolved parent name so agents can rebuild the tree from
// the flat list without re-reading pkname / prefix logic.
type LsblkFlatRow struct {
	Row        LsblkRow
	Parent     string
	Depth      int
	TreePrefix string
}

// HasPknameLabels mirrors `Overview2/Disk/config.ts:267` —
// trim/non-empty pkname on at least one row turns the resolver onto the
// label-aware path; otherwise we fall back to prefix matching.
func HasPknameLabels(rows []LsblkRow) bool {
	for _, r := range rows {
		if strings.TrimSpace(r.Pkname) != "" {
			return true
		}
	}
	return false
}

// CollectSubtreeByPkname BFS-walks the pkname graph from `rootName`,
// returning the rows in their original order. Mirrors
// `collectSubtreeByPkname` in Overview2/Disk/config.ts:273.
//
// When the BFS hits an empty `seen` set (e.g. root itself absent from
// the rows) the SPA recursively gathers descendants directly — we
// replicate that fallback so empty rooted views still produce sane data.
func CollectSubtreeByPkname(allRows []LsblkRow, rootName string) []LsblkRow {
	byName := map[string]bool{}
	for _, r := range allRows {
		byName[r.Name] = true
	}
	seen := map[string]bool{}
	queue := []string{}
	if byName[rootName] {
		seen[rootName] = true
		queue = append(queue, rootName)
	}
	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]
		for _, r := range allRows {
			pk := strings.TrimSpace(r.Pkname)
			if pk == n && !seen[r.Name] {
				seen[r.Name] = true
				queue = append(queue, r.Name)
			}
		}
	}
	if len(seen) == 0 {
		var addDesc func(parent string)
		addDesc = func(parent string) {
			for _, r := range allRows {
				pk := strings.TrimSpace(r.Pkname)
				if pk == parent && !seen[r.Name] {
					seen[r.Name] = true
					addDesc(r.Name)
				}
			}
		}
		addDesc(rootName)
	}
	out := make([]LsblkRow, 0, len(allRows))
	for _, r := range allRows {
		if seen[r.Name] {
			out = append(out, r)
		}
	}
	return out
}

// ResolveParent picks the parent for a row, mirroring
// `Overview2/Disk/config.ts:313`:
//
//  1. root row has no parent
//  2. trimmed pkname wins if it points at a row in the set
//  3. otherwise pick the longest other-name prefix of `r.Name`
//  4. last-ditch fallback: the root itself
func ResolveParent(r LsblkRow, rootName string, nameSet map[string]bool) string {
	if r.Name == rootName {
		return ""
	}
	pk := strings.TrimSpace(r.Pkname)
	if pk != "" && nameSet[pk] {
		return pk
	}
	bestPrefix := ""
	for n := range nameSet {
		if n == "" || n == r.Name {
			continue
		}
		if strings.HasPrefix(r.Name, n) && len(n) > len(bestPrefix) {
			bestPrefix = n
		}
	}
	if bestPrefix != "" {
		return bestPrefix
	}
	if nameSet[rootName] {
		return rootName
	}
	return ""
}

// BuildLsblkTreePrefix mirrors the SPA's `buildLsblkTreePrefix`
// (Overview2/Disk/config.ts:332). `lastStack[i]==true` means the
// ancestor at depth `i` is the last sibling at that level — we draw
// "    " under it so the trunk doesn't dribble down past a finished
// sibling.
func BuildLsblkTreePrefix(depth int, lastStack []bool) string {
	if depth == 0 {
		return ""
	}
	var b strings.Builder
	for i := 0; i < depth-1; i++ {
		if i < len(lastStack) && lastStack[i] {
			b.WriteString("    ")
		} else {
			b.WriteString("│   ")
		}
	}
	if depth-1 < len(lastStack) && lastStack[depth-1] {
		b.WriteString("└── ")
	} else {
		b.WriteString("├── ")
	}
	return b.String()
}

// FlattenLsblkHierarchy walks the rows pre-order, decorating each row
// with depth + tree prefix + resolved parent. Mirrors
// `Overview2/Disk/config.ts:342`. When `rootName` isn't in the row set
// we degrade to a flat list (no tree), matching the SPA.
func FlattenLsblkHierarchy(rows []LsblkRow, rootName string) []LsblkFlatRow {
	nameSet := map[string]bool{}
	byName := map[string]LsblkRow{}
	for _, r := range rows {
		nameSet[r.Name] = true
		byName[r.Name] = r
	}
	if !nameSet[rootName] {
		sorted := append([]LsblkRow(nil), rows...)
		sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].Name < sorted[j].Name })
		out := make([]LsblkFlatRow, len(sorted))
		for i, r := range sorted {
			out[i] = LsblkFlatRow{Row: r, Depth: 0, TreePrefix: ""}
		}
		return out
	}
	children := map[string][]LsblkRow{}
	parents := map[string]string{}
	for _, r := range rows {
		if r.Name == rootName {
			continue
		}
		p := ResolveParent(r, rootName, nameSet)
		if p == "" || !nameSet[p] {
			p = rootName
		}
		parents[r.Name] = p
		children[p] = append(children[p], r)
	}
	for k := range children {
		c := children[k]
		sort.SliceStable(c, func(i, j int) bool { return c[i].Name < c[j].Name })
		children[k] = c
	}

	var out []LsblkFlatRow
	var walk func(rname string, depth int, lastStack []bool)
	walk = func(rname string, depth int, lastStack []bool) {
		r, ok := byName[rname]
		if !ok {
			return
		}
		prefix := ""
		if depth > 0 {
			prefix = BuildLsblkTreePrefix(depth, lastStack)
		}
		out = append(out, LsblkFlatRow{
			Row:        r,
			Parent:     parents[rname],
			Depth:      depth,
			TreePrefix: prefix,
		})
		ch := children[rname]
		for idx, c := range ch {
			isLast := idx == len(ch)-1
			next := append(append([]bool(nil), lastStack...), isLast)
			walk(c.Name, depth+1, next)
		}
	}
	walk(rootName, 0, nil)
	return out
}

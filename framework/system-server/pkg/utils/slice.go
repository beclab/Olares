package utils

import "sort"

// UniqAndSort returns a new slice of strings with duplicate entries removed and the remaining elements sorted.
func UniqAndSort(s []string) []string {
	u := make([]string, 0, len(s))
	m := make(map[string]bool)

	for _, val := range s {
		if _, ok := m[val]; !ok {
			m[val] = true
			u = append(u, val)
		}
	}

	ret := sort.StringSlice(u)
	ret.Sort()

	return ret
}

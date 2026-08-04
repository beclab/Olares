package picker

import "strings"

// matchScore reports whether query matches haystack and, if so, a relevance
// score (higher = better) used to rank results. The matching model is designed
// for "I only half-remember the name" without being noisy:
//
//   - The query is split on whitespace into tokens; EVERY token must appear as
//     a case-insensitive SUBSTRING (AND), but their ORDER is irrelevant —
//     "app olares" and "olares app" both hit ".../olares-app-deployment/olares-app".
//   - Substring (not subsequence): a token's characters must be contiguous, so
//     "ollama" does NOT spuriously match "otel-opentelemetry-operator" by
//     scattering o/l/l/a/m/a across it. To narrow on a half-remembered name,
//     type the fragments you recall as separate words (e.g. "ollama gpu").
//
// haystack is expected to already be lowercase (Entry.haystack()); the query is
// lowercased here. An all-whitespace query matches everything with score 0.
func matchScore(haystack, query string) (int, bool) {
	total := 0
	for _, tok := range strings.Fields(strings.ToLower(query)) {
		s, ok := scoreToken(haystack, tok)
		if !ok {
			return 0, false
		}
		total += s
	}
	return total, true
}

const (
	substringBase = 100 // base for any substring hit
	boundaryBonus = 24  // token landing right after a separator (word start)
	earlyMaxBonus = 20  // reward matches nearer the start of the haystack
)

// scoreToken scores a single token as a case-insensitive substring of haystack
// (both lowercase, ASCII — k8s names are ASCII). Returns (score, matched).
// Longer tokens, word-boundary hits, and earlier positions score higher.
func scoreToken(h, t string) (int, bool) {
	if t == "" {
		return 0, true
	}
	idx := strings.Index(h, t)
	if idx < 0 {
		return 0, false
	}
	score := substringBase + len(t)*2
	if idx == 0 || isBoundary(h[idx-1]) {
		score += boundaryBonus
	}
	if b := earlyMaxBonus - idx; b > 0 {
		score += b
	}
	return score, true
}

// isBoundary reports whether c is a separator that starts a new "word" in a
// namespace/pod/container path, so matches right after it rank higher.
func isBoundary(c byte) bool {
	switch c {
	case '/', '-', '_', '.', ':', ' ':
		return true
	}
	return false
}

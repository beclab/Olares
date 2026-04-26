package settings

// Phase 0a deliberately keeps this file thin: the umbrella ships with no
// real verbs yet, so there are no shared printers / helpers to host. Phase 1
// will populate it as common helpers emerge across read verbs (table
// renderers, error normalizers, etc.) — until then the package compiles and
// links against options.go + client.go alone.
//
// We keep the file in tree (rather than waiting until Phase 1 to add it) so
// the package layout matches market/ on disk; this keeps subsequent diffs
// focused on per-verb additions instead of churning the package skeleton.

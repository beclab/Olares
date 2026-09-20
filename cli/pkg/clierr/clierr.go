// Package clierr holds error sentinels that the process entrypoint must
// understand. It deliberately depends on nothing else so any subtree can
// import it without pulling in a domain package.
package clierr

import "errors"

// ErrAlreadyReported means the command has already written a
// user-visible diagnostic — a structured envelope, an operation result,
// or a typo hint — and the entrypoint must exit non-zero without
// printing anything further.
//
// cmd/main.go is the only place a returned error is printed, so a
// sentinel that reaches it unrecognised is rendered as its own message:
// agents saw a healthy but still-progressing install end with the line
// "Error: (already reported)". Subtrees that own such a sentinel alias
// or wrap this one so errors.Is holds at the entrypoint.
var ErrAlreadyReported = errors.New("already reported")

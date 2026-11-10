// Package handlers hosts one HandlerFunc per state in the application
// lifecycle state machine. Each file corresponds to one state (e.g.
// pending.go, downloading.go, installing.go) and exports pure functions
// of (context.Context, domain.World) domain.Result.
//
// Handlers in this package are subject to the FIVE hard rules listed in
// pkg/applifecycle/CONTRACT.md:
//
//  1. handlers MUST NOT mutate cluster state (no client.Update/Patch/
//     Create/Delete etc.); state writes are returned as PatchAM effects
//  2. handlers MUST NOT spawn goroutines (no `go func`); long-running
//     work is returned as CallWorkload effects
//  3. handlers MUST NOT block (no time.Sleep / wait.Poll*); waits are
//     returned as ScheduleTick effects
//  4. handlers' Result.Next MUST be declared in the TransitionTable
//     (enforced at runtime by Engine.Step's ErrUndeclaredTransition)
//  5. CallWorkload effects MUST specify explicit OnSuccess and OnFailure
//     Actions to close the loop back into the state machine
//
// Rules 1-3 are enforced at build time by the analyzer in
// pkg/applifecycle/lint/handlerpurity and by the forbidigo linter
// rules in .golangci.yml. Rule 4 is enforced at runtime by
// domain.Engine.Step. Rule 5 is enforced by code review (a future
// analyzer may add static coverage).
package handlers

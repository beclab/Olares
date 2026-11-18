// Package goroutine triggers CONTRACT.md rule 2: a handler must not
// spawn a goroutine. The analyzer must flag both the anonymous-func
// and named-func `go` statements below.
package goroutine

func work() {}

func Handle() {
	go work() // want "must not spawn goroutines"

	go func() { // want "must not spawn goroutines"
		_ = 1 + 1
	}()
}

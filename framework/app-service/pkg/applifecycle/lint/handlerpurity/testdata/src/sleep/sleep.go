// Package sleep triggers CONTRACT.md rule 3.1: handlers must not call
// time.Sleep. The analyzer must flag the call below while leaving the
// unrelated time.Now() call untouched.
package sleep

import "time"

func Handle() time.Time {
	time.Sleep(time.Second) // want "must not call time.Sleep"
	return time.Now()
}

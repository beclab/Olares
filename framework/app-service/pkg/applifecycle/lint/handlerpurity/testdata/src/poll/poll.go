// Package poll triggers CONTRACT.md rule 3.2: handlers must not call
// wait.Poll*. Uses the sibling stub "poll/wait" so the analyzer's
// pkg-suffix matcher (matchesPkgSuffix) can hit "wait" here just like
// it hits "k8s.io/apimachinery/pkg/util/wait" in production code.
package poll

import (
	"time"

	"poll/wait"
)

func alwaysTrue() (bool, error) { return true, nil }

func Handle() {
	_ = wait.Poll(time.Millisecond, time.Second, alwaysTrue)                  // want "must not call wait.Poll"
	_ = wait.PollImmediate(time.Millisecond, time.Second, alwaysTrue)         // want "must not call wait.PollImmediate"
	_ = wait.PollUntil(time.Millisecond, alwaysTrue, nil)                     // want "must not call wait.PollUntil"
	_ = wait.PollImmediateUntil(time.Millisecond, alwaysTrue, nil)            // want "must not call wait.PollImmediateUntil"
}

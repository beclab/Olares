package upgrade

import (
	"runtime"
	"testing"
)

func TestReleaseArchIsGOARCH(t *testing.T) {
	if got, want := releaseArch(), runtime.GOARCH; got != want {
		t.Fatalf("releaseArch() = %q, want %q (the old == \"arm\" check never matched arm64)", got, want)
	}
}

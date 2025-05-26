package util

import "testing"

func TestLocalMd5(t *testing.T) {
	sum := LocalMd5Sum("/tmp/main")
	t.Log("md5: ", sum)
}

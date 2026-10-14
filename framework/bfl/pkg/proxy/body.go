package proxy

import (
	"bytes"
	"io"
)

func ioNopCloser(body []byte) io.ReadCloser {
	return io.NopCloser(bytes.NewReader(body))
}

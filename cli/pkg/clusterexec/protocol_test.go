package clusterexec

import "testing"

func TestParseExitStatus(t *testing.T) {
	cases := []struct {
		name    string
		payload string
		want    int
	}{
		{"success", `{"status":"Success"}`, 0},
		{"nonzero", `{"status":"Failure","reason":"NonZeroExitCode","details":{"causes":[{"reason":"ExitCode","message":"127"}]}}`, 127},
		{"failure no cause", `{"status":"Failure","message":"boom"}`, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseExitStatus([]byte(tc.payload))
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if got != tc.want {
				t.Fatalf("got %d, want %d", got, tc.want)
			}
		})
	}
}

func TestSinkCaps(t *testing.T) {
	s := NewSink(4)
	s.Write(ChannelStdout, []byte("ab"))
	s.Write(ChannelStdout, []byte("cdef"))
	s.Write(ChannelStderr, []byte("xy"))
	if string(s.Stdout) != "abcd" {
		t.Fatalf("stdout = %q", s.Stdout)
	}
	if string(s.Stderr) != "xy" {
		t.Fatalf("stderr = %q", s.Stderr)
	}
	if !s.Truncated {
		t.Fatalf("expected Truncated=true")
	}
}

func TestResizeFrame(t *testing.T) {
	f, err := ResizeFrame(80, 24)
	if err != nil {
		t.Fatal(err)
	}
	if f[0] != ChannelResize {
		t.Fatalf("channel byte = %d", f[0])
	}
	if string(f[1:]) != `{"Width":80,"Height":24}` {
		t.Fatalf("payload = %q", f[1:])
	}
}

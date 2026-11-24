package terminal

import (
	"reflect"
	"testing"
)

func TestSelectHostUser(t *testing.T) {
	tests := []struct {
		name   string
		passwd string
		want   hostUser
	}{
		{
			name: "prefer olares",
			passwd: `root:x:0:0:root:/root:/bin/bash
ubuntu:x:1000:1000::/home/ubuntu:/bin/bash
olares:x:1001:1001::/home/olares:/bin/bash
`,
			want: hostUser{Name: "olares", UID: 1001, GID: 1001, Home: "/home/olares"},
		},
		{
			name: "smallest uid regardless of file order",
			passwd: `root:x:0:0:root:/root:/bin/bash
nobody:x:65534:65534:nobody:/nonexistent:/usr/sbin/nologin
bob:x:1001:1001::/home/bob:/bin/bash
alice:x:1000:1002::/home/alice:/bin/bash
`,
			want: hostUser{Name: "alice", UID: 1000, GID: 1002, Home: "/home/alice"},
		},
		{
			name: "only system users and nobody fall back to root",
			passwd: `root:x:0:0:root:/root:/bin/bash
daemon:x:1:1:daemon:/usr/sbin:/usr/sbin/nologin
nobody:x:65534:65534:nobody:/nonexistent:/usr/sbin/nologin
`,
			want: rootHostUser,
		},
		{
			name:   "empty passwd",
			passwd: "",
			want:   rootHostUser,
		},
		{
			name: "malformed lines are skipped",
			passwd: `# comment
broken-line
bad:x:abc:1000::/home/bad:/bin/bash
:x:1000:1000::/home/noname:/bin/bash
good:x:1005:1005::/home/good:/bin/bash
`,
			want: hostUser{Name: "good", UID: 1005, GID: 1005, Home: "/home/good"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := selectHostUser(tt.passwd); got != tt.want {
				t.Errorf("selectHostUser() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestBuildNodeShellCmd(t *testing.T) {
	got := buildNodeShellCmd(hostUser{Name: "olares", UID: 1000, GID: 1000, Home: "/home/olares"})
	want := []string{
		"nsenter", "-t", "1", "-m", "-u", "-i", "-n", "-p",
		"-S", "1000", "-G", "1000", "-w",
		"env", "HOME=/home/olares", "USER=olares",
		"/bin/bash",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("buildNodeShellCmd() = %v, want %v", got, want)
	}
}

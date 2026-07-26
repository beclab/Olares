package preinstall

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"strings"
	"testing"
)

func TestCopyVerifiedRegularFilePolicies(t *testing.T) {
	content := []byte("verified payload")
	sum := sha256.Sum256(content)
	digest := hex.EncodeToString(sum[:])
	sourceRoot, err := os.OpenRoot(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer sourceRoot.Close()
	if err := sourceRoot.WriteFile("source", content, 0o644); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name, target, digest, want string
		size, limit                int64
		mode                       os.FileMode
	}{
		{"bundle chart", "target", digest, "", int64(len(content)), 1024, 0o444},
		{"HF payload", "target", strings.ToUpper(digest), "", int64(len(content)), int64(len(content)), 0o644},
		{"size mismatch", "target", digest, "size mismatch", int64(len(content) + 1), 1024, 0o444},
		{"digest mismatch", "target", strings.Repeat("0", 64), "digest mismatch", int64(len(content)), 1024, 0o644},
		{"size limit", "target", digest, "exceeds", int64(len(content)), int64(len(content) - 1), 0o444},
		{"reserved marker", hfCacheMarkerFileName, "", "reserved", int64(len(content)), 1024, 0o644},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			targetRoot, err := os.OpenRoot(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			defer targetRoot.Close()

			_, err = copyVerifiedRegularFile(sourceRoot, targetRoot, verifiedCopy{
				Source:      "source",
				Target:      tt.target,
				Size:        tt.size,
				MaxSize:     tt.limit,
				SHA256:      tt.digest,
				OutputMode:  tt.mode,
				RejectLinks: true,
			})
			if tt.want != "" {
				if err == nil || !strings.Contains(err.Error(), tt.want) {
					t.Fatalf("copyVerifiedRegularFile() error = %v, want %q", err, tt.want)
				}
				return
			}
			if err != nil {
				t.Fatalf("copyVerifiedRegularFile() error = %v", err)
			}
			info, err := targetRoot.Stat("target")
			if err != nil {
				t.Fatal(err)
			}
			if info.Mode().Perm() != tt.mode {
				t.Fatalf("target mode = %o, want %o", info.Mode().Perm(), tt.mode)
			}
		})
	}
}

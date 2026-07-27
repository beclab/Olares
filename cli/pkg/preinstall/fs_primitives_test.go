package preinstall

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"strings"
	"testing"
)

func testRoot(t *testing.T) *os.Root {
	t.Helper()
	root, err := os.OpenRoot(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = root.Close() })
	return root
}

func TestCopyVerifiedRegularFilePolicies(t *testing.T) {
	content := []byte("verified payload")
	sum := sha256.Sum256(content)
	digest := hex.EncodeToString(sum[:])
	sourceRoot := testRoot(t)
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			targetRoot := testRoot(t)
			_, err := copyVerifiedRegularFile(sourceRoot, targetRoot, verifiedCopy{
				Source:     "source",
				Target:     tt.target,
				Size:       tt.size,
				MaxSize:    tt.limit,
				SHA256:     tt.digest,
				OutputMode: tt.mode,
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

func TestReservedHFMarkerPolicyIsCallerSpecific(t *testing.T) {
	sourceRoot := testRoot(t)
	if err := sourceRoot.WriteFile(hfCacheMarkerFileName, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Run("bundle allows legal filename", func(t *testing.T) {
		_, err := copyVerifiedRegularFile(sourceRoot, testRoot(t), verifiedCopy{
			Source: hfCacheMarkerFileName, Target: hfCacheMarkerFileName, Size: 1, MaxSize: 1, OutputMode: 0o444,
		})
		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("HF rejects completion marker", func(t *testing.T) {
		err := materializeHFEntry(sourceRoot, testRoot(t), ArtifactManifestEntryV1{
			Path: hfCacheMarkerFileName, Type: "file", Size: 1,
		})
		if err == nil || !strings.Contains(err.Error(), "reserved") {
			t.Fatalf("copyHFFile() error = %v", err)
		}
	})
}

func TestTrustedStagingInfoRejectsUnknownOwner(t *testing.T) {
	info, err := os.Stat(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if trustedStagingInfo(info, uint32(os.Geteuid()), []os.FileMode{info.Mode().Perm()}, func(os.FileInfo) (uint32, bool) {
		return 0, false
	}) {
		t.Fatal("unknown owner was trusted")
	}
}

func TestCopyVerifiedRegularFileRejectsSourceChanges(t *testing.T) {
	tests := []struct {
		name, want string
		hook       func(string, *verifiedCopy)
	}{
		{"replacement after lstat", "changed while opening", func(name string, spec *verifiedCopy) {
			spec.AfterLstat = func() error {
				if err := os.Rename(name, name+".old"); err != nil {
					return err
				}
				return os.WriteFile(name, []byte("x"), 0o644)
			}
		}},
		{"mutation after open", "changed while copying", func(name string, spec *verifiedCopy) {
			spec.BeforeCopy = func() error { return os.WriteFile(name, []byte("xx"), 0o644) }
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sourcePath := t.TempDir()
			sourceName := sourcePath + "/source"
			if err := os.WriteFile(sourceName, []byte("x"), 0o644); err != nil {
				t.Fatal(err)
			}
			source, err := os.OpenRoot(sourcePath)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = source.Close() })
			target := testRoot(t)
			spec := verifiedCopy{Source: "source", Target: "target", Size: 1, MaxSize: 1, OutputMode: 0o444}
			tt.hook(sourceName, &spec)
			if _, err := copyVerifiedRegularFile(source, target, spec); err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("copyVerifiedRegularFile() error = %v, want %q", err, tt.want)
			}
		})
	}
}

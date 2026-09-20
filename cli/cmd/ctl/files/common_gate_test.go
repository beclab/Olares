package files

import (
	"context"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// TestIsCommonFrontendPath pins the drive/Common recogniser the
// version gate keys on: only fileType=="drive" AND extend=="Common"
// (case-sensitive) counts. drive/Home, drive/Data and every other
// namespace are NOT Common.
func TestIsCommonFrontendPath(t *testing.T) {
	common := [][2]string{
		{"drive", "Common"},
	}
	notCommon := [][2]string{
		{"drive", "Home"},
		{"drive", "Data"},
		{"drive", "common"}, // case-sensitive: lowercase is not Common
		{"cache", "Common"}, // wrong fileType
		{"external", "node-1"},
		{"sync", "repo"},
		{"", ""},
	}
	for _, c := range common {
		if !isCommonFrontendPath(c[0], c[1]) {
			t.Errorf("isCommonFrontendPath(%q,%q) = false, want true", c[0], c[1])
		}
	}
	for _, c := range notCommon {
		if isCommonFrontendPath(c[0], c[1]) {
			t.Errorf("isCommonFrontendPath(%q,%q) = true, want false", c[0], c[1])
		}
	}
}

// TestRequireCommonBackendVersion_NoopWhenNotCommon confirms the gate
// short-circuits (no factory / network touch) when the operation does
// not involve drive/Common — exercised here with a nil Factory, which
// would panic if the function tried to resolve the backend version.
func TestRequireCommonBackendVersion_NoopWhenNotCommon(t *testing.T) {
	if err := requireCommonBackendVersion(context.Background(), nil, false); err != nil {
		t.Errorf("requireCommonBackendVersion(_, nil, false) = %v, want nil", err)
	}
}

func TestFilesVersionGates(t *testing.T) {
	previous := viper.GetString(cmdutil.FlagOlaresVersion)
	t.Cleanup(func() { viper.Set(cmdutil.FlagOlaresVersion, previous) })

	gates := []struct {
		name string
		run  func(context.Context, *cmdutil.Factory) error
		hint string
	}{
		{"common", func(ctx context.Context, f *cmdutil.Factory) error {
			return requireCommonBackendVersion(ctx, f, true)
		}, "drive/Home or drive/Data"},
		{"archive", requireArchiveBackendVersion, "upgrade the Olares system before using archive commands"},
		{"nfs", requireNFSBackendVersion, "upgrade the Olares system to use NFS mount commands"},
	}

	for _, gate := range gates {
		t.Run(gate.name+" rejects 1.12.5", func(t *testing.T) {
			viper.Set(cmdutil.FlagOlaresVersion, "1.12.5")
			err := gate.run(context.Background(), cmdutil.NewFactory())
			if err == nil {
				t.Fatal("expected version gate error")
			}
			for _, want := range []string{"Olares >= 1.12.6", "this backend is 1.12.5", gate.hint} {
				if !strings.Contains(err.Error(), want) {
					t.Fatalf("error %q missing %q", err, want)
				}
			}
		})

		t.Run(gate.name+" accepts 1.12.6 daily", func(t *testing.T) {
			viper.Set(cmdutil.FlagOlaresVersion, "1.12.6-20260808")
			if err := gate.run(context.Background(), cmdutil.NewFactory()); err != nil {
				t.Fatalf("daily build should pass: %v", err)
			}
		})

		t.Run(gate.name+" rejects undetectable version", func(t *testing.T) {
			viper.Set(cmdutil.FlagOlaresVersion, "not-a-version")
			err := gate.run(context.Background(), cmdutil.NewFactory())
			if err == nil || !strings.Contains(err.Error(), "version could not be determined") {
				t.Fatalf("expected fail-closed version error, got %v", err)
			}
		})
	}
}

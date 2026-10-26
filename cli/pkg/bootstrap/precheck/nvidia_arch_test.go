package precheck

import (
	"errors"
	"strings"
	"testing"

	"github.com/beclab/Olares/cli/pkg/core/connector"
)

type nvidiaArchRuntime struct {
	connector.Runtime
	runner *connector.Runner
}

func (r *nvidiaArchRuntime) GetRunner() *connector.Runner { return r.runner }
func (r *nvidiaArchRuntime) GetSystemInfo() connector.Systems {
	return &nvidiaArchSystem{}
}

type nvidiaArchSystem struct{ connector.Systems }

func (*nvidiaArchSystem) IsDarwin() bool { return false }

type nvidiaArchHost struct {
	connector.Host
	exec func(string) (string, error)
}

func (*nvidiaArchHost) SudoPrefixIfNecessary(cmd string) string { return cmd }
func (h *nvidiaArchHost) ExecExt(cmd string, _, _ bool) (string, int, error) {
	out, err := h.exec(cmd)
	if err != nil {
		return out, 1, err
	}
	return out, 0, nil
}

func TestNvidiaCardArchChecker(t *testing.T) {
	const (
		unknown   = "2f:00.0 3D controller: NVIDIA Corporation Device 2bb5 (rev a1)"
		blackwell = "2f:00.0 3D controller: NVIDIA Corporation GB202GL [RTX PRO 6000 Blackwell Server Edition] (rev a1)"
		pascal    = "01:00.0 VGA compatible controller: NVIDIA Corporation GP104 [GeForce GTX 1080] (rev a1)"
	)
	downloadErr := errors.New("download failed")
	tests := []struct {
		name           string
		scans          []string
		scanErr        error
		refreshErr     error
		wantScans      int
		wantRefreshes  int
		wantErr        string
		wantWrappedErr error
	}{
		{name: "no NVIDIA GPU", scans: []string{""}, wantScans: 1},
		{name: "known supported card needs no download", scans: []string{blackwell}, wantScans: 1},
		{name: "Ada VGA card", scans: []string{"01:00.0 VGA compatible controller: NVIDIA Corporation AD106 [GeForce RTX 4060 Ti]"}, wantScans: 1},
		{name: "Hopper 3D card", scans: []string{"01:00.0 3D controller: NVIDIA Corporation GH100 [H100 PCIe]"}, wantScans: 1},
		{name: "Ampere VGA card", scans: []string{"01:00.0 VGA compatible controller: NVIDIA Corporation GA102 [GeForce RTX 3090]"}, wantScans: 1},
		{name: "Turing VGA card", scans: []string{"01:00.0 VGA compatible controller: NVIDIA Corporation TU104GL [Quadro RTX 5000]"}, wantScans: 1},
		{name: "known unsupported card needs no download", scans: []string{pascal}, wantScans: 1, wantErr: "unsupported NVIDIA card GeForce GTX 1080 of architecture: Pascal"},
		{name: "stale database refresh identifies Blackwell", scans: []string{unknown, blackwell}, wantScans: 2, wantRefreshes: 1},
		{name: "unknown VGA card also refreshes", scans: []string{"ad:00.0 VGA compatible controller: NVIDIA Corporation Device 1ad0 (rev a1)", blackwell}, wantScans: 2, wantRefreshes: 1},
		{name: "unknown after one refresh stays blocked", scans: []string{unknown, unknown}, wantScans: 2, wantRefreshes: 1, wantErr: "cannot determine architecture of NVIDIA card " + unknown + " even after refreshing"},
		{name: "refresh identifies unsupported architecture", scans: []string{unknown, pascal}, wantScans: 2, wantRefreshes: 1, wantErr: "unsupported NVIDIA card GeForce GTX 1080 of architecture: Pascal"},
		{name: "download failure gives recovery instructions", scans: []string{unknown}, refreshErr: downloadErr, wantScans: 1, wantRefreshes: 1, wantErr: "ensure pciutils is installed, run `sudo update-pciids`, and retry", wantWrappedErr: downloadErr},
		{name: "missing updater stays blocked", scans: []string{unknown}, refreshErr: errors.New("update-pciids: No such file or directory"), wantScans: 1, wantRefreshes: 1, wantErr: "PCI ID database refresh failed"},
		{name: "refresh timeout stays blocked", scans: []string{unknown}, refreshErr: errors.New("exit status 124"), wantScans: 1, wantRefreshes: 1, wantErr: "PCI ID database refresh failed"},
		{name: "GPU disappearing on retry stays blocked", scans: []string{unknown, ""}, wantScans: 2, wantRefreshes: 1, wantErr: "NVIDIA card was no longer detected"},
		{name: "initial scan failure is preserved", scans: []string{""}, scanErr: errors.New("scan failed"), wantScans: 1, wantErr: "scan failed"},
		{name: "retry scan failure is preserved", scans: []string{unknown, ""}, scanErr: errors.New("scan failed"), wantScans: 2, wantRefreshes: 1, wantErr: "failed to detect NVIDIA card after refreshing"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var scans, refreshes int
			host := &nvidiaArchHost{exec: func(cmd string) (string, error) {
				if strings.HasPrefix(cmd, "lspci ") {
					if scans >= len(tt.scans) {
						t.Fatalf("unexpected extra GPU scan: %s", cmd)
					}
					out := tt.scans[scans]
					scans++
					if scans == len(tt.scans) {
						return out, tt.scanErr
					}
					return out, nil
				}
				// The timeout must cover the downloader's process group and
				// force termination if it ignores the initial signal.
				if cmd != "timeout --kill-after=5s 30s update-pciids" {
					t.Fatalf("unexpected command: %s", cmd)
				}
				if scans != 1 {
					t.Fatalf("refresh ran after %d scans, want 1", scans)
				}
				refreshes++
				return "", tt.refreshErr
			}}
			runtime := &nvidiaArchRuntime{runner: &connector.Runner{Host: host}}
			err := (&NvidiaCardArchChecker{}).Check(runtime)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("Check() = %v, want success", err)
				}
			} else if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Check() = %v, want error containing %q", err, tt.wantErr)
			}
			if tt.wantWrappedErr != nil && !errors.Is(err, tt.wantWrappedErr) {
				t.Errorf("Check() = %v, want wrapped %v", err, tt.wantWrappedErr)
			}
			if scans != tt.wantScans || refreshes != tt.wantRefreshes {
				t.Errorf("scans/refreshes = %d/%d, want %d/%d", scans, refreshes, tt.wantScans, tt.wantRefreshes)
			}
		})
	}
}

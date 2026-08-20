package devdeploy

import (
	"strings"
	"testing"

	"github.com/beclab/Olares/cli/pkg/cliconfig"
)

func TestImportArgsAlwaysTargetsTheKubeletNamespace(t *testing.T) {
	got := strings.Join(importArgs([]string{"k3s", "ctr"}), " ")
	want := "k3s ctr -n k8s.io images import -"
	if got != want {
		t.Errorf("importArgs = %q, want %q", got, want)
	}
	// An image imported into any other namespace is invisible to the
	// kubelet, so this is not a preference to be made configurable.
	if !strings.Contains(got, "-n "+ContainerdNamespace) {
		t.Errorf("import must target %s", ContainerdNamespace)
	}
}

func TestSelectRejectsUnknownTransport(t *testing.T) {
	_, err := Select("carrier-pigeon", Options{})
	if err == nil {
		t.Fatal("expected an error")
	}
	if !strings.Contains(err.Error(), "auto, local, ssh, registry, api") {
		t.Errorf("error %q should list the valid transports", err)
	}
}

// The unimplemented transports must say what to do instead rather than
// failing with a bare "not implemented".
func TestSelectExplainsUnimplementedTransports(t *testing.T) {
	for _, name := range []string{TransportRegistry, TransportAPI} {
		t.Run(name, func(t *testing.T) {
			_, err := Select(name, Options{})
			if err == nil {
				t.Fatal("expected an error")
			}
			if !strings.Contains(err.Error(), "not implemented yet") {
				t.Errorf("error %q should say it is not implemented", err)
			}
			if !strings.Contains(err.Error(), "set-image") {
				t.Errorf("error %q should point at the manual alternative", err)
			}
		})
	}
}

func TestSSHTransportRequiresAddressAndUser(t *testing.T) {
	tests := map[string]struct {
		node    *cliconfig.DevNodeConfig
		wantMsg string
	}{
		"no node at all": {
			node:    nil,
			wantMsg: "no node configured",
		},
		"blank address": {
			node:    &cliconfig.DevNodeConfig{User: "ronny"},
			wantMsg: "no node configured",
		},
		"missing user": {
			node:    &cliconfig.DevNodeConfig{Address: "olares.lan"},
			wantMsg: "no SSH user",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got, why := newSSHTransport(Options{Node: tc.node})
			if got != nil {
				t.Fatal("expected the ssh transport to be unavailable")
			}
			if !strings.Contains(why, tc.wantMsg) {
				t.Errorf("reason %q does not mention %q", why, tc.wantMsg)
			}
		})
	}
}

// The npm/npx build never runs on the node, so offering to import into
// "this machine's" containerd there would always be wrong — including
// on a developer laptop that happens to run Docker.
func TestLocalTransportRefusedUnderRemoteOnly(t *testing.T) {
	t.Setenv("OLARES_CLI_REMOTE_ONLY", "1")
	got, why := newLocalTransport(Options{})
	if got != nil {
		t.Fatal("expected the local transport to be unavailable under OLARES_CLI_REMOTE_ONLY")
	}
	if !strings.Contains(why, "npm/npx build") {
		t.Errorf("reason %q should explain that this build never runs on the node", why)
	}
}

func TestSelectAutoUnderRemoteOnlyExplainsBothLadderRungs(t *testing.T) {
	t.Setenv("OLARES_CLI_REMOTE_ONLY", "1")
	_, err := Select(TransportAuto, Options{})
	if err == nil {
		t.Fatal("expected an error when neither transport is available")
	}
	for _, want := range []string{"local:", "ssh:", "dev node set"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q should contain %q", err, want)
		}
	}
}

func TestResolveImageToolRejectsMissingOverride(t *testing.T) {
	_, err := resolveImageTool("definitely-not-a-real-binary")
	if err == nil {
		t.Fatal("expected an error for a missing tool")
	}
	if !strings.Contains(err.Error(), "not found on PATH") {
		t.Errorf("error %q should say the tool is missing", err)
	}
}

// hasKubeletMarker is what stops `auto` from picking a developer
// machine's own Docker containerd. Without a marker the local transport
// must decline even though a socket exists.
func TestHasKubeletMarkerIsRequiredForAmbiguousSockets(t *testing.T) {
	for _, marker := range kubeletMarkers {
		if !strings.HasPrefix(marker, "/") {
			t.Errorf("kubelet marker %q should be an absolute path", marker)
		}
	}
	if len(kubeletMarkers) == 0 {
		t.Fatal("there must be at least one kubelet marker or every Docker host looks like a node")
	}
	// The k3s socket is self-proving; the generic ones are not.
	for _, sock := range containerdSockets {
		if strings.Contains(sock.path, "k3s") != sock.k3sOwned {
			t.Errorf("socket %q: k3sOwned=%v does not match its path", sock.path, sock.k3sOwned)
		}
	}
}

package network

import (
	"errors"
	"strings"
	"testing"
	"time"
)

// scriptedRunner answers commands by prefix and records everything it was asked.
type scriptedRunner struct {
	calls   []string
	answers []scriptedAnswer
}

type scriptedAnswer struct {
	prefix string
	out    string
	err    error
}

func (r *scriptedRunner) run(cmd string) (string, error) {
	r.calls = append(r.calls, cmd)
	for _, a := range r.answers {
		if strings.HasPrefix(cmd, a.prefix) {
			return a.out, a.err
		}
	}
	return "", nil
}

func (r *scriptedRunner) indexOf(prefix string) int {
	for i, c := range r.calls {
		if strings.HasPrefix(c, prefix) {
			return i
		}
	}
	return -1
}

func (r *scriptedRunner) indexOfExact(cmd string) int {
	for i, c := range r.calls {
		if c == cmd {
			return i
		}
	}
	return -1
}

func noSleep(time.Duration) {}

func TestSelectWiredInterfaceSkipsWifiAndBridgeSlaves(t *testing.T) {
	devs := parseNMDevices(strings.Join([]string{
		"wlo1:wifi:connected:home-wifi",
		"enp3s0:ethernet:connected:br-olares-slave-enp3s0",
		"enp4s0:ethernet:connected (externally):Wired connection 2",
		"docker0:bridge:connected (externally):docker0",
		"lo:loopback:connected (externally):lo",
	}, "\n"))
	dev, ok := selectWiredInterface(devs)
	if !ok || dev != "enp4s0" {
		t.Fatalf("selectWiredInterface = %q,%v; want enp4s0", dev, ok)
	}
}

func TestSelectWiredInterfaceNoneWhenOnlyWifi(t *testing.T) {
	if dev, ok := selectWiredInterface(parseNMDevices("wlo1:wifi:connected:x\nenp3s0:ethernet:disconnected:")); ok {
		t.Fatalf("expected no wired interface, got %q", dev)
	}
}

func TestOverlayLinkFileContent(t *testing.T) {
	got := overlayLinkFileContent("d8:43:ae:af:5a:33")
	for _, want := range []string{"[Match]", "MACAddress=d8:43:ae:af:5a:33", "Type=ether", "[Link]", "AlternativeName=olares-lan"} {
		if !strings.Contains(got, want) {
			t.Fatalf("link file %q lacks %q", got, want)
		}
	}
	if !strings.Contains(overlayLinkFileCommand("d8:43:ae:af:5a:33"), OverlayLinkFile) {
		t.Fatal("link file command must write the persisted unit path")
	}
}

func TestLinkNameFromIPOutput(t *testing.T) {
	cases := map[string]string{
		"2: enp3s0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500":        "enp3s0",
		"830: br-olares: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500\n": "br-olares",
		"5: net1@if830: <BROADCAST> mtu 1500":                          "net1",
		"":                                                             "",
	}
	for in, want := range cases {
		if got := linkNameFromIPOutput(in); got != want {
			t.Errorf("linkNameFromIPOutput(%q) = %q, want %q", in, got, want)
		}
	}
}

func activeBridgeAnswers(extra ...scriptedAnswer) []scriptedAnswer {
	base := []scriptedAnswer{
		{prefix: "nmcli -g connection.type connection show br-olares", out: "bridge"},
		{prefix: "nmcli -g GENERAL.STATE connection show br-olares", out: "activated"},
		{prefix: "nmcli -g NAME connection show", out: "br-olares\nbr-olares-slave-enp3s0\noriginal-connection\n"},
		{prefix: "ip -4 -o addr show dev enp3s0", out: "2: enp3s0    inet 192.168.50.164/24 brd 192.168.50.255 scope global"},
		{prefix: "ip -4 route show default dev enp3s0", out: "default via 192.168.50.1 proto dhcp"},
	}
	return append(extra, base...)
}

func TestMigrateBridgeToDirectActiveBridge(t *testing.T) {
	r := &scriptedRunner{answers: activeBridgeAnswers()}
	migrated, err := migrateBridgeToDirect(r.run, noSleep)
	if err != nil || !migrated {
		t.Fatalf("migrate = %v,%v; want true,nil (calls %v)", migrated, err, r.calls)
	}
	marker, down, up, delSlave, delBridge := r.indexOf("mkdir -p /var/lib/olares/overlay-gateway && touch /var/lib/olares/overlay-gateway/.migrated-from-bridge"),
		r.indexOf("nmcli connection down br-olares"), r.indexOf("nmcli connection up original-connection"),
		r.indexOf("nmcli connection delete br-olares-slave-enp3s0"), r.indexOfExact("nmcli connection delete br-olares")
	for name, at := range map[string]int{"marker": marker, "down": down, "up": up, "delete slave": delSlave, "delete bridge": delBridge} {
		if at < 0 {
			t.Fatalf("missing step %s in %v", name, r.calls)
		}
	}
	if !(marker < down && down < up && up < delSlave && delSlave < delBridge) {
		t.Fatalf("order must be marker < down bridge < up original < delete slave < delete bridge, got %v", r.calls)
	}
	if r.indexOf("nmcli connection add type ethernet") >= 0 {
		t.Fatal("must not create a physical profile when original-connection exists")
	}
}

func TestMigrateBridgeToDirectCreatesProfileWhenOriginalMissing(t *testing.T) {
	r := &scriptedRunner{answers: activeBridgeAnswers(scriptedAnswer{
		prefix: "nmcli -g NAME connection show", out: "br-olares\nbr-olares-slave-enp3s0\n",
	})}
	if migrated, err := migrateBridgeToDirect(r.run, noSleep); err != nil || !migrated {
		t.Fatalf("migrate = %v,%v; want true,nil", migrated, err)
	}
	add := r.indexOf("nmcli connection add type ethernet ifname enp3s0 con-name original-connection")
	if add < 0 || add > r.indexOf("nmcli connection down br-olares") {
		t.Fatalf("physical profile must be created before the bridge goes down, got %v", r.calls)
	}
}

func TestMigrateBridgeToDirectRestoresBridgeWhenPhyStaysDown(t *testing.T) {
	r := &scriptedRunner{answers: activeBridgeAnswers(scriptedAnswer{prefix: "ip -4 -o addr show dev enp3s0", out: ""})}
	migrated, err := migrateBridgeToDirect(r.run, noSleep)
	if err == nil || migrated {
		t.Fatalf("expected failure without migration, got %v,%v", migrated, err)
	}
	restoreDown, restoreUp := r.indexOf("nmcli connection down original-connection"), r.indexOf("nmcli connection up br-olares")
	if restoreDown < 0 || restoreUp < 0 || restoreDown > restoreUp {
		t.Fatalf("bridge must be restored (down original, up bridge), got %v", r.calls)
	}
	if r.indexOfExact("nmcli connection delete br-olares") >= 0 || r.indexOf("nmcli connection delete br-olares-slave") >= 0 {
		t.Fatal("bridge profiles must survive a failed migration")
	}
}

func TestMigrateBridgeToDirectInactiveBridgeOnlyDeletesProfiles(t *testing.T) {
	r := &scriptedRunner{answers: []scriptedAnswer{
		{prefix: "nmcli -g connection.type connection show br-olares", out: "bridge"},
		{prefix: "nmcli -g GENERAL.STATE connection show br-olares", out: ""},
		{prefix: "nmcli -g NAME connection show", out: "br-olares\nbr-olares-slave-enp3s0\noriginal-connection\n"},
	}}
	migrated, err := migrateBridgeToDirect(r.run, noSleep)
	if err != nil || migrated {
		t.Fatalf("migrate = %v,%v; want false,nil", migrated, err)
	}
	if r.indexOf("nmcli connection down") >= 0 || r.indexOf("nmcli connection up") >= 0 {
		t.Fatalf("an inactive bridge must not touch the active physical connection, got %v", r.calls)
	}
	if r.indexOf("nmcli connection delete br-olares-slave-enp3s0") < 0 || r.indexOfExact("nmcli connection delete br-olares") < 0 {
		t.Fatalf("stale profiles must be deleted, got %v", r.calls)
	}
}

func TestMigrateBridgeToDirectNoBridge(t *testing.T) {
	r := &scriptedRunner{answers: []scriptedAnswer{
		{prefix: "nmcli -g connection.type connection show br-olares", err: errors.New("no such connection profile")},
	}}
	migrated, err := migrateBridgeToDirect(r.run, noSleep)
	if err != nil || migrated {
		t.Fatalf("migrate = %v,%v; want false,nil", migrated, err)
	}
	if len(r.calls) != 1 {
		t.Fatalf("no further commands expected without a bridge, got %v", r.calls)
	}
}

func TestEnsureOverlayAltnameAddsNameAndLinkFile(t *testing.T) {
	r := &scriptedRunner{answers: []scriptedAnswer{
		{prefix: "nmcli -t -f DEVICE,TYPE,STATE,CONNECTION device", out: "enp3s0:ethernet:connected:Wired connection 1\nwlo1:wifi:connected:home\n"},
		{prefix: "ip -o link show olares-lan", err: errors.New("does not exist")},
		{prefix: "cat /sys/class/net/enp3s0/address", out: "d8:43:ae:af:5a:33\n"},
	}}
	dev, err := ensureOverlayAltname(r.run)
	if err != nil || dev != "enp3s0" {
		t.Fatalf("ensureOverlayAltname = %q,%v", dev, err)
	}
	if r.indexOf("ip link property add dev enp3s0 altname olares-lan") < 0 {
		t.Fatalf("alternative name must be added, got %v", r.calls)
	}
	if r.indexOf("mkdir -p /etc/systemd/network && printf") < 0 {
		t.Fatalf("link file must be written, got %v", r.calls)
	}
}

func TestEnsureOverlayAltnameRefusesToRebind(t *testing.T) {
	r := &scriptedRunner{answers: []scriptedAnswer{
		{prefix: "nmcli -t -f DEVICE,TYPE,STATE,CONNECTION device", out: "enp3s0:ethernet:connected:Wired connection 1\n"},
		{prefix: "ip -o link show olares-lan", out: "7: enp2s0: <BROADCAST> mtu 1500"},
	}}
	if _, err := ensureOverlayAltname(r.run); err == nil {
		t.Fatal("expected an error when olares-lan already names another device")
	}
	if r.indexOf("ip link property add") >= 0 {
		t.Fatal("must not add the name to a second device")
	}
}

func TestEnsureOverlayAltnameSkipsWithoutWiredInterface(t *testing.T) {
	r := &scriptedRunner{answers: []scriptedAnswer{
		{prefix: "nmcli -t -f DEVICE,TYPE,STATE,CONNECTION device", out: "wlo1:wifi:connected:home\n"},
	}}
	dev, err := ensureOverlayAltname(r.run)
	if err != nil || dev != "" {
		t.Fatalf("expected a silent skip, got %q,%v", dev, err)
	}
	if len(r.calls) != 1 {
		t.Fatalf("no changes expected without a wired interface, got %v", r.calls)
	}
}

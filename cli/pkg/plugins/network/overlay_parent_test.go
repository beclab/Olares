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

func TestOverlayUdevRuleContentAndCommand(t *testing.T) {
	rule := overlayUdevRuleContent("d8:43:ae:af:5a:33", "/bin/ip")
	for _, want := range []string{`ACTION=="add"`, `SUBSYSTEM=="net"`, `ATTR{address}=="d8:43:ae:af:5a:33"`, `RUN+="/bin/ip link property add dev $env{INTERFACE} altname olares-lan"`} {
		if !strings.Contains(rule, want) {
			t.Fatalf("udev rule %q lacks %q", rule, want)
		}
	}
	if strings.Contains(rule, "'") {
		t.Fatalf("udev rule must not contain single quotes, it is written through a single-quoted printf: %q", rule)
	}
	cmd := overlayUdevRuleCommand("d8:43:ae:af:5a:33", "/bin/ip")
	for _, want := range []string{"mkdir -p /etc/udev/rules.d", "> " + OverlayUdevRuleFile, "rm -f " + legacyOverlayLinkFile, rule} {
		if !strings.Contains(cmd, want) {
			t.Fatalf("rule command %q lacks %q", cmd, want)
		}
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
	marker, down, up, delSlave, delBridge := r.indexOfExact("mkdir -p /var/lib/olares/overlay-gateway && printf '%s' 'enp3s0' > /var/lib/olares/overlay-gateway/.migrated-from-bridge"),
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
	if rm := r.indexOfExact("rm -f /var/lib/olares/overlay-gateway/.migrated-from-bridge"); rm < 0 || rm < restoreUp {
		t.Fatalf("migration marker must be removed after the bridge is restored, got %v", r.calls)
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

func ensureAnswers(extra ...scriptedAnswer) []scriptedAnswer {
	return append(extra, []scriptedAnswer{
		{prefix: "ip -o link show olares-lan", err: errors.New("does not exist")},
		{prefix: "cat /sys/class/net/enp3s0/address", out: "d8:43:ae:af:5a:33\n"},
		{prefix: "command -v ip", out: "/bin/ip\n"},
	}...)
}

func TestEnsureOverlayAltnameAddsNameAndUdevRule(t *testing.T) {
	r := &scriptedRunner{answers: ensureAnswers()}
	if err := ensureOverlayAltname(r.run, "enp3s0"); err != nil {
		t.Fatalf("ensureOverlayAltname: %v", err)
	}
	add := r.indexOfExact("ip link property add dev enp3s0 altname olares-lan")
	if add < 0 {
		t.Fatalf("alternative name must be added, got %v", r.calls)
	}
	rule := r.indexOf("mkdir -p /etc/udev/rules.d && printf")
	if rule < 0 || rule < add {
		t.Fatalf("udev rule must be written after the name is added, got %v", r.calls)
	}
	for _, want := range []string{"/bin/ip link property add dev $env{INTERFACE} altname olares-lan", "> /etc/udev/rules.d/80-olares-lan.rules", "rm -f /etc/systemd/network/10-olares-lan.link"} {
		if !strings.Contains(r.calls[rule], want) {
			t.Fatalf("rule command %q lacks %q", r.calls[rule], want)
		}
	}
	if r.indexOf("nmcli") >= 0 {
		t.Fatalf("the upgrade must not select an interface itself, got %v", r.calls)
	}
}

func TestEnsureOverlayAltnameMovesLeftoverNameToThePhy(t *testing.T) {
	r := &scriptedRunner{answers: ensureAnswers(scriptedAnswer{prefix: "ip -o link show olares-lan", out: "7: enp2s0: <BROADCAST> mtu 1500"})}
	if err := ensureOverlayAltname(r.run, "enp3s0"); err != nil {
		t.Fatalf("a leftover binding must be moved, not fail the upgrade: %v", err)
	}
	del, add := r.indexOfExact("ip link property del dev enp2s0 altname olares-lan"), r.indexOfExact("ip link property add dev enp3s0 altname olares-lan")
	if del < 0 || add < 0 || del > add {
		t.Fatalf("name must be removed from enp2s0 before being added to enp3s0, got %v", r.calls)
	}
}

func TestEnsureOverlayAltnameIdempotentWhenAlreadyOnPhy(t *testing.T) {
	r := &scriptedRunner{answers: ensureAnswers(scriptedAnswer{prefix: "ip -o link show olares-lan", out: "2: enp3s0: <BROADCAST> mtu 1500"})}
	if err := ensureOverlayAltname(r.run, "enp3s0"); err != nil {
		t.Fatalf("ensureOverlayAltname: %v", err)
	}
	if r.indexOf("ip link property add") >= 0 || r.indexOf("ip link property del") >= 0 {
		t.Fatalf("name already on the phy must not be touched, got %v", r.calls)
	}
	if r.indexOf("mkdir -p /etc/udev/rules.d && printf") < 0 {
		t.Fatalf("udev rule must still be refreshed, got %v", r.calls)
	}
}

func TestEnsureOverlayAltnameFailsWithoutIPCommand(t *testing.T) {
	r := &scriptedRunner{answers: ensureAnswers(scriptedAnswer{prefix: "command -v ip", err: errors.New("exit 1")})}
	if err := ensureOverlayAltname(r.run, "enp3s0"); err == nil {
		t.Fatal("a rule with an unknown ip path must not be written")
	}
	if r.indexOf("mkdir -p /etc/udev/rules.d") >= 0 {
		t.Fatalf("rule must not be written without the ip path, got %v", r.calls)
	}
}

func TestOverlayMigratedFromBridgeReadsThePhy(t *testing.T) {
	r := &scriptedRunner{answers: []scriptedAnswer{{prefix: "cat /var/lib/olares/overlay-gateway/.migrated-from-bridge", out: "enp3s0\n"}}}
	if phy, ok := overlayMigratedFromBridge(r.run); !ok || phy != "enp3s0" {
		t.Fatalf("overlayMigratedFromBridge = %q,%v; want enp3s0,true", phy, ok)
	}
	r = &scriptedRunner{answers: []scriptedAnswer{{prefix: "cat /var/lib/olares/overlay-gateway/.migrated-from-bridge", err: errors.New("No such file")}}}
	if phy, ok := overlayMigratedFromBridge(r.run); ok || phy != "" {
		t.Fatalf("missing marker must read as not migrated, got %q,%v", phy, ok)
	}
}

func TestRemoveOverlayAltnameDropsRuleAndLegacyLinkFile(t *testing.T) {
	r := &scriptedRunner{answers: []scriptedAnswer{{prefix: "ip -o link show olares-lan", out: "2: enp3s0: <BROADCAST> mtu 1500"}}}
	removeOverlayAltname(r.run)
	if r.indexOfExact("rm -f /etc/udev/rules.d/80-olares-lan.rules /etc/systemd/network/10-olares-lan.link") < 0 {
		t.Fatalf("both persistence files must be removed, got %v", r.calls)
	}
	if r.indexOfExact("ip link property del dev enp3s0 altname olares-lan") < 0 {
		t.Fatalf("alternative name must be dropped, got %v", r.calls)
	}
}

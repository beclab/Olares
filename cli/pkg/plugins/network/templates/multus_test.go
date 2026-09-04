package templates

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// renderedUnderlayConfig executes MultusDefine and returns the parsed
// NetworkAttachmentDefinition spec.config JSON object.
func renderedUnderlayConfig(t *testing.T) map[string]interface{} {
	t.Helper()
	var buf bytes.Buffer
	if err := MultusDefine.Execute(&buf, nil); err != nil {
		t.Fatalf("render MultusDefine: %v", err)
	}
	lines := strings.Split(buf.String(), "\n")
	var block []string
	in := false
	for _, l := range lines {
		if strings.TrimSpace(l) == "config: |-" {
			in = true
			continue
		}
		if in {
			if strings.TrimSpace(l) == "" && len(block) > 0 {
				break
			}
			block = append(block, strings.TrimPrefix(l, "    "))
		}
	}
	if len(block) == 0 {
		t.Fatalf("no config block in rendered template:\n%s", buf.String())
	}
	var conf map[string]interface{}
	if err := json.Unmarshal([]byte(strings.Join(block, "\n")), &conf); err != nil {
		t.Fatalf("spec.config is not valid JSON: %v\n%s", err, strings.Join(block, "\n"))
	}
	return conf
}

// TestUnderlayMacvlanDefineDisablesDHCPRelease pins the NAD-level RELEASE
// policy: a recreated Overlay Pod must get its LAN address back, which
// requires the DHCP binding to stay alive on the router. Older dhcp binaries
// ignore the unknown field, so shipping the template ahead of the binary is
// safe (see OG-DES-002 section 4.5).
func TestUnderlayMacvlanDefineDisablesDHCPRelease(t *testing.T) {
	conf := renderedUnderlayConfig(t)
	ipam, ok := conf["ipam"].(map[string]interface{})
	if !ok {
		t.Fatalf("ipam block missing: %v", conf)
	}
	v, present := ipam["sendRelease"]
	if !present {
		t.Fatal("ipam.sendRelease must be written explicitly (an absent field means the upstream default: send RELEASE)")
	}
	if b, isBool := v.(bool); !isBool || b {
		t.Fatalf("ipam.sendRelease = %v, want explicit false", v)
	}
	if ipam["type"] != "dhcp" || ipam["omitDefaultGateway"] != true {
		t.Fatalf("ipam type/omitDefaultGateway changed: %v", ipam)
	}
	if _, has := ipam["request"]; !has {
		t.Fatal("ipam.request (PRL) must be preserved")
	}
}

// TestUnderlayMacvlanDefineStructureInvariants pins the shape the dhcp daemon
// and Multus depend on: a single plain plugin conf (not a conflist), cniVersion
// 0.3.1 and the network name "underlay".
func TestUnderlayMacvlanDefineStructureInvariants(t *testing.T) {
	conf := renderedUnderlayConfig(t)
	if _, isList := conf["plugins"]; isList {
		t.Fatal("underlay-macvlan must stay a single plugin conf, not a conflist")
	}
	if conf["cniVersion"] != "0.3.1" {
		t.Fatalf("cniVersion = %v, want 0.3.1 (bumping it is a separate decision)", conf["cniVersion"])
	}
	if conf["name"] != "underlay" || conf["type"] != "macvlan" || conf["master"] != "br-olares" || conf["mode"] != "bridge" {
		t.Fatalf("top-level fields changed: %v", conf)
	}
}

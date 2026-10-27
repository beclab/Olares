package templates

import (
	"bytes"
	"encoding/json"
	"testing"

	"sigs.k8s.io/yaml"
)

// renderedUnderlayDefine executes MultusDefine and decodes the rendered
// NetworkAttachmentDefinition, returning its metadata and the parsed
// spec.config JSON object.
func renderedUnderlayDefine(t *testing.T) (meta map[string]string, conf map[string]interface{}) {
	t.Helper()
	var buf bytes.Buffer
	if err := MultusDefine.Execute(&buf, nil); err != nil {
		t.Fatalf("render MultusDefine: %v", err)
	}
	var nad struct {
		APIVersion string            `json:"apiVersion"`
		Kind       string            `json:"kind"`
		Metadata   map[string]string `json:"metadata"`
		Spec       struct {
			Config string `json:"config"`
		} `json:"spec"`
	}
	if err := yaml.Unmarshal(buf.Bytes(), &nad); err != nil {
		t.Fatalf("rendered template is not valid YAML: %v\n%s", err, buf.String())
	}
	if nad.Kind != "NetworkAttachmentDefinition" || nad.APIVersion != "k8s.cni.cncf.io/v1" {
		t.Fatalf("unexpected kind/apiVersion: %s/%s", nad.Kind, nad.APIVersion)
	}
	if err := json.Unmarshal([]byte(nad.Spec.Config), &conf); err != nil {
		t.Fatalf("spec.config is not valid JSON: %v\n%s", err, nad.Spec.Config)
	}
	return nad.Metadata, conf
}

// TestUnderlayMacvlanDefineDisablesDHCPRelease pins the NAD-level RELEASE
// policy: a recreated Overlay Pod must get its LAN address back, which
// requires the DHCP binding to stay alive on the router. Older dhcp binaries
// ignore the unknown field, so shipping the template ahead of the binary is
// safe.
func TestUnderlayMacvlanDefineDisablesDHCPRelease(t *testing.T) {
	_, conf := renderedUnderlayDefine(t)
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
// and Multus depend on: object name/namespace, a single plain plugin conf (not
// a conflist), cniVersion 0.3.1, the network name "underlay" and the master
// "olares-lan", the alternative name the node adds to its wired NIC.
func TestUnderlayMacvlanDefineStructureInvariants(t *testing.T) {
	meta, conf := renderedUnderlayDefine(t)
	if meta["name"] != "underlay-macvlan" || meta["namespace"] != "kube-system" {
		t.Fatalf("metadata changed: %v", meta)
	}
	if _, isList := conf["plugins"]; isList {
		t.Fatal("underlay-macvlan must stay a single plugin conf, not a conflist")
	}
	if conf["cniVersion"] != "0.3.1" {
		t.Fatalf("cniVersion = %v, want 0.3.1 (bumping it is a separate decision)", conf["cniVersion"])
	}
	if conf["name"] != "underlay" || conf["type"] != "macvlan" || conf["master"] != "olares-lan" || conf["mode"] != "bridge" {
		t.Fatalf("top-level fields changed: %v", conf)
	}
}

// TestMultusMainContainerMemoryBudget pins the kube-multus container budget:
// a 50Mi request with a 256Mi limit. The install init container stays at its
// 15Mi request and has no limit.
func TestMultusMainContainerMemoryBudget(t *testing.T) {
	var buf bytes.Buffer
	if err := Multus.Execute(&buf, map[string]string{
		"MultusImage": "example.invalid/multus:test",
	}); err != nil {
		t.Fatalf("render Multus: %v", err)
	}

	var main, init map[string]interface{}
	for _, doc := range bytes.Split(buf.Bytes(), []byte("\n---")) {
		doc = bytes.TrimSpace(doc)
		if len(doc) == 0 {
			continue
		}
		var obj map[string]interface{}
		if err := yaml.Unmarshal(doc, &obj); err != nil || obj == nil {
			continue
		}
		if obj["kind"] != "DaemonSet" {
			continue
		}
		meta, _ := obj["metadata"].(map[string]interface{})
		if meta["name"] != "kube-multus-ds" {
			t.Fatalf("daemonset name = %v", meta["name"])
		}
		podSpec := obj["spec"].(map[string]interface{})["template"].(map[string]interface{})["spec"].(map[string]interface{})
		for _, c := range podSpec["containers"].([]interface{}) {
			ctr := c.(map[string]interface{})
			if ctr["name"] == "kube-multus" {
				main = ctr
			}
		}
		for _, c := range podSpec["initContainers"].([]interface{}) {
			ctr := c.(map[string]interface{})
			if ctr["name"] == "install-multus-binary" {
				init = ctr
			}
		}
	}
	if main == nil || init == nil {
		t.Fatal("kube-multus or install-multus-binary was not rendered")
	}

	req, lim := containerMemory(main)
	if req != "50Mi" || lim != "256Mi" {
		t.Fatalf("kube-multus memory request/limit = %s/%s, want 50Mi/256Mi", req, lim)
	}
	cpuReq, cpuLim := containerCPU(main)
	if cpuReq != "100m" || cpuLim != "100m" {
		t.Fatalf("kube-multus cpu request/limit = %s/%s, want 100m/100m", cpuReq, cpuLim)
	}
	ireq, ilim := containerMemory(init)
	if ireq != "15Mi" || ilim != "" {
		t.Fatalf("install-multus-binary memory request/limit = %s/%q, want 15Mi and no limit", ireq, ilim)
	}
}

func containerMemory(c map[string]interface{}) (request, limit string) {
	res, _ := c["resources"].(map[string]interface{})
	if req, ok := res["requests"].(map[string]interface{}); ok {
		request, _ = req["memory"].(string)
	}
	if lim, ok := res["limits"].(map[string]interface{}); ok {
		limit, _ = lim["memory"].(string)
	}
	return request, limit
}

func containerCPU(c map[string]interface{}) (request, limit string) {
	res, _ := c["resources"].(map[string]interface{})
	if req, ok := res["requests"].(map[string]interface{}); ok {
		request, _ = req["cpu"].(string)
	}
	if lim, ok := res["limits"].(map[string]interface{}); ok {
		limit, _ = lim["cpu"].(string)
	}
	return request, limit
}

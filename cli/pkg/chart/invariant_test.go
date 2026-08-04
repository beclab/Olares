package chart

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/client-go/kubernetes/scheme"
)

// helmActionRE matches the Helm actions the conversion writes into a template.
// Substituting 1 leaves both of them decodable: spec.replicas takes it as the
// number it is, metadata.namespace as the quoted string it is.
var helmActionRE = regexp.MustCompile(`\{\{[^{}]*\}\}`)

// assertChartInvariants fails on any name, label value or cross-object reference
// in the rendered chart that Kubernetes would not accept, so a conversion that
// silently renames one side of a reference cannot pass. Local lint checks the
// chart's structure and Olares' own manifest rules, not this.
//
// allowed lists substrings of the findings the caller expects, for the compose
// input the conversion can only report and not repair. Every entry has to match
// a finding, so an exception cannot outlive the thing it excuses.
func assertChartInvariants(t *testing.T, chartDir string, allowed ...string) {
	t.Helper()

	findings := chartFindings(decodeTemplates(t, chartDir))

	unexpected := make([]string, 0, len(findings))
	matched := make(map[string]bool, len(allowed))
	for _, finding := range findings {
		expected := false
		for _, want := range allowed {
			if strings.Contains(finding, want) {
				matched[want], expected = true, true
			}
		}
		if !expected {
			unexpected = append(unexpected, finding)
		}
	}
	if len(unexpected) > 0 {
		t.Fatalf("chart would not install:\n  %s", strings.Join(unexpected, "\n  "))
	}
	for _, want := range allowed {
		if !matched[want] {
			t.Fatalf("nothing is wrong with %q any more, so drop it from the expected findings (findings: %v)", want, findings)
		}
	}
}

// chartObject pairs an object with the kind the decoder read it as: a typed
// object's own TypeMeta is empty once decoded.
type chartObject struct {
	kind string
	obj  runtime.Object
	meta metav1.Object
}

func (o chartObject) String() string { return o.kind + "/" + o.meta.GetName() }

func decodeTemplates(t *testing.T, chartDir string) []chartObject {
	t.Helper()

	templatesDir := filepath.Join(chartDir, "templates")
	entries, err := os.ReadDir(templatesDir)
	if err != nil {
		t.Fatal(err)
	}
	objects := make([]chartObject, 0, len(entries))
	for _, entry := range entries {
		raw := readFile(t, filepath.Join(templatesDir, entry.Name()))
		obj, gvk, err := scheme.Codecs.UniversalDeserializer().Decode(
			[]byte(helmActionRE.ReplaceAllString(raw, "1")), nil, nil)
		if err != nil {
			t.Fatalf("template %s does not decode as a Kubernetes object: %v", entry.Name(), err)
		}
		meta, ok := obj.(metav1.Object)
		if !ok {
			t.Fatalf("template %s decodes to %T, which carries no metadata", entry.Name(), obj)
		}
		objects = append(objects, chartObject{kind: gvk.Kind, obj: obj, meta: meta})
	}
	return objects
}

func chartFindings(objects []chartObject) []string {
	findings := make([]string, 0)

	present := make(map[string]bool)
	podLabels := make([]map[string]string, 0, len(objects))
	for _, o := range objects {
		present[o.String()] = true
		if spec := podTemplateOf(o.obj); spec != nil {
			podLabels = append(podLabels, spec.Labels)
		}
	}
	// resolves reports a reference that names no object the chart carries; every
	// one of them is a silent failure, since the API server accepts most of them
	// and only the pod never comes up.
	resolves := func(owner, field, kind, name string) {
		if name != "" && !present[kind+"/"+name] {
			findings = append(findings, fmt.Sprintf("%s: %s %q names no %s in the chart", owner, field, name, kind))
		}
	}

	for _, o := range objects {
		findings = append(findings, nameFindings(o)...)
		findings = append(findings, labelFindings(o.String()+" metadata.labels", o.meta.GetLabels())...)

		switch obj := o.obj.(type) {
		case *corev1.Service:
			findings = append(findings, labelFindings(o.String()+" spec.selector", obj.Spec.Selector)...)
			if len(obj.Spec.Selector) > 0 && !selectsAnyPod(obj.Spec.Selector, podLabels) {
				findings = append(findings, fmt.Sprintf("%s: spec.selector %v selects no pod template in the chart", o, obj.Spec.Selector))
			}
		case *appsv1.StatefulSet:
			resolves(o.String(), "spec.serviceName", "Service", obj.Spec.ServiceName)
		case *networkingv1.Ingress:
			for i := range obj.Spec.Rules {
				http := obj.Spec.Rules[i].HTTP
				if http == nil {
					continue
				}
				for j := range http.Paths {
					if backend := http.Paths[j].Backend.Service; backend != nil {
						resolves(o.String(), fmt.Sprintf("spec.rules[%d].http.paths[%d].backend.service.name", i, j), "Service", backend.Name)
					}
				}
			}
		}

		if spec := podTemplateOf(o.obj); spec != nil {
			findings = append(findings, labelFindings(o.String()+" spec.template.metadata.labels", spec.Labels)...)
		}
		if selector := podSelectorOf(o.obj); selector != nil {
			findings = append(findings, labelFindings(o.String()+" spec.selector.matchLabels", selector.MatchLabels)...)
		}
		if spec := podSpecOf(o.obj); spec != nil {
			findings = append(findings, podFindings(o, spec, resolves)...)
		}
	}
	return findings
}

// podFindings checks the names a pod template carries and the objects it points
// at. A volume mount may also name a StatefulSet claim template, whose volume is
// implicit rather than listed in spec.volumes.
func podFindings(o chartObject, spec *corev1.PodSpec, resolves func(owner, field, kind, name string)) []string {
	findings := make([]string, 0)

	mountable := make(map[string]bool)
	if sts, ok := o.obj.(*appsv1.StatefulSet); ok {
		for _, claim := range sts.Spec.VolumeClaimTemplates {
			mountable[claim.Name] = true
			findings = append(findings, labelNameFindings(o.String(), fmt.Sprintf("volumeClaimTemplates %q", claim.Name), claim.Name)...)
		}
	}
	for i := range spec.Volumes {
		vol := &spec.Volumes[i]
		mountable[vol.Name] = true
		findings = append(findings, labelNameFindings(o.String(), fmt.Sprintf("volume %q", vol.Name), vol.Name)...)

		field := fmt.Sprintf("spec.volumes[%d]", i)
		if claim := vol.PersistentVolumeClaim; claim != nil {
			resolves(o.String(), field+".persistentVolumeClaim.claimName", "PersistentVolumeClaim", claim.ClaimName)
		}
		if cm := vol.ConfigMap; cm != nil {
			resolves(o.String(), field+".configMap.name", "ConfigMap", cm.Name)
		}
		if secret := vol.Secret; secret != nil {
			resolves(o.String(), field+".secret.secretName", "Secret", secret.SecretName)
		}
		if vol.Projected != nil {
			for j := range vol.Projected.Sources {
				source := vol.Projected.Sources[j]
				if cm := source.ConfigMap; cm != nil {
					resolves(o.String(), fmt.Sprintf("%s.projected.sources[%d].configMap.name", field, j), "ConfigMap", cm.Name)
				}
				if secret := source.Secret; secret != nil {
					resolves(o.String(), fmt.Sprintf("%s.projected.sources[%d].secret.name", field, j), "Secret", secret.Name)
				}
			}
		}
	}

	for _, containers := range [][]corev1.Container{spec.Containers, spec.InitContainers} {
		for i := range containers {
			container := &containers[i]
			findings = append(findings, labelNameFindings(o.String(), fmt.Sprintf("container %q", container.Name), container.Name)...)
			for _, mount := range container.VolumeMounts {
				if !mountable[mount.Name] {
					findings = append(findings, fmt.Sprintf("%s: container %q mounts %q, which is not a volume of its pod", o, container.Name, mount.Name))
				}
			}
			for j := range container.EnvFrom {
				if ref := container.EnvFrom[j].ConfigMapRef; ref != nil {
					resolves(o.String(), fmt.Sprintf("container %q envFrom[%d].configMapRef.name", container.Name, j), "ConfigMap", ref.Name)
				}
				if ref := container.EnvFrom[j].SecretRef; ref != nil {
					resolves(o.String(), fmt.Sprintf("container %q envFrom[%d].secretRef.name", container.Name, j), "Secret", ref.Name)
				}
			}
			for j := range container.Env {
				from := container.Env[j].ValueFrom
				if from == nil {
					continue
				}
				if ref := from.ConfigMapKeyRef; ref != nil {
					resolves(o.String(), fmt.Sprintf("container %q env[%d].valueFrom.configMapKeyRef.name", container.Name, j), "ConfigMap", ref.Name)
				}
				if ref := from.SecretKeyRef; ref != nil {
					resolves(o.String(), fmt.Sprintf("container %q env[%d].valueFrom.secretKeyRef.name", container.Name, j), "Secret", ref.Name)
				}
			}
		}
	}
	return findings
}

// nameFindings validates an object's own name: a Service name has to be a
// DNS-1035 label, while any other object name may be a DNS subdomain.
func nameFindings(o chartObject) []string {
	name := o.meta.GetName()
	var msgs []string
	if o.kind == "Service" {
		msgs = validation.IsDNS1035Label(name)
	} else {
		msgs = validation.IsDNS1123Subdomain(name)
	}
	findings := make([]string, 0, len(msgs))
	for _, msg := range msgs {
		findings = append(findings, fmt.Sprintf("%s: metadata.name %q is invalid: %s", o, name, msg))
	}
	return findings
}

// labelNameFindings validates the names that have to be DNS labels rather than
// subdomains: pod volumes, claim templates and containers.
func labelNameFindings(owner, field, name string) []string {
	msgs := validation.IsDNS1123Label(name)
	findings := make([]string, 0, len(msgs))
	for _, msg := range msgs {
		findings = append(findings, fmt.Sprintf("%s: %s is invalid: %s", owner, field, msg))
	}
	return findings
}

func labelFindings(where string, labels map[string]string) []string {
	findings := make([]string, 0)
	for key, value := range labels {
		for _, msg := range validation.IsQualifiedName(key) {
			findings = append(findings, fmt.Sprintf("%s: key %q is invalid: %s", where, key, msg))
		}
		for _, msg := range validation.IsValidLabelValue(value) {
			findings = append(findings, fmt.Sprintf("%s: %q is invalid: %s", where, value, msg))
		}
	}
	return findings
}

func selectsAnyPod(selector map[string]string, pods []map[string]string) bool {
	for _, labels := range pods {
		matched := true
		for key, value := range selector {
			if labels[key] != value {
				matched = false
				break
			}
		}
		if matched {
			return true
		}
	}
	return false
}

func podTemplateOf(resource runtime.Object) *metav1.ObjectMeta {
	switch obj := resource.(type) {
	case *appsv1.Deployment:
		return &obj.Spec.Template.ObjectMeta
	case *appsv1.StatefulSet:
		return &obj.Spec.Template.ObjectMeta
	case *appsv1.DaemonSet:
		return &obj.Spec.Template.ObjectMeta
	case *corev1.Pod:
		return &obj.ObjectMeta
	}
	return nil
}

func podSelectorOf(resource runtime.Object) *metav1.LabelSelector {
	switch obj := resource.(type) {
	case *appsv1.Deployment:
		return obj.Spec.Selector
	case *appsv1.StatefulSet:
		return obj.Spec.Selector
	case *appsv1.DaemonSet:
		return obj.Spec.Selector
	}
	return nil
}

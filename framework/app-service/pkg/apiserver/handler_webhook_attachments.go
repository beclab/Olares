package apiserver

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	userspacev1 "github.com/beclab/Olares/framework/app-service/pkg/users/userspace/v1"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	admissionv1 "k8s.io/api/admission/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	attachmentInjectedAnnotation = "app.bytetrade.io/attachments-injected"
	attachmentGroupsAnnotation   = "app.bytetrade.io/attachment-supplemental-groups"
	attachmentVolumePrefix       = "olares-attachment-"
	deviceResourcePrefix         = "devices.bytetrade.io/"
	bluezVolumeName              = "olares-bluez"
	bluezProxyContainerName      = "bluez-dbus-proxy"
)

// Olares hosts use the standard dialout, audio, and video group IDs for
// serial, audio, and video device nodes. Kubernetes supplementalGroups accepts
// numeric IDs only, so keep the device capability-to-GID policy here.
var deviceSupplementalGroups = map[string]int64{
	"device.serial": 20, // dialout
	"device.audio":  29, // audio
	"device.video":  44, // video
}

func (h *Handler) createAttachmentPatch(ctx context.Context, req *admissionv1.AdmissionRequest, tpl *corev1.PodTemplateSpec, cfg *appcfg.ApplicationConfig) ([]byte, error) {
	if tpl == nil || cfg == nil {
		return nil, nil
	}
	if _, declared := cfg.WorkloadOptions[req.Name]; !declared {
		return nil, nil
	}
	var applications appv1alpha1.ApplicationList
	if err := h.ctrlClient.List(ctx, &applications); err != nil {
		return nil, err
	}
	var application *appv1alpha1.Application
	for i := range applications.Items {
		if applications.Items[i].Spec.Namespace == req.Namespace {
			application = &applications.Items[i]
			break
		}
	}
	if application == nil {
		return nil, nil
	}
	before := append([]byte(nil), req.Object.Raw...)
	managedBefore := tpl.Annotations != nil && tpl.Annotations[attachmentInjectedAnnotation] == "true"
	cleanupManagedAttachments(tpl, managedBefore)
	if err := h.injectAttachments(ctx, tpl, req.Name, cfg, application); err != nil {
		return nil, err
	}
	var current []byte
	var err error
	switch req.Kind.Kind {
	case deployment:
		var workload appsv1.Deployment
		if err := json.Unmarshal(req.Object.Raw, &workload); err != nil {
			return nil, err
		}
		workload.Spec.Template = *tpl
		current, err = json.Marshal(&workload)
	case statefulSet:
		var workload appsv1.StatefulSet
		if err := json.Unmarshal(req.Object.Raw, &workload); err != nil {
			return nil, err
		}
		workload.Spec.Template = *tpl
		current, err = json.Marshal(&workload)
	default:
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	patch := admission.PatchResponseFromRaw(before, current)
	return json.Marshal(patch.Patches)
}

func cleanupManagedAttachments(tpl *corev1.PodTemplateSpec, managedBefore bool) {
	if !managedBefore {
		return
	}
	volumes := tpl.Spec.Volumes[:0]
	for _, volume := range tpl.Spec.Volumes {
		if strings.HasPrefix(volume.Name, attachmentVolumePrefix) || volume.Name == bluezVolumeName {
			continue
		}
		volumes = append(volumes, volume)
	}
	tpl.Spec.Volumes = volumes
	containers := tpl.Spec.Containers[:0]
	for _, container := range tpl.Spec.Containers {
		if container.Name == bluezProxyContainerName {
			continue
		}
		mounts := container.VolumeMounts[:0]
		for _, mount := range container.VolumeMounts {
			if strings.HasPrefix(mount.Name, attachmentVolumePrefix) || mount.Name == bluezVolumeName {
				continue
			}
			mounts = append(mounts, mount)
		}
		container.VolumeMounts = mounts
		envs := container.Env[:0]
		for _, env := range container.Env {
			if env.Name != "DBUS_SYSTEM_BUS_ADDRESS" {
				envs = append(envs, env)
			}
		}
		container.Env = envs
		if managedBefore {
			for name := range container.Resources.Requests {
				if strings.HasPrefix(string(name), deviceResourcePrefix) {
					delete(container.Resources.Requests, name)
				}
			}
			for name := range container.Resources.Limits {
				if strings.HasPrefix(string(name), deviceResourcePrefix) {
					delete(container.Resources.Limits, name)
				}
			}
		}
		containers = append(containers, container)
	}
	tpl.Spec.Containers = containers
	if managedBefore {
		delete(tpl.Spec.NodeSelector, corev1.LabelHostname)
		if tpl.Spec.SecurityContext != nil {
			managedGroups := previouslyInjectedDeviceGroups(tpl.Annotations)
			groups := tpl.Spec.SecurityContext.SupplementalGroups[:0]
			for _, group := range tpl.Spec.SecurityContext.SupplementalGroups {
				if _, managed := managedGroups[group]; !managed {
					groups = append(groups, group)
				}
			}
			tpl.Spec.SecurityContext.SupplementalGroups = groups
		}
		removeBluetoothAffinity(tpl)
	}
	if tpl.Annotations != nil {
		delete(tpl.Annotations, attachmentInjectedAnnotation)
		delete(tpl.Annotations, attachmentGroupsAnnotation)
	}
}

func (h *Handler) injectAttachments(ctx context.Context, tpl *corev1.PodTemplateSpec, workload string, cfg *appcfg.ApplicationConfig, application *appv1alpha1.Application) error {
	option, declared := cfg.WorkloadOptions[workload]
	if !declared {
		return nil
	}
	capabilities := make(map[string][]string)
	for _, capability := range option.Allow {
		capabilities[capability.Type] = capability.Containers
	}
	var nodeName string
	deviceAttached := false
	deviceGroups := make(map[int64]struct{})
	injected := false
	for _, attachment := range application.Spec.Attachments {
		if attachment.Workload != workload {
			continue
		}
		injected = true
		targets, allowed := capabilities[attachment.Type]
		if !allowed {
			continue
		}
		if attachment.Node != "" {
			if nodeName != "" && nodeName != attachment.Node {
				return fmt.Errorf("workload %s attachments require different nodes", workload)
			}
			nodeName = attachment.Node
		}
		if attachment.Type == "folder" {
			hostPath, err := h.resolveAttachmentPath(ctx, cfg, attachment)
			if err != nil {
				return err
			}
			volumeName := attachmentVolumePrefix + attachment.Name
			hostPathType := corev1.HostPathDirectory
			tpl.Spec.Volumes = append(tpl.Spec.Volumes, corev1.Volume{Name: volumeName, VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: hostPath, Type: &hostPathType}}})
			if err := addMountToContainers(tpl, targets, corev1.VolumeMount{Name: volumeName, MountPath: "/olares/attachments/" + attachment.Name}); err != nil {
				return err
			}
			continue
		}
		if strings.HasPrefix(attachment.Type, "device.") {
			if len(targets) != 1 {
				return fmt.Errorf("device capability %s must target exactly one container", attachment.Type)
			}
			container := findContainer(tpl, targets[0])
			if container == nil {
				return fmt.Errorf("target container %s not found", targets[0])
			}
			if container.Resources.Requests == nil {
				container.Resources.Requests = corev1.ResourceList{}
			}
			if container.Resources.Limits == nil {
				container.Resources.Limits = corev1.ResourceList{}
			}
			quantity := resource.MustParse("1")
			container.Resources.Requests[corev1.ResourceName(attachment.Ref)] = quantity
			container.Resources.Limits[corev1.ResourceName(attachment.Ref)] = quantity
			if group, ok := deviceSupplementalGroups[attachment.Type]; ok {
				deviceGroups[group] = struct{}{}
			}
			deviceAttached = true
		}
	}
	if nodeName != "" {
		if tpl.Spec.NodeSelector == nil {
			tpl.Spec.NodeSelector = map[string]string{}
		}
		if existing := tpl.Spec.NodeSelector[corev1.LabelHostname]; existing != "" && existing != nodeName {
			return fmt.Errorf("workload %s is already pinned to node %s, attachment requires %s", workload, existing, nodeName)
		}
		if !requiredAffinityAllowsHostname(tpl.Spec.Affinity, nodeName) {
			return fmt.Errorf("workload %s required node affinity excludes attachment node %s", workload, nodeName)
		}
		tpl.Spec.NodeSelector[corev1.LabelHostname] = nodeName
	}
	if deviceAttached {
		if tpl.Spec.SecurityContext == nil {
			tpl.Spec.SecurityContext = &corev1.PodSecurityContext{}
		}
		groups := make([]int64, 0, len(deviceGroups))
		for group := range deviceGroups {
			groups = append(groups, group)
		}
		sort.Slice(groups, func(i, j int) bool { return groups[i] < groups[j] })
		addedGroups := make([]string, 0, len(groups))
		for _, group := range groups {
			if !containsInt64(tpl.Spec.SecurityContext.SupplementalGroups, group) {
				tpl.Spec.SecurityContext.SupplementalGroups = append(tpl.Spec.SecurityContext.SupplementalGroups, group)
				addedGroups = append(addedGroups, strconv.FormatInt(group, 10))
			}
		}
		if len(addedGroups) > 0 {
			if tpl.Annotations == nil {
				tpl.Annotations = map[string]string{}
			}
			tpl.Annotations[attachmentGroupsAnnotation] = strings.Join(addedGroups, ",")
		}
	}
	bluetooth, _ := strconv.ParseBool(application.Spec.Settings["bluetooth"])
	if bluetooth {
		if targets, ok := capabilities["bluetooth"]; ok {
			if err := injectBluetooth(tpl, targets); err != nil {
				return err
			}
			injected = true
		}
	}
	if injected {
		if tpl.Annotations == nil {
			tpl.Annotations = map[string]string{}
		}
		tpl.Annotations[attachmentInjectedAnnotation] = "true"
	}
	return nil
}

func previouslyInjectedDeviceGroups(annotations map[string]string) map[int64]struct{} {
	groups := make(map[int64]struct{})
	raw := annotations[attachmentGroupsAnnotation]
	for _, value := range strings.Split(raw, ",") {
		group, err := strconv.ParseInt(value, 10, 64)
		if err == nil {
			groups[group] = struct{}{}
		}
	}
	return groups
}

func containsInt64(values []int64, want int64) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func requiredAffinityAllowsHostname(affinity *corev1.Affinity, nodeName string) bool {
	if affinity == nil || affinity.NodeAffinity == nil || affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
		return true
	}
	terms := affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
	if len(terms) == 0 {
		return true
	}
	for _, term := range terms {
		allowed := true
		for _, expression := range append(append([]corev1.NodeSelectorRequirement{}, term.MatchExpressions...), term.MatchFields...) {
			if expression.Key != corev1.LabelHostname && expression.Key != "metadata.name" {
				continue
			}
			switch expression.Operator {
			case corev1.NodeSelectorOpIn:
				allowed = allowed && containsString(expression.Values, nodeName)
			case corev1.NodeSelectorOpNotIn:
				allowed = allowed && !containsString(expression.Values, nodeName)
			case corev1.NodeSelectorOpDoesNotExist:
				allowed = false
			}
		}
		if allowed {
			return true
		}
	}
	return false
}

func containsString(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}

func (h *Handler) resolveAttachmentPath(ctx context.Context, cfg *appcfg.ApplicationConfig, attachment appv1alpha1.Attachment) (string, error) {
	parts := strings.Split(attachment.Ref, "/")
	if len(parts) < 3 {
		return "", fmt.Errorf("invalid folder ref %q", attachment.Ref)
	}
	if parts[0] == "external" {
		if len(parts) < 4 || attachment.Node == "" || attachment.Node != parts[1] {
			return "", fmt.Errorf("external folder node must match ref %q", attachment.Ref)
		}
		sharedRoot := os.Getenv("SHARED_LIB_PATH")
		if sharedRoot == "" {
			return "", fmt.Errorf("external folder root is unavailable")
		}
		return filepath.Join(append([]string{sharedRoot}, parts[2:]...)...), nil
	}
	rootPath := userspacev1.DefaultRootPath
	if configured := os.Getenv(userspacev1.OlaresRootPath); configured != "" {
		rootPath = configured
	}
	if parts[0] == "drive" && parts[1] == "Common" {
		root := filepath.Join(rootPath, "rootfs", "Common")
		return filepath.Join(append([]string{root}, parts[2:]...)...), nil
	}
	var bfl appsv1.StatefulSet
	if err := h.ctrlClient.Get(ctx, types.NamespacedName{Name: "bfl", Namespace: "user-space-" + cfg.OwnerName}, &bfl); err != nil {
		return "", err
	}
	userspaceRoot := bfl.Annotations[constants.UserSpaceDirKey]
	if userspaceRoot == "" {
		return "", fmt.Errorf("userspace path is unavailable for %s", cfg.OwnerName)
	}
	switch {
	case parts[1] == "Home":
		root := filepath.Join(userspaceRoot, "Home")
		return filepath.Join(append([]string{root}, parts[2:]...)...), nil
	case parts[1] == "Data" && len(parts) >= 4 && parts[2] == cfg.AppName:
		root := filepath.Join(userspaceRoot, "Data", cfg.AppName)
		return filepath.Join(append([]string{root}, parts[3:]...)...), nil
	default:
		return "", fmt.Errorf("folder ref %q is outside an allowed root", attachment.Ref)
	}
}

func addMountToContainers(tpl *corev1.PodTemplateSpec, targets []string, mount corev1.VolumeMount) error {
	for _, target := range targets {
		container := findContainer(tpl, target)
		if container == nil {
			return fmt.Errorf("target container %s not found", target)
		}
		container.VolumeMounts = append(container.VolumeMounts, mount)
	}
	return nil
}

func findContainer(tpl *corev1.PodTemplateSpec, name string) *corev1.Container {
	for i := range tpl.Spec.Containers {
		if tpl.Spec.Containers[i].Name == name {
			return &tpl.Spec.Containers[i]
		}
	}
	return nil
}

func injectBluetooth(tpl *corev1.PodTemplateSpec, targets []string) error {
	image := os.Getenv("BLUEZ_DBUS_PROXY_IMAGE")
	if image == "" {
		return fmt.Errorf("BLUEZ_DBUS_PROXY_IMAGE is required when bluetooth is enabled")
	}
	socketType := corev1.HostPathSocket
	tpl.Spec.Volumes = append(tpl.Spec.Volumes,
		corev1.Volume{Name: bluezVolumeName, VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
		corev1.Volume{Name: attachmentVolumePrefix + "bluez-host", VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: "/run/dbus/system_bus_socket", Type: &socketType}}},
	)
	tpl.Spec.Containers = append(tpl.Spec.Containers, corev1.Container{
		Name: bluezProxyContainerName, Image: image,
		Command: []string{"sh", "-ec"},
		Args:    []string{"umask 0000; exec xdg-dbus-proxy unix:path=/run/dbus/system_bus_socket /run/olares-bluez/system_bus_socket --filter --talk=org.bluez --talk=org.freedesktop.DBus"},
		SecurityContext: &corev1.SecurityContext{
			AllowPrivilegeEscalation: ptr.To(false),
			AppArmorProfile:          &corev1.AppArmorProfile{Type: corev1.AppArmorProfileTypeUnconfined},
			ReadOnlyRootFilesystem:   ptr.To(true),
			RunAsNonRoot:             ptr.To(false),
			RunAsUser:                ptr.To(int64(0)),
		},
		VolumeMounts: []corev1.VolumeMount{
			{Name: attachmentVolumePrefix + "bluez-host", MountPath: "/run/dbus/system_bus_socket", ReadOnly: true},
			{Name: bluezVolumeName, MountPath: "/run/olares-bluez"},
		},
		StartupProbe: &corev1.Probe{
			ProbeHandler:     corev1.ProbeHandler{Exec: &corev1.ExecAction{Command: []string{"test", "-S", "/run/olares-bluez/system_bus_socket"}}},
			PeriodSeconds:    1,
			FailureThreshold: 30,
		},
		ReadinessProbe: &corev1.Probe{
			ProbeHandler:     corev1.ProbeHandler{Exec: &corev1.ExecAction{Command: []string{"test", "-S", "/run/olares-bluez/system_bus_socket"}}},
			PeriodSeconds:    5,
			FailureThreshold: 3,
		},
	})
	for _, target := range targets {
		container := findContainer(tpl, target)
		if container == nil {
			return fmt.Errorf("target container %s not found", target)
		}
		container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{Name: bluezVolumeName, MountPath: "/run/olares-bluez"})
		container.Env = append(container.Env, corev1.EnvVar{Name: "DBUS_SYSTEM_BUS_ADDRESS", Value: "unix:path=/run/olares-bluez/system_bus_socket"})
	}
	if tpl.Spec.Affinity == nil {
		tpl.Spec.Affinity = &corev1.Affinity{}
	}
	if tpl.Spec.Affinity.NodeAffinity == nil {
		tpl.Spec.Affinity.NodeAffinity = &corev1.NodeAffinity{}
	}
	if tpl.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
		tpl.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution = &corev1.NodeSelector{}
	}
	required := tpl.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution
	requirement := corev1.NodeSelectorRequirement{Key: bluetoothNodeLabel, Operator: corev1.NodeSelectorOpIn, Values: []string{"true"}}
	if len(required.NodeSelectorTerms) == 0 {
		required.NodeSelectorTerms = []corev1.NodeSelectorTerm{{MatchExpressions: []corev1.NodeSelectorRequirement{requirement}}}
	} else {
		for i := range required.NodeSelectorTerms {
			required.NodeSelectorTerms[i].MatchExpressions = append(required.NodeSelectorTerms[i].MatchExpressions, requirement)
		}
	}
	return nil
}

func removeBluetoothAffinity(tpl *corev1.PodTemplateSpec) {
	if tpl.Spec.Affinity == nil || tpl.Spec.Affinity.NodeAffinity == nil || tpl.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
		return
	}
	required := tpl.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution
	terms := required.NodeSelectorTerms[:0]
	for _, term := range required.NodeSelectorTerms {
		filtered := term.MatchExpressions[:0]
		for _, expression := range term.MatchExpressions {
			if expression.Key != bluetoothNodeLabel {
				filtered = append(filtered, expression)
			}
		}
		term.MatchExpressions = filtered
		if len(term.MatchExpressions) > 0 || len(term.MatchFields) > 0 {
			terms = append(terms, term)
		}
	}
	required.NodeSelectorTerms = terms
}

func combineJSONPatches(first, second []byte) []byte {
	if len(first) == 0 || string(first) == "[]" {
		return second
	}
	if len(second) == 0 || string(second) == "[]" {
		return first
	}
	return append(append(append([]byte{}, first[:len(first)-1]...), ','), second[1:]...)
}

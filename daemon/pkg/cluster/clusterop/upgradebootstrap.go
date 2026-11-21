package clusterop

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/beclab/Olares/daemon/pkg/cluster/inventory"
	"github.com/beclab/Olares/daemon/pkg/commands"
	"github.com/beclab/Olares/daemon/pkg/utils"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

// legacyBootstrapScript is the bootstrap stage (upgradecmd.Bootstrap) in shell,
// run on the host of a worker whose olaresd is too old to run it. It depends on
// the same three frozen things: the request fields, the CLI archive name on the
// CDN, and the target CLI's bootstrap-node command line.
//
// The check that the work is already done covers only olares-cli. The Job only
// runs on a node whose olaresd could not bootstrap itself, so that olaresd is
// never the target yet.
const legacyBootstrapScript = `
set -eu
. /etc/systemd/system/olaresd.service.env
: "${BASE_DIR:?olaresd service has no BASE_DIR}"
cli=/usr/local/bin/olares-cli

installed=$("$cli" -v 2>/dev/null | head -n 1 || true)
case " $installed " in
*" $UPGRADE_VERSION "*|*" v$UPGRADE_VERSION "*) ;;
*)
  case "$(uname -m)" in
  x86_64|amd64) arch=amd64 ;;
  aarch64|arm64) arch=arm64 ;;
  *) echo "unsupported architecture $(uname -m)" >&2; exit 1 ;;
  esac
  url=${CLI_URL:-${UPGRADE_CDN%/}/olares-cli-v${UPGRADE_VERSION}_linux_${arch}.tar.gz}
  work=$(mktemp -d "$BASE_DIR/.olares-cli-bootstrap.XXXXXX")
  trap 'rm -rf "$work"' EXIT
  curl -fsSL --retry 3 -o "$work/cli.tar.gz" "$url"
  tar -xzf "$work/cli.tar.gz" -C "$work"
  chmod 0755 "$work/olares-cli"
  case " $("$work/olares-cli" -v | head -n 1) " in
  *" $UPGRADE_VERSION "*|*" v$UPGRADE_VERSION "*) ;;
  *) echo "the downloaded olares-cli is not $UPGRADE_VERSION" >&2; exit 1 ;;
  esac
  cp "$work/olares-cli" "$cli.bootstrap"
  mv -f "$cli.bootstrap" "$cli"
  ;;
esac

set -- upgrade bootstrap-node --version "$UPGRADE_VERSION" --base-dir "$BASE_DIR" --cdn-service "$UPGRADE_CDN"
if [ -n "$WIZARD_URL" ]; then
  set -- "$@" --url-override "$WIZARD_URL"
fi
"$cli" "$@"
`

// legacyBootstrapImages are the images the Job may run in, tried in order. The
// script only needs chroot and sh from the image, so any image with a shell
// does; what matters is that the node already has one, because an offline
// worker cannot pull.
//
// TODO: inferred from the charts and not yet confirmed on a 1.12.8 machine.
// aboveos-busybox is injected by app-service into user application pods, and
// busybox:1.28 runs in bfl and certificate jobs; the pinned alpine is the
// previous Job's image and needs a registry.
var legacyBootstrapImages = []string{
	"docker.io/beclab/aboveos-busybox:1.37.0",
	"docker.io/library/busybox:1.28",
	"docker.io/library/alpine@sha256:d9e853e87e55526f6b2917df91a2115c36dd7c696a35be12163d44e6e2a4b6bc",
}

var legacyBootstrapPoll = 5 * time.Second

const legacyBootstrapLogLines = 40

// LegacyBootstrapError is a Job that ran and failed. Log is the end of its
// output, which is where olares-cli prints the reason.
type LegacyBootstrapError struct {
	Job string
	Log string
}

func (e *LegacyBootstrapError) Error() string {
	return "legacy bootstrap job " + e.Job + " failed: " + e.Log
}

func legacyBootstrapReason(node inventory.Node, err error) string {
	var failed *LegacyBootstrapError
	if errors.As(err, &failed) {
		klog.Warningf("clusterop: bootstrap node %s through the legacy Job: %v", node.NodeName, err)
		if detail := commands.CommandErrorSummary(failed.Log); detail != "" {
			return detail
		}
	}
	return suppress(CodeStageFailed, "bootstrap node "+node.NodeName+" through the legacy Job", err)
}

// BootstrapLegacyUpgradeWorker runs the bootstrap stage on a worker through a
// Kubernetes Job pinned to it: privileged, in the host's PID namespace so that
// systemctl can restart olaresd, and chrooted into the host. No SSH credential
// or owner signature reaches the worker.
func BootstrapLegacyUpgradeWorker(ctx context.Context, node inventory.Node, operationID, version string, locations UpgradeLocations) error {
	client, err := utils.GetKubeClient()
	if err != nil {
		return err
	}
	return bootstrapLegacyUpgradeWorker(ctx, client, node, operationID, version, locations)
}

// bootstrapLegacyUpgradeWorker goes through the images in order. A Job whose
// pod cannot get its image is deleted and the next image tried; any other
// outcome is the answer. The Job names come from the operation and the node,
// so a run resumed after olaresd restarted finds the Job it started rather
// than starting a second one.
func bootstrapLegacyUpgradeWorker(ctx context.Context, client kubernetes.Interface, node inventory.Node,
	operationID, version string, locations UpgradeLocations) error {
	if node.NodeName == "" || operationID == "" || version == "" {
		return fmt.Errorf("legacy worker bootstrap needs a node, operation and version")
	}
	if locations.CDN == "" {
		locations.CDN = commands.OLARES_CDN_SERVICE
	}
	sum := sha256.Sum256([]byte(operationID + "\x00" + node.NodeName))
	base := "olares-upgrade-bootstrap-" + hex.EncodeToString(sum[:12])

	var pullErrs []error
	for i, image := range legacyBootstrapImages {
		name := base + "-" + strconv.Itoa(i)
		err := runLegacyBootstrapJob(ctx, client, legacyBootstrapJob(name, node.NodeName, image, version, locations))
		var pull *imagePullError
		if !errors.As(err, &pull) {
			return err
		}
		klog.Warningf("clusterop: node %s cannot run the bootstrap Job in %s: %s; trying the next image",
			node.NodeName, image, pull.reason)
		pullErrs = append(pullErrs, err)
	}
	return fmt.Errorf("node %s has none of the bootstrap images: %w", node.NodeName, errors.Join(pullErrs...))
}

type imagePullError struct {
	image, reason string
}

func (e *imagePullError) Error() string { return e.image + ": " + e.reason }

func runLegacyBootstrapJob(ctx context.Context, client kubernetes.Interface, want *batchv1.Job) error {
	jobs := client.BatchV1().Jobs(upgradeSecretNamespace)
	job, err := jobs.Get(ctx, want.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		job, err = jobs.Create(ctx, want, metav1.CreateOptions{})
		if apierrors.IsAlreadyExists(err) {
			job, err = jobs.Get(ctx, want.Name, metav1.GetOptions{})
		}
	}
	if err != nil {
		return fmt.Errorf("create legacy bootstrap job %s: %w", want.Name, err)
	}
	if !sameBootstrapJob(job, want) {
		return fmt.Errorf("legacy bootstrap job %s exists for another node, image or version", want.Name)
	}

	for {
		switch {
		case job.Status.Succeeded > 0:
			return nil
		case job.Status.Failed > 0:
			return &LegacyBootstrapError{Job: job.Name, Log: legacyBootstrapLog(ctx, client, job.Name)}
		}
		if reason := imagePullFailure(ctx, client, job.Name); reason != "" {
			propagation := metav1.DeletePropagationBackground
			if err := jobs.Delete(ctx, job.Name, metav1.DeleteOptions{PropagationPolicy: &propagation}); err != nil && !apierrors.IsNotFound(err) {
				return fmt.Errorf("delete legacy bootstrap job %s: %w", job.Name, err)
			}
			return &imagePullError{image: want.Spec.Template.Spec.Containers[0].Image, reason: reason}
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("legacy bootstrap job %s did not finish: %w", job.Name, ctx.Err())
		case <-time.After(legacyBootstrapPoll):
		}
		if job, err = jobs.Get(ctx, want.Name, metav1.GetOptions{}); err != nil {
			return fmt.Errorf("read legacy bootstrap job %s: %w", want.Name, err)
		}
	}
}

func jobPods(ctx context.Context, client kubernetes.Interface, job string) []corev1.Pod {
	pods, err := client.CoreV1().Pods(upgradeSecretNamespace).List(ctx, metav1.ListOptions{LabelSelector: "job-name=" + job})
	if err != nil {
		klog.V(4).Infof("clusterop: list pods of %s: %v", job, err)
		return nil
	}
	return pods.Items
}

func imagePullFailure(ctx context.Context, client kubernetes.Interface, job string) string {
	for _, pod := range jobPods(ctx, client, job) {
		for _, c := range pod.Status.ContainerStatuses {
			if w := c.State.Waiting; w != nil {
				switch w.Reason {
				case "ErrImagePull", "ImagePullBackOff", "ErrImageNeverPull", "InvalidImageName":
					return w.Reason
				}
			}
		}
	}
	return ""
}

func legacyBootstrapLog(ctx context.Context, client kubernetes.Interface, job string) string {
	tail := int64(legacyBootstrapLogLines)
	for _, pod := range jobPods(ctx, client, job) {
		out, err := client.CoreV1().Pods(upgradeSecretNamespace).
			GetLogs(pod.Name, &corev1.PodLogOptions{TailLines: &tail}).DoRaw(ctx)
		if err == nil && len(out) > 0 {
			return string(out)
		}
	}
	return ""
}

func legacyBootstrapJob(name, nodeName, image, version string, locations UpgradeLocations) *batchv1.Job {
	backoff := int32(0)
	ttl := int32(3600)
	active := int64(3600)
	privileged := true
	automountToken := false
	hostPathType := corev1.HostPathDirectory
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: upgradeSecretNamespace,
			Labels: map[string]string{"olares.io/upgrade-bootstrap": "true"}},
		Spec: batchv1.JobSpec{
			BackoffLimit: &backoff, TTLSecondsAfterFinished: &ttl, ActiveDeadlineSeconds: &active,
			Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{
				// systemctl authenticates the host systemd peer by PID. Without
				// host PID visibility, bootstrap-node cannot restart olaresd.
				NodeName: nodeName, HostNetwork: true, HostPID: true, RestartPolicy: corev1.RestartPolicyNever,
				AutomountServiceAccountToken: &automountToken,
				Tolerations:                  []corev1.Toleration{{Operator: corev1.TolerationOpExists}},
				Volumes: []corev1.Volume{{Name: "host", VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{Path: "/", Type: &hostPathType},
				}}},
				Containers: []corev1.Container{{
					Name: "bootstrap", Image: image, ImagePullPolicy: corev1.PullIfNotPresent,
					Command: []string{"/bin/sh", "-ec", `chroot /host /bin/sh -ec "$BOOTSTRAP_SCRIPT"`},
					Env: []corev1.EnvVar{
						{Name: "BOOTSTRAP_SCRIPT", Value: legacyBootstrapScript},
						{Name: "UPGRADE_VERSION", Value: version},
						{Name: "UPGRADE_CDN", Value: locations.CDN},
						{Name: "CLI_URL", Value: locations.CliURL},
						{Name: "WIZARD_URL", Value: locations.WizardURL},
					},
					SecurityContext: &corev1.SecurityContext{Privileged: &privileged},
					VolumeMounts:    []corev1.VolumeMount{{Name: "host", MountPath: "/host"}},
				}},
			}},
		},
	}
}

func sameBootstrapJob(job, want *batchv1.Job) bool {
	got, exp := job.Spec.Template.Spec, want.Spec.Template.Spec
	if got.NodeName != exp.NodeName || len(got.Containers) != 1 || got.Containers[0].Image != exp.Containers[0].Image {
		return false
	}
	return jobEnv(job, "UPGRADE_VERSION") == jobEnv(want, "UPGRADE_VERSION")
}

func jobEnv(job *batchv1.Job, name string) string {
	for _, env := range job.Spec.Template.Spec.Containers[0].Env {
		if env.Name == name {
			return env.Value
		}
	}
	return ""
}

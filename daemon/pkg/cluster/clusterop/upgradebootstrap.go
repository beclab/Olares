package clusterop

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/beclab/Olares/daemon/pkg/cluster/inventory"
	"github.com/beclab/Olares/daemon/pkg/commands"
	"github.com/beclab/Olares/daemon/pkg/utils"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const legacyBootstrapScript = `
set -eu
. /etc/systemd/system/olaresd.service.env
: "${BASE_DIR:?olaresd service has no BASE_DIR}"
set -- /usr/local/bin/olares-cli download wizard --version "$UPGRADE_VERSION" --base-dir "$BASE_DIR" --cdn-service "$UPGRADE_CDN"
if [ -n "$WIZARD_URL" ]; then
  set -- "$@" --url-override "$WIZARD_URL"
fi
"$@"
/usr/local/bin/olares-cli download component --version "$UPGRADE_VERSION" --base-dir "$BASE_DIR" --cdn-service "$UPGRADE_CDN"
/usr/local/bin/olares-cli prepare olaresd --version "$UPGRADE_VERSION" --base-dir "$BASE_DIR"
`

// BootstrapLegacyUpgradeWorker upgrades only olaresd on a worker whose old
// daemon cannot serve the staged-upgrade protocol. The old CLI runs on its
// host through a temporary Kubernetes Job; once the new daemon answers, the
// ordinary node-prepare stage installs the target CLI and images. No SSH
// credentials or owner signature are copied to a worker.
func BootstrapLegacyUpgradeWorker(ctx context.Context, node inventory.Node, operationID, version string, locations UpgradeLocations) error {
	client, err := utils.GetKubeClient()
	if err != nil {
		return err
	}
	return bootstrapLegacyUpgradeWorker(ctx, client, node, operationID, version, locations)
}

func bootstrapLegacyUpgradeWorker(ctx context.Context, client kubernetes.Interface, node inventory.Node,
	operationID, version string, locations UpgradeLocations) error {
	if node.NodeName == "" || operationID == "" || version == "" {
		return fmt.Errorf("legacy worker bootstrap needs a node, operation and version")
	}
	sum := sha256.Sum256([]byte(operationID + "\x00" + node.NodeName))
	name := "olares-upgrade-bootstrap-" + hex.EncodeToString(sum[:16])
	jobs := client.BatchV1().Jobs(upgradeSecretNamespace)
	job, err := jobs.Get(ctx, name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		job, err = jobs.Create(ctx, legacyBootstrapJob(name, node.NodeName, version, locations), metav1.CreateOptions{})
		if apierrors.IsAlreadyExists(err) {
			job, err = jobs.Get(ctx, name, metav1.GetOptions{})
		}
	}
	if err != nil {
		return fmt.Errorf("create legacy bootstrap job on %s: %w", node.NodeName, err)
	}
	if job.Spec.Template.Spec.NodeName != node.NodeName || !bootstrapJobVersionIs(job, version) {
		return fmt.Errorf("legacy bootstrap job %s does not match node and version", name)
	}

	deadline := time.NewTimer(2 * time.Hour)
	defer deadline.Stop()
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		if job.Status.Succeeded > 0 {
			return nil
		}
		if job.Status.Failed > 0 {
			return fmt.Errorf("legacy bootstrap job %s failed on %s", name, node.NodeName)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-deadline.C:
			return fmt.Errorf("legacy bootstrap job %s timed out on %s", name, node.NodeName)
		case <-ticker.C:
		}
		job, err = jobs.Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("read legacy bootstrap job %s: %w", name, err)
		}
	}
}

func legacyBootstrapJob(name, nodeName, version string, locations UpgradeLocations) *batchv1.Job {
	backoff := int32(0)
	ttl := int32(3600)
	active := int64(7200)
	privileged := true
	automountToken := false
	hostPathType := corev1.HostPathDirectory
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: upgradeSecretNamespace,
			Labels: map[string]string{"olares.io/upgrade-bootstrap": "true"}},
		Spec: batchv1.JobSpec{
			BackoffLimit: &backoff, TTLSecondsAfterFinished: &ttl, ActiveDeadlineSeconds: &active,
			Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{
				NodeName: nodeName, HostNetwork: true, RestartPolicy: corev1.RestartPolicyNever,
				AutomountServiceAccountToken: &automountToken,
				Volumes: []corev1.Volume{{Name: "host", VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{Path: "/", Type: &hostPathType},
				}}},
				Containers: []corev1.Container{{
					Name: "bootstrap", Image: "docker.io/library/alpine@sha256:d9e853e87e55526f6b2917df91a2115c36dd7c696a35be12163d44e6e2a4b6bc", ImagePullPolicy: corev1.PullIfNotPresent,
					Command: []string{"/bin/sh", "-ec", `chroot /host /bin/sh -ec "$BOOTSTRAP_SCRIPT"`},
					Env: []corev1.EnvVar{
						{Name: "BOOTSTRAP_SCRIPT", Value: legacyBootstrapScript},
						{Name: "UPGRADE_VERSION", Value: version},
						{Name: "UPGRADE_CDN", Value: commands.OLARES_CDN_SERVICE},
						{Name: "WIZARD_URL", Value: locations.WizardURL},
					},
					SecurityContext: &corev1.SecurityContext{Privileged: &privileged},
					VolumeMounts:    []corev1.VolumeMount{{Name: "host", MountPath: "/host"}},
				}},
			}},
		},
	}
}

func bootstrapJobVersionIs(job *batchv1.Job, version string) bool {
	if len(job.Spec.Template.Spec.Containers) != 1 {
		return false
	}
	for _, env := range job.Spec.Template.Spec.Containers[0].Env {
		if env.Name == "UPGRADE_VERSION" {
			return env.Value == version
		}
	}
	return false
}

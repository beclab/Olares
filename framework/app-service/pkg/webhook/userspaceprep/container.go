package userspaceprep

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
)

const userspacePrepareImage = "beclab/aboveos-busybox:1.37.0"

// Script renders the shell the prepare container runs.
//
// List A comes first and is stat-guarded: a directory that is already
// 1000:1000 is left alone, which makes the whole thing a no-op on every
// start after the first. List B then materializes the static subPaths,
// because kubernetes refuses to mount a subPath that does not exist yet.
//
// Every chown is non-recursive. That is the point of the design: the
// cost stays proportional to the number of entries here rather than to
// the size of the data tree, and we never touch the deep inodes that
// make `chown -R` fail with EPERM on FUSE / JuiceFS mounts.
//
// Paths are the prepare container's own mount points, never host paths.
func Script(p Plan) string {
	var b strings.Builder
	b.WriteString("set -e\n")
	for _, t := range p.Targets {
		fmt.Fprintf(&b, "[ \"$(stat -c %%u:%%g %q)\" = \"%d:%d\" ] || chown %d:%d %q\n",
			t.MountPath, constants.UserspaceUID, constants.UserspaceGID,
			constants.UserspaceUID, constants.UserspaceGID, t.MountPath)
	}
	for _, t := range p.Targets {
		for _, sp := range t.SubPaths {
			dir := t.MountPath + "/" + sp
			fmt.Fprintf(&b, "mkdir -p %q\n", dir)
			fmt.Fprintf(&b, "chown %d:%d %q\n", constants.UserspaceUID, constants.UserspaceGID, dir)
		}
	}
	return b.String()
}

// InitContainer builds the prepare container for plan p.
//
// It declares runAsUser/runAsGroup 0 at container level. Container-level
// securityContext wins over the pod-level 1000:1000 injected alongside
// it, so prepare keeps the privileges it needs to chown while the app
// containers still drop to 1000 — no image allow-list required. The
// runAsGroup is spelled out rather than left to inheritance, which would
// otherwise leave this container running as 0:1000.
func InitContainer(p Plan) corev1.Container {
	root := int64(0)

	mounts := make([]corev1.VolumeMount, 0, len(p.Targets))
	for _, t := range p.Targets {
		mounts = append(mounts, corev1.VolumeMount{
			Name:      t.VolumeName,
			MountPath: t.MountPath,
			// Explicitly writable: the app may well mount the same
			// volume read-only, but prepare has to chown it.
			ReadOnly: false,
		})
	}

	return corev1.Container{
		Name:            constants.UserspacePrepareInitContainerName,
		Image:           userspacePrepareImage,
		ImagePullPolicy: corev1.PullIfNotPresent,
		SecurityContext: &corev1.SecurityContext{
			RunAsUser: &root,
		},
		Command:      []string{"sh", "-c", Script(p)},
		VolumeMounts: mounts,
	}
}

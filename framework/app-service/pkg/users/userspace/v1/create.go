package userspace

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/beclab/Olares/framework/app-service/pkg/userstorage"
	iamv1 "github.com/beclab/api/iam/v1alpha2"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	OlaresRootPath  = "OLARES_SYSTEM_ROOT_PATH"
	DefaultRootPath = "/olares"
)

type Creator struct {
	client.Client
	user string
}

func NewCreator(c client.Client, user string) *Creator {
	return &Creator{Client: c, user: user}
}

// CreateStorage creates the user's PVs and PVCs. Shared workloads are installed
// by the CLI, independently of user creation.
func (c *Creator) CreateStorage(ctx context.Context) error {
	var user iamv1.User
	if err := c.Get(ctx, types.NamespacedName{Name: c.user}, &user); err != nil {
		return err
	}
	if user.UID == "" {
		return fmt.Errorf("user %s has no UID", user.Name)
	}
	var existing corev1.PersistentVolumeClaimList
	if err := c.List(ctx, &existing, client.InNamespace(userstorage.Namespace(c.user)), client.HasLabels{userstorage.RoleLabel}); err != nil {
		return err
	}
	nodeName := ""
	for _, claim := range existing.Items {
		if claim.Labels[userstorage.RoleLabel] == userstorage.Userspace {
			nodeName = claim.Annotations[userstorage.NodeAnnotation]
		}
	}
	if nodeName == "" {
		var nodes corev1.NodeList
		if err := c.List(ctx, &nodes); err != nil {
			return err
		}
		sort.Slice(nodes.Items, func(i, j int) bool { return nodes.Items[i].Name < nodes.Items[j].Name })
		for _, node := range nodes.Items {
			if node.DeletionTimestamp != nil || node.Spec.Unschedulable || node.Labels["kubernetes.io/os"] != "linux" {
				continue
			}
			if _, controlPlane := node.Labels["node-role.kubernetes.io/control-plane"]; !controlPlane {
				continue
			}
			for _, condition := range node.Status.Conditions {
				if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue {
					nodeName = node.Name
					break
				}
			}
			if nodeName != "" {
				break
			}
		}
		if nodeName == "" {
			return fmt.Errorf("no ready control-plane node for user storage")
		}
	}
	root := os.Getenv(OlaresRootPath)
	if root == "" {
		root = DefaultRootPath
	}
	suffix := strings.ReplaceAll(string(user.UID), "-", "")
	if len(suffix) > 16 {
		suffix = suffix[:16]
	}
	owner := metav1.OwnerReference{APIVersion: iamv1.SchemeGroupVersion.String(), Kind: "User", Name: user.Name, UID: user.UID}
	for _, spec := range []struct{ role, directory, storageClass, size string }{
		{userstorage.Userspace, "rootfs/userspace", "juicefs-localpath-userspace-", "50Gi"},
		{userstorage.AppCache, "userdata/Cache", "hostpath-appcache-", "50Gi"},
		{userstorage.DBData, "userdata/dbdata", "hostpath-dbdata-", "80Gi"},
	} {
		name := fmt.Sprintf("pvc-%s-%s-%s", spec.role, user.Name, suffix)
		pvName := fmt.Sprintf("pv-%s-%s-%s", spec.role, user.Name, suffix)
		relativePath := filepath.Join(spec.directory, name)
		hostPath := filepath.Join(root, relativePath)
		storageClass := spec.storageClass + user.Name
		directoryOrCreate := corev1.HostPathDirectoryOrCreate
		pv := &corev1.PersistentVolume{
			ObjectMeta: metav1.ObjectMeta{Name: pvName, Labels: map[string]string{userstorage.RoleLabel: spec.role}, OwnerReferences: []metav1.OwnerReference{owner}},
			Spec: corev1.PersistentVolumeSpec{
				Capacity:                      corev1.ResourceList{corev1.ResourceStorage: resource.MustParse(spec.size)},
				AccessModes:                   []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
				PersistentVolumeReclaimPolicy: corev1.PersistentVolumeReclaimDelete,
				StorageClassName:              storageClass,
				ClaimRef:                      &corev1.ObjectReference{Namespace: userstorage.Namespace(user.Name), Name: name},
				PersistentVolumeSource:        corev1.PersistentVolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: hostPath, Type: &directoryOrCreate}},
			},
		}
		if err := c.Create(ctx, pv); err != nil && !apierrors.IsAlreadyExists(err) {
			return err
		}
		claim := &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: userstorage.Namespace(user.Name), Labels: map[string]string{userstorage.RoleLabel: spec.role},
				Annotations:     map[string]string{userstorage.HostPathAnnotation: hostPath, userstorage.RelativePathAnnotation: relativePath, userstorage.NodeAnnotation: nodeName},
				OwnerReferences: []metav1.OwnerReference{owner}},
			Spec: corev1.PersistentVolumeClaimSpec{AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}, StorageClassName: &storageClass, VolumeName: pvName,
				Resources: corev1.VolumeResourceRequirements{Requests: corev1.ResourceList{corev1.ResourceStorage: resource.MustParse(spec.size)}}},
		}
		if err := c.Create(ctx, claim); err != nil && !apierrors.IsAlreadyExists(err) {
			return err
		}
	}
	return nil
}

package userstorage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var ErrNotReady = errors.New("personal storage is not ready")

const (
	RoleLabel                  = "bytetrade.io/user-storage"
	HostPathAnnotation         = "bytetrade.io/host-path"
	RelativePathAnnotation     = "bytetrade.io/storage-relative-path"
	NodeAnnotation             = "bytetrade.io/storage-node"
	InitializedNodesAnnotation = "bytetrade.io/initialized-nodes"
	Userspace                  = "userspace"
	AppCache                   = "appcache"
	DBData                     = "dbdata"
)

type Volume struct {
	Claim corev1.PersistentVolumeClaim
	Path  string
}

type Layout struct {
	Userspace Volume
	AppCache  Volume
	DBData    Volume
	NodeName  string
}

func Namespace(user string) string { return "user-space-" + user }

func FromClaims(user string, claims []corev1.PersistentVolumeClaim) (*Layout, error) {
	layout := &Layout{}
	volumes := map[string]*Volume{Userspace: &layout.Userspace, AppCache: &layout.AppCache, DBData: &layout.DBData}
	for _, claim := range claims {
		volume, ok := volumes[claim.Labels[RoleLabel]]
		if !ok || claim.Namespace != Namespace(user) || claim.DeletionTimestamp != nil {
			continue
		}
		if volume.Claim.Name != "" {
			return nil, fmt.Errorf("multiple %s PVCs for user %s", claim.Labels[RoleLabel], user)
		}
		path := claim.Annotations[HostPathAnnotation]
		if !filepath.IsAbs(path) || filepath.Clean(path) == "/" {
			return nil, fmt.Errorf("PVC %s has an invalid storage path", claim.Name)
		}
		*volume = Volume{Claim: claim, Path: filepath.Clean(path)}
	}
	for role, volume := range volumes {
		if volume.Claim.Name == "" {
			return nil, fmt.Errorf("%w: missing %s PVC for user %s", ErrNotReady, role, user)
		}
	}
	layout.NodeName = layout.Userspace.Claim.Annotations[NodeAnnotation]
	if layout.NodeName == "" {
		return nil, fmt.Errorf("%w: storage node for user %s is missing", ErrNotReady, user)
	}
	return layout, nil
}

func Read(ctx context.Context, c client.Client, user string) (*Layout, error) {
	var claims corev1.PersistentVolumeClaimList
	if err := c.List(ctx, &claims, client.InNamespace(Namespace(user)), client.HasLabels{RoleLabel}); err != nil {
		return nil, err
	}
	return FromClaims(user, claims.Items)
}

func Load(ctx context.Context, c kubernetes.Interface, user string) (*Layout, error) {
	claims, err := c.CoreV1().PersistentVolumeClaims(Namespace(user)).List(ctx, metav1.ListOptions{LabelSelector: RoleLabel})
	if err != nil {
		return nil, err
	}
	return FromClaims(user, claims.Items)
}

func (l *Layout) Ready() bool {
	for _, volume := range []Volume{l.Userspace, l.AppCache, l.DBData} {
		if volume.Claim.Status.Phase != corev1.ClaimBound {
			return false
		}
		var initialized map[string]bool
		if json.Unmarshal([]byte(volume.Claim.Annotations[InitializedNodesAnnotation]), &initialized) != nil || !initialized[l.NodeName] {
			return false
		}
	}
	return true
}

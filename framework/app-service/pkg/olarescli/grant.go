package olarescli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

const (
	// CredentialSecretName is the Secret the pod webhook mounts. It lives in
	// the app's own namespace; uninstall revokes by presenting the same
	// refresh_token that the mount exposes.
	CredentialSecretName = "olares-cli-credential"

	// CredentialMountPath is where the Secret is mounted, read-only.
	CredentialMountPath = "/olares/credentials"
	// CacheMountPath is a writable emptyDir next to it: olares-cli needs
	// somewhere to keep refreshed access tokens, and the credential mount is
	// read-only and owned by the platform.
	CacheMountPath = "/olares/cache"

	// KeyCredentialFile is the only file under the mount: a JSON document
	// with refreshToken and olaresId.
	KeyCredentialFile = "credential.json"

	// managedByLabel marks the Secret so an operator can tell at a glance
	// which credentials the platform minted.
	managedByLabel = "app.kubernetes.io/managed-by"
	managedByValue = "app-service"
	appLabel       = "applications.app.bytetrade.io/name"
)

// Store reads and writes the credential Secret that backs a grant.
type Store struct {
	client kubernetes.Interface
	cli    *Client
}

func NewStore(client kubernetes.Interface, cli *Client) *Store {
	return &Store{client: client, cli: cli}
}

// EnsureCredential returns the app's existing grant Secret, minting one if
// there is none. The Secret is written into the app's own namespace so the
// pod webhook can mount it directly.
//
// Reusing the existing grant matters: install is a retry path, and a fresh
// token on every attempt would leave the earlier ones valid for a decade with
// nothing left pointing at them.
func (s *Store) EnsureCredential(ctx context.Context, app, owner, olaresID, appNamespace string) (*Grant, error) {
	olaresID = strings.TrimSpace(olaresID)
	if olaresID == "" {
		return nil, fmt.Errorf("olares-cli credential: empty olaresId for user %s", owner)
	}

	secrets := s.client.CoreV1().Secrets(appNamespace)

	existing, err := secrets.Get(ctx, CredentialSecretName, metav1.GetOptions{})
	replaceIncomplete := false
	if err == nil {
		if grant, ok := grantFromSecret(existing); ok {
			return grant, nil
		}
		klog.Warningf("olares-cli credential secret %s/%s is incomplete, re-deriving",
			appNamespace, CredentialSecretName)
		replaceIncomplete = true
	} else if !apierrors.IsNotFound(err) {
		return nil, err
	}

	grant, err := s.cli.Derive(ctx, owner, fmt.Sprintf("app:%s:%s", app, owner))
	if err != nil {
		return nil, err
	}
	grant.OlaresID = olaresID

	if err = s.writeSecret(ctx, existing, app, appNamespace, grant, replaceIncomplete); err != nil {
		s.revokeQuietly(ctx, grant.RefreshToken)
		return nil, err
	}
	return grant, nil
}

func (s *Store) writeSecret(ctx context.Context, existing *corev1.Secret, app, appNamespace string, grant *Grant, update bool) error {
	raw, err := encodeCredentialFile(grant)
	if err != nil {
		return err
	}
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      CredentialSecretName,
			Namespace: appNamespace,
			Labels: map[string]string{
				managedByLabel: managedByValue,
				appLabel:       app,
			},
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			KeyCredentialFile: raw,
		},
	}

	secrets := s.client.CoreV1().Secrets(appNamespace)
	if update {
		secret.ObjectMeta.ResourceVersion = existing.ResourceVersion
		_, err = secrets.Update(ctx, secret, metav1.UpdateOptions{})
		return err
	}

	if _, err = secrets.Create(ctx, secret, metav1.CreateOptions{}); err == nil {
		return nil
	}
	if !apierrors.IsAlreadyExists(err) {
		return err
	}
	_, err = secrets.Update(ctx, secret, metav1.UpdateOptions{})
	return err
}

// LoadCredential reads the app's credential Secret. A missing Secret is not
// an error: uninstall runs for apps that never had a grant.
func (s *Store) LoadCredential(ctx context.Context, appNamespace string) (*Grant, bool, error) {
	secret, err := s.client.CoreV1().Secrets(appNamespace).
		Get(ctx, CredentialSecretName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	grant, ok := grantFromSecret(secret)
	if !ok {
		return nil, false, fmt.Errorf("olares-cli credential secret %s/%s is incomplete",
			secret.Namespace, secret.Name)
	}
	return grant, true, nil
}

// Release revokes the grant in lldap and deletes the credential Secret. It
// reports the first error but always attempts every step, so a failure to
// reach lldap still cleans up the Kubernetes side.
func (s *Store) Release(ctx context.Context, app, owner, appNamespace string) error {
	var firstErr error
	fail := func(err error) {
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}

	if appNamespace == "" {
		return firstErr
	}

	grant, found, err := s.LoadCredential(ctx, appNamespace)
	fail(err)
	if found {
		fail(s.cli.Revoke(ctx, grant.RefreshToken))
	}

	err = s.client.CoreV1().Secrets(appNamespace).
		Delete(ctx, CredentialSecretName, metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		fail(err)
	}
	return firstErr
}

func (s *Store) revokeQuietly(ctx context.Context, refreshToken string) {
	if err := s.cli.Revoke(ctx, refreshToken); err != nil {
		klog.Errorf("Failed to revoke orphaned olares-cli grant: %v", err)
	}
}

func encodeCredentialFile(grant *Grant) ([]byte, error) {
	return json.Marshal(struct {
		RefreshToken string `json:"refreshToken"`
		OlaresID     string `json:"olaresId"`
	}{
		RefreshToken: grant.RefreshToken,
		OlaresID:     grant.OlaresID,
	})
}

func grantFromSecret(secret *corev1.Secret) (*Grant, bool) {
	raw, ok := secret.Data[KeyCredentialFile]
	if !ok {
		return nil, false
	}
	var file struct {
		RefreshToken string `json:"refreshToken"`
		OlaresID     string `json:"olaresId"`
	}
	if err := json.Unmarshal(raw, &file); err != nil {
		return nil, false
	}
	token := file.RefreshToken
	if token == "" {
		return nil, false
	}
	if file.OlaresID == "" {
		return nil, false
	}
	return &Grant{
		RefreshToken: token,
		OlaresID:     file.OlaresID,
	}, true
}

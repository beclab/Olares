package network

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"bytetrade.io/web3os/bfl/pkg/constants"
	sysv1 "github.com/beclab/api/api/sys.bytetrade.io/v1alpha1"
	iamv1 "github.com/beclab/api/iam/v1alpha2"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var ErrAlreadyBound = errors.New("the default domain is already bound and cannot be changed")

type Store struct{ Client client.Client }

func (s *Store) Binding(ctx context.Context) (DomainBinding, error) {
	var terminus sysv1.Terminus
	if err := s.Client.Get(ctx, types.NamespacedName{Name: "terminus"}, &terminus); err != nil {
		return DomainBinding{}, err
	}
	domain := terminus.Spec.Settings[DomainKey]
	if domain == "olares.local" {
		domain = ""
	}
	return DomainBinding{Domain: domain, Username: terminus.Annotations[DomainUserAnnotation]}, nil
}

func (s *Store) RequireAdmin(ctx context.Context, username string) error {
	var user iamv1.User
	if username == "" {
		return fmt.Errorf("request user is required")
	}
	if err := s.Client.Get(ctx, types.NamespacedName{Name: username}, &user); err != nil {
		return err
	}
	if user.DeletionTimestamp != nil || user.Annotations[constants.UserAnnotationOwnerRole] != constants.RoleAdmin {
		return fmt.Errorf("only an admin may configure the default network")
	}
	return nil
}

func (s *Store) Bind(ctx context.Context, username, domain, jwsSignature string) error {
	if err := s.RequireAdmin(ctx, username); err != nil {
		return err
	}
	normalized, _, err := NormalizeDomain(domain)
	if err != nil {
		return err
	}
	jwsSignature = strings.TrimSpace(jwsSignature)
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		var terminus sysv1.Terminus
		if err := s.Client.Get(ctx, types.NamespacedName{Name: "terminus"}, &terminus); err != nil {
			return err
		}
		current := terminus.Spec.Settings[DomainKey]
		if current != "" && current != "olares.local" {
			if current != normalized {
				return ErrAlreadyBound
			}
			if jwsSignature == "" {
				return nil
			}
		}
		// Cloud credentials belong to this domain binding, independently of the
		// local identity JWS saved during user activation.
		if jwsSignature != "" {
			if err := s.PutSecret(ctx, DomainCredentialsName, map[string][]byte{
				"domain": []byte(normalized), "username": []byte(username), "jws": []byte(jwsSignature),
			}); err != nil {
				return err
			}
		}
		original := terminus.DeepCopy()
		if terminus.Spec.Settings == nil {
			terminus.Spec.Settings = map[string]string{}
		}
		if terminus.Annotations == nil {
			terminus.Annotations = map[string]string{}
		}
		terminus.Spec.Settings[DomainKey] = normalized
		terminus.Annotations[DomainUserAnnotation] = username
		// The pinned API may not know every Terminus field. Patch only this
		// binding, retaining localName and other fields omitted by typed reads.
		return s.Client.Patch(ctx, &terminus, client.MergeFromWithOptions(original, client.MergeFromWithOptimisticLock{}))
	})
}

func (s *Store) Config(ctx context.Context) (*corev1.ConfigMap, ProxySpec, Status, error) {
	cm := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: ConfigName, Namespace: Namespace}, Data: map[string]string{}}
	spec := ProxySpec{Config: ProxyConfig{Mode: ModeNone}}
	status := Status{Domain: DomainStatus{Phase: "unbound"}, Proxy: ProxyStatus{Phase: "not_configured"}}
	err := s.Client.Get(ctx, client.ObjectKeyFromObject(cm), cm)
	if err != nil && !apierrors.IsNotFound(err) {
		return nil, spec, status, err
	}
	if apierrors.IsNotFound(err) {
		return cm, spec, status, nil
	}
	if raw := cm.Data[ConfigKey]; raw != "" {
		if err := json.Unmarshal([]byte(raw), &spec); err != nil {
			return nil, spec, status, err
		}
	}
	if raw := cm.Data[StatusKey]; raw != "" {
		if err := json.Unmarshal([]byte(raw), &status); err != nil {
			return nil, spec, status, err
		}
	}
	return cm, spec, status, nil
}

func (s *Store) DomainJWS(ctx context.Context, binding DomainBinding) (string, error) {
	var secret corev1.Secret
	if err := s.Client.Get(ctx, types.NamespacedName{Name: DomainCredentialsName, Namespace: Namespace}, &secret); err != nil {
		if apierrors.IsNotFound(err) {
			return "", fmt.Errorf("default domain cloud credentials are missing; submit jws_signature when binding the domain")
		}
		return "", err
	}
	if string(secret.Data["domain"]) != binding.Domain || string(secret.Data["username"]) != binding.Username {
		return "", fmt.Errorf("cloud credentials do not match the default domain binding")
	}
	jws := strings.TrimSpace(string(secret.Data["jws"]))
	if jws == "" {
		return "", fmt.Errorf("default domain cloud credentials are missing; submit jws_signature when binding the domain")
	}
	return jws, nil
}

// The selected authentication method determines how this credential is used.
// A token credential must never be reused as a JWS, or vice versa.
func (s *Store) AuthToken(ctx context.Context, method string) (string, error) {
	if method != "token" && method != "jws" {
		return "", nil
	}
	var secret corev1.Secret
	err := s.Client.Get(ctx, types.NamespacedName{Name: ProxyCredentialsName, Namespace: Namespace}, &secret)
	if apierrors.IsNotFound(err) {
		return "", nil
	}
	return string(secret.Data[method]), err
}

func (s *Store) Entrance(ctx context.Context) (Entrance, error) {
	cm, _, _, err := s.Config(ctx)
	if err != nil {
		return Entrance{}, err
	}
	return entranceFromData(cm.Data), nil
}

func (s *Store) SaveEntrance(ctx context.Context, entrance Entrance) error {
	return s.mutateConfig(ctx, func(cm *corev1.ConfigMap, _ *ProxySpec, _ *Status) error {
		if cm.Data == nil {
			cm.Data = map[string]string{}
		}
		cm.Data[LocalIPKey] = entrance.LocalIP
		cm.Data[L4ReadyKey] = strconv.FormatBool(entrance.L4Ready)
		cm.Data[LocalDNSIPKey] = entrance.LocalIP
		if natIP := cm.Data[NATGatewayIPKey]; natIP != "" {
			cm.Data[LocalDNSIPKey] = natIP
		}
		cm.Data[PublicIPKey] = entrance.PublicIP
		cm.Data[PublicIPErrorKey] = entrance.PublicIPError
		return nil
	})
}

func (s *Store) SetProxy(ctx context.Context, username string, conf ProxyConfig) error {
	if err := s.RequireAdmin(ctx, username); err != nil {
		return err
	}
	if err := conf.Normalize(); err != nil {
		return err
	}
	token := conf.FRPAuthToken
	conf.FRPAuthToken = ""
	if conf.Mode == ModeFRP {
		if token == "" {
			var err error
			token, err = s.AuthToken(ctx, conf.FRPAuthMethod)
			if err != nil {
				return err
			}
		}
		if token == "" && conf.FRPAuthMethod == "token" {
			return fmt.Errorf("frp_auth_token is required for token authentication")
		}
		if token != "" {
			if err := s.PutSecret(ctx, ProxyCredentialsName, map[string][]byte{conf.FRPAuthMethod: []byte(token)}); err != nil {
				return err
			}
		}
	}
	return s.mutateConfig(ctx, func(cm *corev1.ConfigMap, spec *ProxySpec, status *Status) error {
		if reflect.DeepEqual(spec.Config, conf) && spec.Revision != 0 && token == "" {
			return nil
		}
		spec.Config = conf
		spec.Username = username
		spec.Revision++
		status.Proxy = ProxyStatus{Phase: "pending"}
		if status.Domain.Phase != "unbound" {
			status.Domain.Phase = "pending"
		}
		return nil
	})
}

func (s *Store) SaveDomainStatus(ctx context.Context, revision int64, status DomainStatus) error {
	return s.mutateConfig(ctx, func(cm *corev1.ConfigMap, spec *ProxySpec, current *Status) error {
		if spec.Revision != revision {
			return fmt.Errorf("network configuration changed; reconcile again")
		}
		if !reflect.DeepEqual(current.Domain, status) {
			current.Domain = status
			current.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
		}
		return nil
	})
}

func (s *Store) SaveProxyStatus(ctx context.Context, revision int64, status ProxyStatus) error {
	return s.mutateConfig(ctx, func(_ *corev1.ConfigMap, spec *ProxySpec, current *Status) error {
		if spec.Revision != revision {
			return fmt.Errorf("network configuration changed; reconcile again")
		}
		if !reflect.DeepEqual(current.Proxy, status) {
			current.Proxy = status
			current.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
		}
		return nil
	})
}

func (s *Store) mutateConfig(ctx context.Context, mutate func(*corev1.ConfigMap, *ProxySpec, *Status) error) error {
	return retry.OnError(retry.DefaultRetry, func(err error) bool { return apierrors.IsConflict(err) || apierrors.IsAlreadyExists(err) }, func() error {
		cm, spec, status, err := s.Config(ctx)
		if err != nil {
			return err
		}
		before := cm.DeepCopy()
		if err := mutate(cm, &spec, &status); err != nil {
			return err
		}
		if cm.Data == nil {
			cm.Data = map[string]string{}
		}
		specJSON, err := json.Marshal(spec)
		if err != nil {
			return err
		}
		statusJSON, err := json.Marshal(status)
		if err != nil {
			return err
		}
		cm.Data[ConfigKey], cm.Data[StatusKey] = string(specJSON), string(statusJSON)
		if cm.ResourceVersion == "" {
			return s.Client.Create(ctx, cm)
		}
		if reflect.DeepEqual(before.Data, cm.Data) {
			return nil
		}
		return s.Client.Update(ctx, cm)
	})
}

func (s *Store) PutSecret(ctx context.Context, name string, data map[string][]byte) error {
	secretType := corev1.SecretTypeOpaque
	if name == CertificateName {
		secretType = corev1.SecretTypeTLS
	}
	return retry.OnError(retry.DefaultRetry, func(err error) bool { return apierrors.IsConflict(err) || apierrors.IsAlreadyExists(err) }, func() error {
		var secret corev1.Secret
		err := s.Client.Get(ctx, types.NamespacedName{Name: name, Namespace: Namespace}, &secret)
		if err != nil && !apierrors.IsNotFound(err) {
			return err
		}
		if apierrors.IsNotFound(err) {
			secret = corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: Namespace}, Type: secretType, Data: data}
			return s.Client.Create(ctx, &secret)
		}
		if reflect.DeepEqual(secret.Data, data) && secret.Type == secretType {
			return nil
		}
		secret.Data = data
		secret.Type = secretType
		return s.Client.Update(ctx, &secret)
	})
}

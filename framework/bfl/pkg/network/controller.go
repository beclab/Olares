package network

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net"
	"reflect"
	"time"

	"bytetrade.io/web3os/bfl/pkg/constants"
	"bytetrade.io/web3os/bfl/pkg/utils/certmanager"
	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	sysv1 "github.com/beclab/api/api/sys.bytetrade.io/v1alpha1"
	iamv1 "github.com/beclab/api/iam/v1alpha2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type Controller struct {
	Store *Store
	Cloud func(context.Context, constants.TerminusName, string) certmanager.Interface
}

func (c *Controller) SetupWithManager(mgr ctrl.Manager) error {
	return setupNetworkController(mgr, "default-network", c, true)
}

func setupNetworkController(mgr ctrl.Manager, name string, reconciler reconcile.Reconciler, watchProxyStatus bool) error {
	queue := handler.EnqueueRequestsFromMapFunc(func(context.Context, client.Object) []reconcile.Request {
		return []reconcile.Request{{NamespacedName: types.NamespacedName{Name: ConfigName}}}
	})
	relevant := func(obj client.Object) bool {
		switch v := obj.(type) {
		case *sysv1.Terminus:
			return v.Name == "terminus"
		case *corev1.ConfigMap:
			return v.Namespace == Namespace && v.Name == ConfigName
		case *corev1.Secret:
			return v.Namespace == Namespace && (v.Name == DomainCredentialsName || v.Name == ProxyCredentialsName || v.Name == CertificateName)
		case *appsv1.Deployment:
			return v.Namespace == Namespace && v.Name == AgentName
		case *corev1.Pod:
			return v.Namespace == Namespace && (v.Labels["app"] == "l4-bfl-proxy" || v.Labels["app"] == AgentName)
		case *iamv1.User, *appv1.Application:
			return true
		}
		return false
	}
	p := predicate.Funcs{
		CreateFunc:  func(e event.CreateEvent) bool { return relevant(e.Object) },
		DeleteFunc:  func(e event.DeleteEvent) bool { return relevant(e.Object) },
		GenericFunc: func(e event.GenericEvent) bool { return relevant(e.Object) },
		UpdateFunc: func(e event.UpdateEvent) bool {
			if !relevant(e.ObjectNew) {
				return false
			}
			switch current := e.ObjectNew.(type) {
			case *corev1.ConfigMap:
				old := e.ObjectOld.(*corev1.ConfigMap)
				for _, key := range []string{ConfigKey, LocalIPKey, L4ReadyKey, PublicIPKey, PublicIPErrorKey} {
					if old.Data[key] != current.Data[key] {
						return true
					}
				}
				if watchProxyStatus {
					var previous, next Status
					if json.Unmarshal([]byte(old.Data[StatusKey]), &previous) != nil || json.Unmarshal([]byte(current.Data[StatusKey]), &next) != nil {
						return old.Data[StatusKey] != current.Data[StatusKey]
					}
					return !reflect.DeepEqual(previous.Proxy, next.Proxy)
				}
				return false
			case *iamv1.User:
				old := e.ObjectOld.(*iamv1.User)
				return old.Annotations[constants.UserCertManagerJWSToken] != current.Annotations[constants.UserCertManagerJWSToken] || !reflect.DeepEqual(old.DeletionTimestamp, current.DeletionTimestamp)
			case *appv1.Application:
				return !reflect.DeepEqual(e.ObjectOld.(*appv1.Application).Spec, current.Spec)
			case *sysv1.Terminus:
				old := e.ObjectOld.(*sysv1.Terminus)
				return old.Spec.Settings[DomainKey] != current.Spec.Settings[DomainKey] || old.Annotations[DomainUserAnnotation] != current.Annotations[DomainUserAnnotation]
			}
			return true
		},
	}
	return ctrl.NewControllerManagedBy(mgr).Named(name).
		Watches(&sysv1.Terminus{}, queue, builder.WithPredicates(p)).
		Watches(&corev1.ConfigMap{}, queue, builder.WithPredicates(p)).
		Watches(&corev1.Secret{}, queue, builder.WithPredicates(p)).
		Watches(&iamv1.User{}, queue, builder.WithPredicates(p)).
		Watches(&appsv1.Deployment{}, queue, builder.WithPredicates(p)).
		Watches(&corev1.Pod{}, queue, builder.WithPredicates(p)).
		Watches(&appv1.Application{}, queue, builder.WithPredicates(p)).Complete(reconciler)
}

func (c *Controller) Reconcile(ctx context.Context, _ ctrl.Request) (ctrl.Result, error) {
	binding, err := c.Store.Binding(ctx)
	if err != nil {
		return ctrl.Result{}, err
	}
	_, spec, status, err := c.Store.Config(ctx)
	if err != nil {
		return ctrl.Result{}, err
	}
	err = c.configure(ctx, binding, spec, &status)
	if err != nil {
		status.Domain.Phase, status.Domain.Message = "failed", err.Error()
	}
	if saveErr := c.Store.SaveDomainStatus(ctx, spec.Revision, status.Domain); saveErr != nil {
		return ctrl.Result{}, saveErr
	}
	// Also retry cloud failures, certificate renewal and changes missed while
	// BFL was unavailable. These tasks never depend on wizard status.
	if err != nil {
		return ctrl.Result{RequeueAfter: 15 * time.Second}, nil
	}
	return ctrl.Result{RequeueAfter: time.Minute}, nil
}

func (c *Controller) configure(ctx context.Context, binding DomainBinding, spec ProxySpec, status *Status) error {
	if binding.Domain == "" {
		status.Domain = DomainStatus{Phase: "unbound"}
		return nil
	}
	jws, err := c.bindingJWS(ctx, binding)
	if err != nil {
		return err
	}
	cloudName, err := certmanager.CloudNameFromJWS(jws)
	if err != nil {
		return err
	}
	cloud := c.Cloud
	if cloud == nil {
		cloud = certmanager.NewCertManagerWithCredentials
	}
	cm := cloud(ctx, cloudName, jws)
	if err := c.ensureCertificate(ctx, binding, spec.Revision, status, cm); err != nil {
		return err
	}
	_, latestSpec, latestStatus, err := c.Store.Config(ctx)
	if err != nil {
		return err
	}
	proxy := latestStatus.Proxy
	if latestSpec.Revision != spec.Revision || proxy.ObservedRevision != spec.Revision {
		status.Domain.Phase, status.Domain.Message = "waiting_dns_target", "Waiting for the current proxy configuration"
		return nil
	}
	if proxy.Phase != "ready" {
		status.Domain.Phase, status.Domain.Message = "waiting_dns_target", proxy.Message
		if spec.Config.Mode == ModeNone && status.Domain.DNSConfigured {
			if err := cm.DeleteDNSRecord(); err != nil {
				return fmt.Errorf("remove DNS record: %w", err)
			}
			status.Domain.DNSConfigured, status.Domain.DNSTarget = false, ""
		}
		return nil
	}
	if !status.Domain.DNSConfigured || status.Domain.DNSTarget != proxy.DNSTarget {
		status.Domain.Phase, status.Domain.Message = "configuring_dns", ""
		if err := c.Store.SaveDomainStatus(ctx, spec.Revision, status.Domain); err != nil {
			return err
		}
		if net.ParseIP(proxy.DNSTarget) != nil {
			err = cm.AddDNSRecord(&proxy.DNSTarget, nil)
		} else {
			err = cm.AddDNSRecord(nil, &proxy.DNSTarget)
		}
		if err != nil {
			return fmt.Errorf("configure DNS: %w", err)
		}
		status.Domain.DNSTarget, status.Domain.DNSConfigured = proxy.DNSTarget, true
	}
	status.Domain.Phase, status.Domain.Message = "completed", ""
	return nil
}

func (c *Controller) ensureCertificate(ctx context.Context, binding DomainBinding, revision int64, status *Status, cm certmanager.Interface) error {
	var secret corev1.Secret
	err := c.Store.Client.Get(ctx, types.NamespacedName{Name: CertificateName, Namespace: Namespace}, &secret)
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	cert, certErr := certificate(secret.Data[corev1.TLSCertKey], secret.Data[corev1.TLSPrivateKeyKey], binding.Domain)
	if certErr == nil && cert.NotAfter.After(time.Now().Add(7*24*time.Hour)) {
		status.Domain.CertificateReady, status.Domain.CertificateExpiresAt = true, cert.NotAfter.UTC().Format(time.RFC3339)
		return nil
	}
	status.Domain.Phase, status.Domain.Message = "configuring_certificate", ""
	status.Domain.CertificateReady = certErr == nil
	if err := c.Store.SaveDomainStatus(ctx, revision, status.Domain); err != nil {
		return err
	}
	if err := cm.GenerateCert(); err != nil {
		return fmt.Errorf("generate certificate: %w", err)
	}
	material, err := cm.DownloadCert()
	if err != nil {
		return fmt.Errorf("download certificate: %w", err)
	}
	if material == nil || material.Zone != binding.Domain {
		return fmt.Errorf("the downloaded certificate zone does not match the default domain")
	}
	cert, err = certificate([]byte(material.Cert), []byte(material.Key), binding.Domain)
	if err != nil {
		return fmt.Errorf("validate certificate: %w", err)
	}
	if err := c.Store.PutSecret(ctx, CertificateName, map[string][]byte{corev1.TLSCertKey: []byte(material.Cert), corev1.TLSPrivateKeyKey: []byte(material.Key)}); err != nil {
		return err
	}
	status.Domain.CertificateReady, status.Domain.CertificateExpiresAt = true, cert.NotAfter.UTC().Format(time.RFC3339)
	return nil
}

func (c *Controller) bindingJWS(ctx context.Context, binding DomainBinding) (string, error) {
	return c.Store.DomainJWS(ctx, binding)
}

func certificate(certPEM, keyPEM []byte, domain string) (*x509.Certificate, error) {
	pair, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}
	cert, err := x509.ParseCertificate(pair.Certificate[0])
	if err != nil {
		return nil, err
	}
	if time.Now().Before(cert.NotBefore) || !time.Now().Before(cert.NotAfter) {
		return nil, fmt.Errorf("certificate is not currently valid")
	}
	if err := cert.VerifyHostname(domain); err != nil {
		return nil, err
	}
	if err := cert.VerifyHostname("app." + domain); err != nil {
		return nil, err
	}
	return cert, nil
}

package network_activation

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"bytetrade.io/web3os/bfl/internal/log"
	"bytetrade.io/web3os/bfl/pkg/apis/iam/v1alpha1/operator"
	settingsv1alpha1 "bytetrade.io/web3os/bfl/pkg/apis/settings/v1alpha1"
	v1alpha1client "bytetrade.io/web3os/bfl/pkg/client/clientset/v1alpha1"
	"bytetrade.io/web3os/bfl/pkg/constants"
	"bytetrade.io/web3os/bfl/pkg/utils"
	"bytetrade.io/web3os/bfl/pkg/utils/certmanager"
	"bytetrade.io/web3os/bfl/pkg/watchers"

	iamV1alpha2 "github.com/beclab/api/iam/v1alpha2"
	batchV1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	batchapply "k8s.io/client-go/applyconfigurations/batch/v1"
	applyCorev1 "k8s.io/client-go/applyconfigurations/core/v1"
	applyMetav1 "k8s.io/client-go/applyconfigurations/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"
)

var GVR = schema.GroupVersionResource{Group: iamV1alpha2.SchemeGroupVersion.Group, Version: iamV1alpha2.SchemeGroupVersion.Version, Resource: iamV1alpha2.ResourcesPluralUser}

type Subscriber struct {
	*watchers.Watchers
	client v1alpha1client.ClientInterface
}

func NewSubscriber(w *watchers.Watchers) *Subscriber {
	return &Subscriber{Watchers: w}
}

func (s *Subscriber) WithKubeConfig(config *rest.Config) *Subscriber {
	c, _ := v1alpha1client.NewKubeClient(config)
	s.client = c
	return s
}

func (s *Subscriber) Handler() cache.ResourceEventHandler {
	handle := func(obj interface{}) {
		s.Watchers.Enqueue(watchers.EnqueueObj{Subscribe: s, Obj: obj, Action: watchers.UPDATE})
	}
	return cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			u, ok := obj.(*iamV1alpha2.User)
			if !ok || u == nil {
				klog.Error("not user resource, invalid obj")
				return false
			}
			if u.Name != constants.Username {
				return false
			}
			status := u.Annotations[constants.UserTerminusWizardStatus]
			return status == string(constants.NetworkActivating)
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc:    handle,
			UpdateFunc: func(_, newObj interface{}) { handle(newObj) },
			DeleteFunc: func(_ interface{}) {},
		},
	}
}

func (s *Subscriber) Do(ctx context.Context, obj interface{}, _ watchers.Action) error {
	u, ok := obj.(*iamV1alpha2.User)
	if !ok {
		return fmt.Errorf("invalid object")
	}

	userOp, err := operator.NewUserOperatorWithContext(ctx)
	if err != nil {
		return fmt.Errorf("new user operator: %w", err)
	}
	current, err := userOp.GetUser(u.Name)
	if err != nil {
		return fmt.Errorf("get user failed: %w", err)
	}

	terminusName := current.Annotations[constants.UserAnnotationTerminusNameKey]
	// this should never happen
	if terminusName == "" {
		// terminal failure: update status only
		return s.markFailed(ctx, current, "no terminus name found")
	}

	if err := s.ensureL4ProxyDeployment(ctx); err != nil {
		return err
	}

	proxyType := current.Annotations[constants.UserAnnotationReverseProxyType]
	if proxyType == "" {
		return fmt.Errorf("reverse proxy type not set yet")
	}
	if proxyType == constants.ReverseProxyTypeFRP || proxyType == constants.ReverseProxyTypeCloudflare {
		if err := s.waitDeploymentReadyOnce(ctx, constants.Namespace, settingsv1alpha1.ReverseProxyAgentDeploymentName, settingsv1alpha1.ReverseProxyAgentDeploymentReplicas); err != nil {
			return err
		}
	}

	cm := certmanager.NewCertManager(constants.TerminusName(terminusName))
	if err := cm.GenerateCert(); err != nil {
		return err
	}
	c, err := cm.DownloadCert()
	if err != nil {
		return err
	}

	if err := s.applyRenewCertCronJob(ctx, c.ExpiredAt); err != nil {
		return err
	}

	if err := s.applySSLConfigMap(ctx, c); err != nil {
		return err
	}

	if err := s.markSuccess(ctx, current, c.Zone); err != nil {
		return err
	}

	log.Info("Network activation completed, moved wizard status to wait_reset_password")
	return nil
}

func (s *Subscriber) markSuccess(ctx context.Context, user *iamV1alpha2.User, zone string) error {
	userOp, err := operator.NewUserOperatorWithContext(ctx)
	if err != nil {
		return err
	}
	return userOp.UpdateUser(user, []func(*iamV1alpha2.User){
		func(u *iamV1alpha2.User) {
			if u.Annotations == nil {
				u.Annotations = map[string]string{}
			}
			if zone != "" {
				u.Annotations[constants.UserAnnotationZoneKey] = zone
				u.Annotations[constants.UserAnnotationIsEphemeral] = "false"
			}
			u.Annotations[constants.UserTerminusWizardStatus] = string(constants.WaitResetPassword)
			u.Annotations[constants.UserTerminusWizardError] = ""
		},
	})
}

func (s *Subscriber) markFailed(ctx context.Context, user *iamV1alpha2.User, errorMsg string) error {
	userOp, err := operator.NewUserOperatorWithContext(ctx)
	if err != nil {
		return err
	}
	return userOp.UpdateUser(user, []func(*iamV1alpha2.User){
		func(u *iamV1alpha2.User) {
			if u.Annotations == nil {
				u.Annotations = map[string]string{}
			}
			u.Annotations[constants.UserTerminusWizardStatus] = string(constants.NetworkActivateFailed)
			u.Annotations[constants.UserTerminusWizardError] = errorMsg
		},
	})
}

func (s *Subscriber) ensureL4ProxyDeployment(ctx context.Context) error {
	namespace := utils.EnvOrDefault("L4_PROXY_NAMESPACE", constants.OSSystemNamespace)
	serviceAccount := utils.EnvOrDefault("L4_PROXY_SERVICE_ACCOUNT", constants.L4ProxyServiceAccountName)
	portStr := utils.EnvOrDefault("L4_PROXY_LISTEN", constants.L4ListenSSLPort)
	port, _ := strconv.Atoi(portStr)

	deploy, err := s.client.Kubernetes().AppsV1().Deployments(namespace).Get(ctx, settingsv1alpha1.L4ProxyDeploymentName, metav1.GetOptions{})
	if err != nil || deploy == nil {
		// apply and requeue
		apply := settingsv1alpha1.NewL4ProxyDeploymentApplyConfiguration(namespace, serviceAccount, port)
		if _, e := s.client.Kubernetes().AppsV1().Deployments(namespace).Apply(ctx, &apply, metav1.ApplyOptions{Force: true, FieldManager: constants.ApplyPatchFieldManager}); e != nil {
			return fmt.Errorf("apply l4 proxy failed: %w", e)
		}
		return fmt.Errorf("l4 proxy applied, waiting for ready")
	}
	return s.waitDeploymentReadyOnce(ctx, namespace, settingsv1alpha1.L4ProxyDeploymentName, settingsv1alpha1.L4ProxyDeploymentReplicas)
}

func (s *Subscriber) waitDeploymentReadyOnce(ctx context.Context, ns, name string, replicas int32) error {
	dep, err := s.client.Kubernetes().AppsV1().Deployments(ns).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	if dep == nil || dep.Status.AvailableReplicas < replicas {
		return fmt.Errorf("deployment %s/%s not ready", ns, name)
	}
	return nil
}

func (s *Subscriber) applySSLConfigMap(ctx context.Context, c *certmanager.ResponseCert) error {
	cmApply := &applyCorev1.ConfigMapApplyConfiguration{
		TypeMetaApplyConfiguration: applyMetav1.TypeMetaApplyConfiguration{
			Kind:       pointer.String("ConfigMap"),
			APIVersion: pointer.String(corev1.SchemeGroupVersion.String()),
		},
		ObjectMetaApplyConfiguration: &applyMetav1.ObjectMetaApplyConfiguration{
			Name:      pointer.String(constants.NameSSLConfigMapName),
			Namespace: pointer.String(constants.Namespace),
		},
		Data: map[string]string{
			"zone":       c.Zone,
			"cert":       c.Cert,
			"key":        c.Key,
			"expired_at": c.ExpiredAt,
		},
	}
	_, err := s.client.Kubernetes().CoreV1().ConfigMaps(constants.Namespace).Apply(ctx, cmApply, metav1.ApplyOptions{FieldManager: constants.ApplyPatchFieldManager})
	return err
}

func (s *Subscriber) applyRenewCertCronJob(ctx context.Context, expiredAt string) error {
	forbidConcurrent := batchV1.ForbidConcurrent
	restartOnFailure := corev1.RestartPolicyOnFailure

	parsedTime, err := time.Parse(certmanager.CertExpiredDateTimeLayout, expiredAt)
	if err != nil {
		return fmt.Errorf("parse expired time err, %v", err)
	}

	expiredTime := parsedTime.AddDate(0, 0, certmanager.DefaultAheadRenewalCertDays)
	schedule := fmt.Sprintf(certmanager.ReDownloadCertCronJobScheduleFormat, expiredTime.Minute(), expiredTime.Hour(), expiredTime.Day(), int(expiredTime.Month()))

	cronjob := batchapply.CronJob(certmanager.ReDownloadCertCronJobName, constants.Namespace)
	cronjob.Spec = &batchapply.CronJobSpecApplyConfiguration{
		ConcurrencyPolicy:       &forbidConcurrent,
		Schedule:                pointer.String(schedule),
		StartingDeadlineSeconds: pointer.Int64(3),
	}
	cronjob.Spec.JobTemplate = batchapply.JobTemplateSpec()
	cronjob.Spec.JobTemplate.Spec = batchapply.JobSpec()
	cronjob.Spec.JobTemplate.Spec.Template = applyCorev1.PodTemplateSpec()
	cronjob.Spec.JobTemplate.Spec.Template.Spec = applyCorev1.PodSpec()
	cronjob.Spec.JobTemplate.Spec.Template.Spec.RestartPolicy = &restartOnFailure
	cronjob.Spec.JobTemplate.Spec.Template.Spec.Containers = []applyCorev1.ContainerApplyConfiguration{{
		Name:  pointer.String("trigger"),
		Image: pointer.String("busybox:1.28"),
		Command: []string{
			"wget",
			"--header",
			"X-FROM-CRONJOB: true",
			"-qSO - ",
			fmt.Sprintf(certmanager.ReDownloadCertificateAPIFormat, constants.Namespace),
		},
	}}

	_, err = s.client.Kubernetes().BatchV1().CronJobs(constants.Namespace).Apply(ctx, cronjob, metav1.ApplyOptions{FieldManager: constants.ApplyPatchFieldManager})
	return err
}

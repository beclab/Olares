package provider

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	v1alpha1App "bytetrade.io/web3os/bfl/internal/ingress/api/app.bytetrade.io/v1alpha1"
	"bytetrade.io/web3os/bfl/internal/ingress/message"
	"bytetrade.io/web3os/bfl/pkg/constants"

	iamV1alpha2 "github.com/beclab/api/iam/v1alpha2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	toolscache "k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	ctrlcache "sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	mapKey        = "default"
	debounceDelay = 100 * time.Millisecond
	retryDelay    = time.Second
)

type Config struct {
	Username  string
	Namespace string
}

type Provider struct {
	cache      ctrlcache.Cache
	resources  *message.ProviderResources
	cfg        *Config
	synced     int32
	debounceCh chan struct{}
}

func New(c ctrlcache.Cache, resources *message.ProviderResources, cfg *Config) *Provider {
	return &Provider{
		cache:      c,
		resources:  resources,
		cfg:        cfg,
		debounceCh: make(chan struct{}, 1),
	}
}

func (p *Provider) Name() string { return "provider" }

func (p *Provider) SetupWithManager(ctx context.Context) error {
	cmInformer, err := p.cache.GetInformer(ctx, &corev1.ConfigMap{})
	if err != nil {
		return fmt.Errorf("get configmap informer: %w", err)
	}

	appInformer, err := p.cache.GetInformer(ctx, &v1alpha1App.Application{})
	if err != nil {
		return fmt.Errorf("get application informer: %w", err)
	}

	podInformer, err := p.cache.GetInformer(ctx, &corev1.Pod{})
	if err != nil {
		return fmt.Errorf("get pod informer: %w", err)
	}

	userInformer, err := p.cache.GetInformer(ctx, &iamV1alpha2.User{})
	if err != nil {
		return fmt.Errorf("get user informer: %w", err)
	}

	baseHandler := toolscache.ResourceEventHandlerFuncs{
		AddFunc:    func(_ interface{}) { p.notifyChanged() },
		UpdateFunc: func(_, _ interface{}) { p.notifyChanged() },
		DeleteFunc: func(_ interface{}) { p.notifyChanged() },
	}

	if _, err = cmInformer.AddEventHandler(toolscache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			cm, ok := obj.(*corev1.ConfigMap)
			if !ok {
				return false
			}
			return isSSLConfigMap(cm, p.cfg.Namespace) || isCustomDomainCertConfigMap(cm)
		},
		Handler: baseHandler,
	}); err != nil {
		return fmt.Errorf("add configmap event handler: %w", err)
	}

	if _, err = appInformer.AddEventHandler(toolscache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			app, ok := obj.(*v1alpha1App.Application)
			if !ok {
				return false
			}
			return app.Spec.Owner == p.cfg.Username
		},
		Handler: baseHandler,
	}); err != nil {
		return fmt.Errorf("add application event handler: %w", err)
	}

	if _, err = podInformer.AddEventHandler(toolscache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			pod, ok := obj.(*corev1.Pod)
			if !ok {
				return false
			}
			return isFileServerPod(pod)
		},
		Handler: baseHandler,
	}); err != nil {
		return fmt.Errorf("add pod event handler: %w", err)
	}

	if _, err = userInformer.AddEventHandler(toolscache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			_, ok := obj.(*iamV1alpha2.User)
			return ok
		},
		Handler: baseHandler,
	}); err != nil {
		return fmt.Errorf("add user event handler: %w", err)
	}

	klog.Info("provider: informers and event handlers registered...")
	return nil
}

func (p *Provider) Start(ctx context.Context) error {
	atomic.StoreInt32(&p.synced, 1)
	klog.Info("provider: cache synced, publishing initial snapshot")
	if err := p.publishResources(ctx); err != nil {
		klog.Warningf("provider: initial publish failed, will retry: %v", err)
		p.scheduleRetry(ctx)
	}
	p.debounceLoop(ctx)
	klog.Info("provider: stopped")
	return nil
}

func (p *Provider) notifyChanged() {
	if atomic.LoadInt32(&p.synced) == 0 {
		return
	}
	select {
	case p.debounceCh <- struct{}{}:
	default:
	}
}

// NotifyChanged is the exported version of notifyChanged, used as a callback
// by FileserverReconciler to trigger a re-publish after proxy Services are created.
func (p *Provider) NotifyChanged() {
	p.notifyChanged()
}

func (p *Provider) debounceLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-p.debounceCh:
			timer := time.NewTimer(debounceDelay)
		drain:
			for {
				select {
				case <-p.debounceCh:
					if !timer.Stop() {
						select {
						case <-timer.C:
						default:
						}
					}
					timer.Reset(debounceDelay)
				case <-timer.C:
					break drain
				case <-ctx.Done():
					timer.Stop()
					return
				}
			}
			if err := p.publishResources(ctx); err != nil {
				klog.Warningf("provider: publish failed, retrying in %v: %v", retryDelay, err)
				p.scheduleRetry(ctx)
			}
		}
	}
}

func (p *Provider) scheduleRetry(ctx context.Context) {
	go func() {
		select {
		case <-time.After(retryDelay):
			p.notifyChanged()
		case <-ctx.Done():
		}
	}()
}

func (p *Provider) publishResources(ctx context.Context) error {
	snapshot, err := p.buildResources(ctx)
	if err != nil {
		return fmt.Errorf("build resources: %w", err)
	}

	if old, ok := p.resources.Load(mapKey); ok && old.Equal(snapshot) {
		klog.V(4).Info("provider: snapshot unchanged, skipping publish")
		return nil
	}

	p.resources.Store(mapKey, snapshot)
	klog.Infof("provider: published snapshot with %d apps", len(snapshot.Apps))
	return nil
}

func (p *Provider) buildResources(ctx context.Context) (*message.Resources, error) {
	res := &message.Resources{
		UserName: p.cfg.Username,
	}

	sslConfig, err := p.getSSLConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("get ssl config: %w", err)
	}
	res.SSL = sslConfig
	if sslConfig != nil {
		res.UserZone = sslConfig.Zone
		res.IsEphemeralUser = sslConfig.Ephemeral
	}

	language, err := p.getUserLanguage(ctx)
	if err != nil {
		return nil, fmt.Errorf("get user language: %w", err)
	}
	res.Language = language

	apps, err := p.getOwnerApps(ctx)
	if err != nil {
		return nil, fmt.Errorf("get owner apps: %w", err)
	}
	res.Apps = apps

	fsNodes, err := p.getFileserverNodes(ctx)
	if err != nil {
		return nil, fmt.Errorf("get fileserver nodes: %w", err)
	}
	res.FileserverNodes = fsNodes

	certs, err := p.getCustomDomainCerts(ctx)
	if err != nil {
		return nil, fmt.Errorf("get custom domain certs: %w", err)
	}
	res.CustomDomainCerts = certs

	return res, nil
}

func (p *Provider) getSSLConfig(ctx context.Context) (*message.SSLConfig, error) {
	var cm corev1.ConfigMap
	key := client.ObjectKey{
		Namespace: p.cfg.Namespace,
		Name:      constants.NameSSLConfigMapName,
	}
	if err := p.cache.Get(ctx, key, &cm); err != nil {
		if client.IgnoreNotFound(err) == nil {
			klog.V(4).Info("provider: ssl configmap not found")
			return nil, nil
		}
		return nil, fmt.Errorf("get ssl configmap: %w", err)
	}

	if cm.Data == nil {
		return nil, nil
	}

	zone := cm.Data["zone"]
	if zone == "" {
		return nil, nil
	}

	cfg := &message.SSLConfig{
		Zone:     zone,
		CertData: cm.Data["cert"],
		KeyData:  cm.Data["key"],
	}

	if ephStr, ok := cm.Data["ephemeral"]; ok {
		cfg.Ephemeral, _ = strconv.ParseBool(ephStr)
	}

	return cfg, nil
}

func (p *Provider) getUserLanguage(ctx context.Context) (string, error) {
	var userList iamV1alpha2.UserList
	if err := p.cache.List(ctx, &userList); err != nil {
		return "", fmt.Errorf("list users: %w", err)
	}

	for _, user := range userList.Items {
		if user.Name == p.cfg.Username {
			if user.Annotations != nil {
				return user.Annotations["bytetrade.io/language"], nil
			}
		}
	}
	return "", nil
}

func (p *Provider) getOwnerApps(ctx context.Context) ([]*message.AppInfo, error) {
	var appList v1alpha1App.ApplicationList
	if err := p.cache.List(ctx, &appList, client.InNamespace("")); err != nil {
		return nil, fmt.Errorf("list applications: %w", err)
	}
	var svc corev1.Service
	if err := p.cache.Get(ctx, types.NamespacedName{Name: constants.BFLServiceName, Namespace: constants.Namespace}, &svc); err != nil {
		return nil, fmt.Errorf("get bfl-svc failed: %w", err)
	}

	var apps []*message.AppInfo
	for _, app := range appList.Items {
		if app.Spec.Owner != p.cfg.Username {
			continue
		}
		if app.Spec.Name == "" || app.Spec.Appid == "" {
			continue
		}

		entrances := make([]*message.EntranceInfo, 0, len(app.Spec.Entrances))
		for _, e := range app.Spec.Entrances {
			entrances = append(entrances, &message.EntranceInfo{
				Name:            e.Name,
				Host:            e.Host,
				Port:            e.Port,
				AuthLevel:       e.AuthLevel,
				WindowPushState: e.WindowPushState,
			})
		}

		ports := make([]*message.PortInfo, 0, len(app.Spec.Ports))
		for _, sp := range app.Spec.Ports {
			portName := fmt.Sprintf("stream-%s-%d", strings.ToLower(sp.Protocol), sp.ExposePort)
			if !p.checkIfStreamPortExist(&svc, portName) {
				klog.Warningf("port %s not in bfl-svc, skipping...", portName)
				continue
			}
			ports = append(ports, &message.PortInfo{
				Name:       sp.Name,
				Host:       sp.Host,
				Port:       sp.Port,
				ExposePort: sp.ExposePort,
				Protocol:   sp.Protocol,
			})
		}

		settings := make(map[string]string)
		for k, v := range app.Spec.Settings {
			settings[k] = v
		}

		apps = append(apps, &message.AppInfo{
			Name:      app.Spec.Name,
			Appid:     app.Spec.Appid,
			IsSysApp:  app.Spec.IsSysApp,
			Namespace: app.Spec.Namespace,
			Owner:     app.Spec.Owner,
			Entrances: entrances,
			Ports:     ports,
			Settings:  settings,
		})
	}

	sort.Slice(apps, func(i, j int) bool {
		return apps[i].Name < apps[j].Name
	})

	return apps, nil
}

func (p *Provider) getFileserverNodes(ctx context.Context) ([]*message.FileserverNodeInfo, error) {
	var podList corev1.PodList
	if err := p.cache.List(ctx, &podList, client.MatchingLabels{"app": "files"}); err != nil {
		return nil, fmt.Errorf("list files pods: %w", err)
	}

	var nodeList corev1.NodeList
	if err := p.cache.List(ctx, &nodeList); err != nil {
		return nil, fmt.Errorf("list nodes: %w", err)
	}

	masterNodes := make(map[string]bool)
	for _, node := range nodeList.Items {
		if _, ok := node.Labels["node-role.kubernetes.io/control-plane"]; ok {
			masterNodes[node.Name] = true
		}
	}

	podMap := make(map[string]*corev1.Pod)
	for i := range podList.Items {
		pod := &podList.Items[i]
		if pod.Labels["app"] == "files" && pod.Status.PodIP != "" {
			podMap[pod.Spec.NodeName] = pod
		}
	}

	serviceNamespace := fmt.Sprintf("user-system-%s", p.cfg.Username)
	var nodes []*message.FileserverNodeInfo
	for nodeName, pod := range podMap {
		svcKey := client.ObjectKey{
			Namespace: serviceNamespace,
			Name:      fmt.Sprintf("files-%s", nodeName),
		}
		var svc corev1.Service
		if err := p.cache.Get(ctx, svcKey, &svc); err != nil {
			klog.V(4).Infof("provider: proxy service %s not ready, skipping node %s", svcKey, nodeName)
			continue
		}

		nodes = append(nodes, &message.FileserverNodeInfo{
			NodeName: nodeName,
			PodIP:    pod.Status.PodIP,
			IsMaster: masterNodes[nodeName],
		})
	}

	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].NodeName < nodes[j].NodeName
	})

	return nodes, nil
}

func (p *Provider) getCustomDomainCerts(ctx context.Context) ([]*message.CertInfo, error) {
	var cmList corev1.ConfigMapList
	if err := p.cache.List(ctx, &cmList, client.MatchingLabels{
		v1alpha1App.AppEntranceCertConfigMapLabel: "true",
	}); err != nil {
		return nil, fmt.Errorf("list custom domain cert configmaps: %w", err)
	}

	var certs []*message.CertInfo
	for _, cm := range cmList.Items {
		domain := cm.Data[v1alpha1App.AppEntranceCertConfigMapZoneKey]
		certData := cm.Data[v1alpha1App.AppEntranceCertConfigMapCertKey]
		keyData := cm.Data[v1alpha1App.AppEntranceCertConfigMapKeyKey]
		if domain == "" || certData == "" || keyData == "" {
			continue
		}
		certs = append(certs, &message.CertInfo{
			Domain:   domain,
			CertData: certData,
			KeyData:  keyData,
		})
	}
	sort.Slice(certs, func(i, j int) bool {
		return certs[i].Domain < certs[j].Domain
	})
	return certs, nil
}

func (p *Provider) checkIfStreamPortExist(svc *corev1.Service, portName string) bool {
	for _, port := range svc.Spec.Ports {
		if port.Name == portName {
			return true
		}
	}
	return false
}

func isSSLConfigMap(cm *corev1.ConfigMap, namespace string) bool {
	return cm.Namespace == namespace && cm.Name == constants.NameSSLConfigMapName
}

func isCustomDomainCertConfigMap(cm *corev1.ConfigMap) bool {
	return strings.Index(cm.Name, constants.ApplicationThirdPartyDomainCertKeySuffix) > 0
}

func isFileServerPod(pod *corev1.Pod) bool {
	return pod.Labels["app"] == "files"
}

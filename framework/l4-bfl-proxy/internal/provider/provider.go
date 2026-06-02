package provider

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	iamv1alpha2 "github.com/beclab/api/iam/v1alpha2"
	"github.com/beclab/l4-bfl-proxy/internal/message"
	toolscache "k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	ctrlcache "sigs.k8s.io/controller-runtime/pkg/cache"
)

const (
	mapKey           = "default-bfl-proxy"
	retryDelay       = time.Second
	debounceInterval = 100 * time.Millisecond

	userAnnotationDid         = "bytetrade.io/did"
	userAnnotationZone        = "bytetrade.io/zone"
	userAnnotationOwnerRole   = "bytetrade.io/owner-role"
	userLauncherAccessLevel   = "bytetrade.io/launcher-access-level"
	userLauncherAllowCIDR     = "bytetrade.io/launcher-allow-cidr"
	userAnnotationCreator     = "bytetrade.io/creator"
	userAnnotationIsEphemeral = "bytetrade.io/is-ephemeral"
	userDenyAllPolicy         = "bytetrade.io/deny-all"
	userLocalDomainIPDns      = "bytetrade.io/local-domain-dns-record"

	settingsCustomDomain                 = "customDomain"
	settingsCustomDomainThirdLevelDomain = "third_level_domain"
	settingsCustomDomainThirdPartyDomain = "third_party_domain"
	applicationAuthLevelPublic           = "public"

	nameSSLConfigMapName                     = "zone-ssl-config"
	applicationThirdPartyDomainCertKeySuffix = "-domain-ssl-config"
	appEntranceCertConfigMapLabel            = "app.bytetrade.io/custom-domain-cert"
	appEntranceCertConfigMapCertKey          = "cert"
	appEntranceCertConfigMapKeyKey           = "key"
	appEntranceCertConfigMapZoneKey          = "zone"
)

type Config struct {
	UserNamespacePrefix string
	SSLServerPort       int
	SSLProxyServerPort  int
}

type Provider struct {
	cache      ctrlcache.Cache
	resources  *message.ProviderResources
	cfg        *Config
	synced     atomic.Bool
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

// SetupWithManager pre-registers informers and event handlers before the
// Manager starts. This ensures the cache includes User and Application
// informers in its initial sync, so cacheReadyCheck is accurate.
// Must be called before mgr.Start().
func (p *Provider) SetupWithManager(ctx context.Context) error {
	cmInformer, err := p.cache.GetInformer(ctx, &corev1.ConfigMap{})
	if err != nil {
		return fmt.Errorf("get configmap informer: %w", err)
	}

	appInformer, err := p.cache.GetInformer(ctx, &appv1alpha1.Application{})
	if err != nil {
		return fmt.Errorf("get application informer: %w", err)
	}

	userInformer, err := p.cache.GetInformer(ctx, &iamv1alpha2.User{})
	if err != nil {
		return fmt.Errorf("get user informer: %w", err)
	}

	podInformer, err := p.cache.GetInformer(ctx, &corev1.Pod{})
	if err != nil {
		return fmt.Errorf("get pod informer: %w", err)
	}

	baseHandler := cache.ResourceEventHandlerFuncs{
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
			return isSSLConfigMap(cm, p.cfg.UserNamespacePrefix) || isCustomDomainCertConfigMap(cm)
		},
		Handler: baseHandler,
	}); err != nil {
		return fmt.Errorf("add configmap event handler: %w", err)
	}

	if _, err := appInformer.AddEventHandler(toolscache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			_, ok := obj.(*appv1alpha1.Application)
			return ok
		},
		Handler: baseHandler,
	}); err != nil {
		return fmt.Errorf("add application event handler failed: %w", err)
	}

	if _, err = userInformer.AddEventHandler(toolscache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			_, ok := obj.(*iamv1alpha2.User)
			return ok
		},
		Handler: baseHandler,
	}); err != nil {
		return fmt.Errorf("add user event handler: %w", err)
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

	klog.Info("provider: informers and event handlers registered...")
	return nil
}

// Start is called by the Manager after the cache has synced.
// Informers are already registered and synced via SetupWithManager.
func (p *Provider) Start(ctx context.Context) error {
	p.synced.Store(true)
	klog.Info("provider: cache synced, publishing initial snapshot")
	if err := p.publishResources(ctx); err != nil {
		klog.Warningf("provider: initial publish failed, will retry: %v", err)
		p.scheduleRetry(ctx)
	}

	p.debounceLoop(ctx)
	klog.Info("provider: stopped...")
	return nil
}

func (p *Provider) notifyChanged() {
	if !p.synced.Load() {
		return
	}
	select {
	case p.debounceCh <- struct{}{}:
	default:
	}
}

// NotifyChanged is the exported version of notifyChanged, used as a callback
func (p *Provider) NotifyChanged() {
	p.notifyChanged()
}

func (p *Provider) debounceLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-p.debounceCh:
			timer := time.NewTimer(debounceInterval)
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
					timer.Reset(debounceInterval)
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
		return err
	}

	snapshot.Sort()

	if old, ok := p.resources.Load(mapKey); ok && old.Equal(snapshot) {
		klog.V(4).Info("provider: snapshot unchanged, skipping...")
		return nil
	}
	p.resources.Store(mapKey, snapshot)
	klog.Infof("provider: published snapshot success...")

	return nil
}

func (p *Provider) buildResources(ctx context.Context) (*message.Resources, error) {
	res := &message.Resources{}
	rawAppsMap, err := p.getAppsMap(ctx)
	if err != nil {
		return nil, err
	}
	userList, err := p.getUsers(ctx)
	if err != nil {
		return nil, err
	}
	rawAppsMap = fanOutSharedApps(rawAppsMap, userList)
	users, err := p.listUsers(ctx, userList, rawAppsMap)
	if err != nil {
		return nil, err
	}
	res.Users = users

	return res, nil
}

// fanOutSharedApps mutates the per-owner apps map so cluster-wide shared apps
// appear in EVERY user's slice unconditionally. Shared apps are open to all
// users, so every user gets the real upstream cluster routes — there is no
// per-user gate.
//
// Qualifying apps: v3 label, or v2 with spec.sharedEntrances and
// settings.clusterScoped=true (see isSharedGatewayApp).
//
// v1 / non-shared v2 apps remain only under their installer (Spec.Owner).
func fanOutSharedApps(rawAppsMap map[string][]*appv1alpha1.Application,
	userList []iamv1alpha2.User,
) map[string][]*appv1alpha1.Application {
	if len(rawAppsMap) == 0 || len(userList) == 0 {
		return rawAppsMap
	}

	// Collect shared gateway apps and de-dupe by name.
	sharedApps := make(map[string]*appv1alpha1.Application)
	for _, apps := range rawAppsMap {
		for _, app := range apps {
			if isSharedGatewayApp(app) {
				if _, ok := sharedApps[app.Spec.Name]; !ok {
					sharedApps[app.Spec.Name] = app
				}
			}
		}
	}
	if len(sharedApps) == 0 {
		return rawAppsMap
	}

	sharedNames := make([]string, 0, len(sharedApps))
	for name := range sharedApps {
		sharedNames = append(sharedNames, name)
	}
	sort.Strings(sharedNames)

	for _, u := range userList {
		existing := make(map[string]bool, len(rawAppsMap[u.Name]))
		for _, app := range rawAppsMap[u.Name] {
			existing[app.Spec.Name] = true
		}

		for _, name := range sharedNames {
			if existing[name] {
				continue
			}
			rawAppsMap[u.Name] = append(rawAppsMap[u.Name], sharedApps[name])
			existing[name] = true
		}
	}
	return rawAppsMap
}

// isV3App reports whether the Application is a v3 app, based on
// the AppApiVersionLabel that the v3 install handler / Application controller
// stamp on it.
func isV3App(app *appv1alpha1.Application) bool {
	return appv1alpha1.IsV3(app)
}

// isSharedGatewayApp reports whether the Application uses the shared gateway
// path (v3, or v2 cluster-scoped with sharedEntrances).
func isSharedGatewayApp(app *appv1alpha1.Application) bool {
	if app == nil || len(app.Spec.SharedEntrances) == 0 {
		return false
	}
	if appv1alpha1.IsV3(app) {
		return true
	}
	if app.Spec.Settings != nil && app.Spec.Settings["clusterScoped"] == "true" {
		return true
	}
	return false
}

// fileserverGlobalData holds pod/node data shared across all users within a
// single reconcile cycle. Fetching it once avoids N redundant cache List calls.
type fileserverGlobalData struct {
	podMap      map[string]*corev1.Pod // nodeName → files pod
	masterNodes map[string]bool        // nodeName → is control-plane
}

func (p *Provider) getFileserverGlobalData(ctx context.Context) (*fileserverGlobalData, error) {
	var podList corev1.PodList
	if err := p.cache.List(ctx, &podList, client.MatchingLabels{"app": "files"}); err != nil {
		return nil, fmt.Errorf("list files pods: %w", err)
	}

	var nodeList corev1.NodeList
	if err := p.cache.List(ctx, &nodeList); err != nil {
		return nil, fmt.Errorf("list nodes: %w", err)
	}

	masterNodes := make(map[string]bool, len(nodeList.Items))
	for _, node := range nodeList.Items {
		if _, ok := node.Labels["node-role.kubernetes.io/control-plane"]; ok {
			masterNodes[node.Name] = true
		}
	}

	podMap := make(map[string]*corev1.Pod, len(podList.Items))
	for i := range podList.Items {
		pod := &podList.Items[i]
		if pod.Labels["app"] == "files" && pod.Status.PodIP != "" {
			podMap[pod.Spec.NodeName] = pod
		}
	}

	return &fileserverGlobalData{podMap: podMap, masterNodes: masterNodes}, nil
}

func (p *Provider) getFileserverNodesForUser(ctx context.Context, username string, gd *fileserverGlobalData) ([]*message.FileserverNodeInfo, error) {
	serviceNamespace := fmt.Sprintf("user-system-%s", username)
	var nodes []*message.FileserverNodeInfo
	for nodeName, pod := range gd.podMap {
		svcKey := client.ObjectKey{
			Namespace: serviceNamespace,
			Name:      fmt.Sprintf("files-%s", nodeName),
		}
		var svc corev1.Service
		if err := p.cache.Get(ctx, svcKey, &svc); err != nil {
			if apierrors.IsNotFound(err) {
				// Service not yet created by fileserver-reconciler; skip this node for now.
				klog.V(4).Infof("provider: files proxy service %s not found, skipping node %s", svcKey, nodeName)
				continue
			}
			return nil, fmt.Errorf("get files proxy service %s: %w", svcKey, err)
		}

		nodes = append(nodes, &message.FileserverNodeInfo{
			NodeName: nodeName,
			PodIP:    pod.Status.PodIP,
			IsMaster: gd.masterNodes[nodeName],
		})
	}

	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].NodeName < nodes[j].NodeName
	})

	return nodes, nil
}

// buildAppInfos translates Application CRs into the AppInfo shape the
// translator consumes. For v3 apps the per-user overlay
// (Spec.UserSettings[username]) is applied to both Settings and Entrances
// so each user's xDS sees their own customDomain / authLevel choices.
//
// Owner on the resulting AppInfo is the viewer (username), matching the
// per-user fan-out semantics already used by `fanOutSharedApps`. The
// install-owner is still recoverable from app.Spec.Owner if a downstream
// consumer needs it (none today).
func (p *Provider) buildAppInfos(username string, appList []*appv1alpha1.Application) []*message.AppInfo {
	var result []*message.AppInfo
	for _, app := range appList {
		if app.Spec.Name == "" || app.Spec.Appid == "" {
			continue
		}

		// Per-user entrances always come first and carry IsShared=false so the
		// translator keeps them on the legacy <appid><idx>.<zone> direct path.
		// SharedEntrances (gateway-only) follow with IsShared=true so the
		// translator can apply the multi-tenant <hash8>.<viewer>.<domain>
		// rewrite without disturbing the per-user entrances of the same app.
		effectiveEntrances := app.EffectiveEntrances(username)
		entrances := make([]*message.EntranceInfo, 0, len(effectiveEntrances)+len(app.Spec.SharedEntrances))
		for _, e := range effectiveEntrances {
			entrances = append(entrances, &message.EntranceInfo{
				Name:            e.Name,
				Host:            e.Host,
				Port:            e.Port,
				AuthLevel:       e.AuthLevel,
				WindowPushState: e.WindowPushState,
				IsShared:        false,
			})
		}
		if isSharedGatewayApp(app) {
			baseAppID := ownerAppID(app.Spec.Name, app.Spec.IsSysApp)
			sharedCount := len(app.Spec.SharedEntrances)
			for i, e := range app.Spec.SharedEntrances {
				sharedEntranceID := ""
				if baseAppID != "" {
					sharedEntranceID = baseAppID
					if sharedCount > 1 {
						sharedEntranceID = fmt.Sprintf("%s%d", baseAppID, i)
					}
				}
				entrances = append(entrances, &message.EntranceInfo{
					Name:             e.Name,
					Host:             e.Host,
					Port:             e.Port,
					AuthLevel:        e.AuthLevel,
					WindowPushState:  e.WindowPushState,
					IsShared:         true,
					SharedEntranceID: sharedEntranceID,
				})
			}
		}

		ports := make([]*message.PortInfo, 0, len(app.Spec.Ports))
		for _, sp := range app.Spec.Ports {
			ports = append(ports, &message.PortInfo{
				Name:       sp.Name,
				Host:       sp.Host,
				Port:       sp.Port,
				ExposePort: sp.ExposePort,
				Protocol:   sp.Protocol,
			})
		}

		settings := app.EffectiveSettings(username)

		// For shared gateway apps, expose the viewer as Owner so downstream
		// per-user grouping by AppInfo.Owner keeps working.
		owner := app.Spec.Owner
		if isSharedGatewayApp(app) {
			owner = username
		}

		annotations := map[string]string{}
		for k, v := range app.Annotations {
			annotations[k] = v
		}

		result = append(result, &message.AppInfo{
			Name:        app.Spec.Name,
			Appid:       app.Spec.Appid,
			IsSysApp:    app.Spec.IsSysApp,
			Namespace:   app.Spec.Namespace,
			Owner:       owner,
			Entrances:   entrances,
			Ports:       ports,
			Settings:    settings,
			Annotations: annotations,
			IsShared:    isSharedGatewayApp(app),
		})
	}
	return result
}

func ownerAppID(appName string, isSysApp bool) string {
	appName = strings.TrimSpace(appName)
	if appName == "" {
		return ""
	}
	if isSysApp {
		return appName
	}
	sum := md5.Sum([]byte(appName))
	return hex.EncodeToString(sum[:])[:8]
}

func (p *Provider) listUsers(ctx context.Context, userList []iamv1alpha2.User, rawAppsMap map[string][]*appv1alpha1.Application) ([]*message.UserInfo, error) {
	var result []*message.UserInfo

	// Fetch cluster-wide data once; reused for every user below.
	allCerts, err := p.getCustomDomainCerts(ctx)
	if err != nil {
		return nil, err
	}
	fsGlobal, err := p.getFileserverGlobalData(ctx)
	if err != nil {
		return nil, err
	}

	getUserByName := func(name string) *iamv1alpha2.User {
		for i := range userList {
			if userList[i].Name == name {
				return &userList[i]
			}
			if name == "cli" && userList[i].Annotations[userAnnotationOwnerRole] == "owner" {
				return &userList[i]
			}
		}
		return nil
	}

	getPublicAccessDomain := func(zone string, publicAppIDs, publicCustomDomainApps []string, denied string) []string {
		r := []string{zone}
		if denied != "1" {
			return r
		}
		for _, appID := range publicAppIDs {
			r = append(r, appID+"."+zone)
		}
		r = append(r, publicCustomDomainApps...)
		return r
	}

	for _, user := range userList {
		publicAppIDs, publicCustomDomainApps, _, customDomainAppsWithUsers := p.listApplicationDetails(user.Name, rawAppsMap[user.Name])

		isEphemeralAnno := getAnnotation(&user, userAnnotationIsEphemeral)
		if !isValidUser(&user) && isEphemeralAnno == "" {
			continue
		}

		isEphemeral := false
		if ok, parseErr := strconv.ParseBool(isEphemeralAnno); parseErr == nil && ok {
			isEphemeral = true
		}

		var (
			did, zone, localDomainIP string
			accLevel, allowCIDR      string
			denyAllStatus            string
			allowedDomains           []string
			serverNameDomains        []string
		)
		annoUser := &user

		if isEphemeral {
			creator := getAnnotation(&user, userAnnotationCreator)
			creatorUser := getUserByName(creator)
			if creatorUser == nil {
				klog.Warningf("provider: ephemeral user %q has no creator", user.Name)
				continue
			}
			annoUser = creatorUser
		}
		did = getAnnotation(annoUser, userAnnotationDid)
		zone = getAnnotation(annoUser, userAnnotationZone)
		localDomainIP = getAnnotation(annoUser, userLocalDomainIPDns)
		accLevel = getAnnotation(annoUser, userLauncherAccessLevel)
		allowCIDR = getAnnotation(annoUser, userLauncherAllowCIDR)
		serverNameDomains = []string{zone, annoUser.Name + ".olares.local"}
		denyAllStatus = getAnnotation(annoUser, userDenyAllPolicy)
		allowedDomains = getPublicAccessDomain(zone, publicAppIDs, publicCustomDomainApps, denyAllStatus)

		if userCustomDomains, ok := customDomainAppsWithUsers[annoUser.Name]; ok && len(userCustomDomains) > 0 {
			serverNameDomains = append(serverNameDomains, userCustomDomains...)
		}

		var accessLevel uint64
		if accLevel != "" {
			var err error
			accessLevel, err = strconv.ParseUint(accLevel, 10, 64)
			if err != nil {
				klog.Errorf("provider: user %q parse access level: %v", user.Name, err)
				continue
			}
		}

		denyAll, err := strconv.Atoi(denyAllStatus)
		if err != nil && denyAllStatus != "" {
			klog.Warningf("provider: user %q has invalid deny-all annotation value %q", user.Name, denyAllStatus)
		}

		cidrs := parseAllowCIDRs(allowCIDR)

		language := getUserLanguage(&user)
		sslConfig, err := p.getSSLConfig(ctx, user.Name)
		if err != nil {
			if apierrors.IsNotFound(err) {
				klog.V(4).Infof("provider: user %s has no SSL configmap yet, skipping", user.Name)
				continue
			}
			return nil, err
		}
		fileserverNodes, err := p.getFileserverNodesForUser(ctx, user.Name, fsGlobal)
		if err != nil {
			return nil, err
		}
		masterNodeCIDR, err := p.getMasterNodeCIDR(ctx)
		if err != nil {
			return nil, err
		}

		info := &message.UserInfo{
			Name:              user.Name,
			Namespace:         fmt.Sprintf("%s-%s", p.cfg.UserNamespacePrefix, user.Name),
			Did:               did,
			Zone:              zone,
			IsEphemeral:       isEphemeral,
			AccessLevel:       accessLevel,
			AllowCIDRs:        cidrs,
			DenyAll:           denyAll == 1,
			AllowedDomains:    allowedDomains,
			ServerNameDomains: serverNameDomains,
			LocalDomainIP:     localDomainIP,
			CreateTimestamp:   user.CreationTimestamp.Unix(),
			Language:          language,
			Apps:              p.buildAppInfos(user.Name, rawAppsMap[user.Name]),
			SSL:               sslConfig,
			CustomDomainCerts: allCerts[user.Name],
			FileserverNodes:   fileserverNodes,
			MasterNodeCIDR:    masterNodeCIDR,
		}
		result = append(result, info)
	}

	return result, nil
}

func (p *Provider) getCustomDomainCerts(ctx context.Context) (map[string][]*message.CertInfo, error) {
	var cmList corev1.ConfigMapList
	if err := p.cache.List(ctx, &cmList, client.MatchingLabels{
		appEntranceCertConfigMapLabel: "true",
	}); err != nil {
		return nil, fmt.Errorf("list custom domain cert configmaps: %w", err)
	}

	certs := make(map[string][]*message.CertInfo)
	for _, cm := range cmList.Items {
		owner := strings.TrimPrefix(cm.Namespace, "user-space-")
		domain := cm.Data[appEntranceCertConfigMapZoneKey]
		certData := cm.Data[appEntranceCertConfigMapCertKey]
		keyData := cm.Data[appEntranceCertConfigMapKeyKey]
		if domain == "" || certData == "" || keyData == "" {
			continue
		}
		certs[owner] = append(certs[owner], &message.CertInfo{
			Domain:   domain,
			CertData: certData,
			KeyData:  keyData,
		})
	}
	for owner := range certs {
		ownerCerts := certs[owner]
		sort.Slice(ownerCerts, func(i, j int) bool {
			return ownerCerts[i].Domain < ownerCerts[j].Domain
		})
	}

	return certs, nil
}

func (p *Provider) getSSLConfig(ctx context.Context, username string) (*message.SSLConfig, error) {
	var cm corev1.ConfigMap
	key := client.ObjectKey{
		Namespace: fmt.Sprintf("%s-%s", p.cfg.UserNamespacePrefix, username),
		Name:      nameSSLConfigMapName,
	}
	if err := p.cache.Get(ctx, key, &cm); err != nil {
		return nil, fmt.Errorf("get ssl configmap: %w", err)
	}

	if cm.Data == nil {
		return nil, fmt.Errorf("ssl configmap with empty data")
	}
	zone := cm.Data["zone"]
	if zone == "" {
		return nil, fmt.Errorf("ssl configmap with empty zone")
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
func getUserLanguage(user *iamv1alpha2.User) string {
	if user.Annotations != nil {
		return user.Annotations["bytetrade.io/language"]
	}
	return ""
}

// listApplicationDetails summarises the viewer's per-user view of a set
// of Applications. For v3 apps the Settings/Entrances overlays are
// applied so the customDomain / authLevel slots read from
// UserSettings[username] instead of the global Settings. The viewer is
// also recorded as the per-domain "owner" so the rest of the pipeline
// associates this user's frontends with their own server_name list.
func (p *Provider) listApplicationDetails(username string, appList []*appv1alpha1.Application) ([]string, []string, []string, map[string][]string) {
	publicApps := []string{"headscale"}
	var publicCustomDomainApps []string
	var customDomainApps []string
	customDomainAppsWithUsers := make(map[string][]string)

	getAppPrefix := func(entranceCount, index int, appid string) string {
		if entranceCount == 1 {
			return appid
		}
		return fmt.Sprintf("%s%d", appid, index)
	}

	for _, app := range appList {
		effectiveEntrances := app.EffectiveEntrances(username)
		if len(effectiveEntrances) == 0 {
			continue
		}

		var customDomains []string
		var customDomainsPrefix []string
		entranceCount := len(effectiveEntrances)
		// For shared gateway apps the install-owner is irrelevant — every
		// user gets their own customDomain bucket.
		owner := app.Spec.Owner
		if isSharedGatewayApp(app) {
			owner = username
		}
		customDomainEntrancesMap := getSettingsKeyMap(app, username, settingsCustomDomain)

		for index, entrance := range effectiveEntrances {
			prefix := getAppPrefix(entranceCount, index, app.Spec.Appid)
			authLevel := entrance.AuthLevel

			if cdEntrance, ok := customDomainEntrancesMap[entrance.Name]; ok {
				if entrancePrefix := cdEntrance[settingsCustomDomainThirdLevelDomain]; entrancePrefix != "" {
					if authLevel == applicationAuthLevelPublic {
						customDomainsPrefix = append(customDomainsPrefix, entrancePrefix)
					}
				}
				if entranceCustomDomain := cdEntrance[settingsCustomDomainThirdPartyDomain]; entranceCustomDomain != "" {
					customDomainApps = append(customDomainApps, entranceCustomDomain)

					val := customDomainAppsWithUsers[owner]
					customDomainAppsWithUsers[owner] = append(val, entranceCustomDomain)

					if authLevel == applicationAuthLevelPublic {
						customDomains = append(customDomains, entranceCustomDomain)
					}
				}
			}

			if authLevel == applicationAuthLevelPublic {
				publicApps = append(publicApps, prefix)
			}
		}

		publicApps = append(publicApps, customDomainsPrefix...)
		publicCustomDomainApps = append(publicCustomDomainApps, customDomains...)
	}

	return publicApps, publicCustomDomainApps, customDomainApps, customDomainAppsWithUsers
}

func (p *Provider) getAppsMap(ctx context.Context) (map[string][]*appv1alpha1.Application, error) {
	var appList appv1alpha1.ApplicationList
	if err := p.cache.List(ctx, &appList); err != nil {
		klog.Errorf("provider: list apps from cache: %v", err)
		return nil, fmt.Errorf("list apps from cache failed: %v", err)
	}
	appsMap := make(map[string][]*appv1alpha1.Application)
	for _, app := range appList.Items {
		appsMap[app.Spec.Owner] = append(appsMap[app.Spec.Owner], app.DeepCopy())
	}
	for owner := range appsMap {
		sort.Slice(appsMap[owner], func(i, j int) bool {
			return appsMap[owner][i].Name < appsMap[owner][j].Name
		})
	}
	return appsMap, nil
}

func (p *Provider) getUsers(ctx context.Context) ([]iamv1alpha2.User, error) {
	var userList iamv1alpha2.UserList
	if err := p.cache.List(ctx, &userList); err != nil {
		klog.Errorf("provider: list users from cache failed: %v", err)
		return nil, fmt.Errorf("list users from cache failed: %v", err)
	}
	users := make([]iamv1alpha2.User, 0)
	for _, user := range userList.Items {
		if user.Status.State != "Created" {
			continue
		}
		users = append(users, user)
	}
	return users, nil
}

func (p *Provider) getMasterNodeCIDR(ctx context.Context) (string, error) {
	var nodeList corev1.NodeList
	if err := p.cache.List(ctx, &nodeList, client.HasLabels{"node-role.kubernetes.io/control-plane"}); err != nil {
		klog.Errorf("provider: list node failed: %v", err)
		return "", err
	}
	if len(nodeList.Items) == 0 {
		return "", errors.New("no master node found")
	}
	node := nodeList.Items[0]
	if len(node.Annotations) == 0 {
		klog.Warningf("provider: node %s with empty annotations", node.Name)
		return "", nil
	}
	cidr, ok := node.Annotations["projectcalico.org/IPv4Address"]
	if !ok {
		klog.Warningf("provider: node %s has no projectcalico.org/IPv4Address annotation", node.Name)
		return "", nil
	}
	return cidr, nil
}

func getAnnotation(user *iamv1alpha2.User, key string) string {
	if v, ok := user.Annotations[key]; ok && v != "" {
		return v
	}
	return ""
}

func isValidUser(user *iamv1alpha2.User) bool {
	return getAnnotation(user, userAnnotationDid) != "" && getAnnotation(user, userAnnotationZone) != ""
}

// getSettingsKeyMap unmarshals a Settings blob (e.g. "customDomain" or
// "policy") into a map keyed by entrance name. For v3 apps it reads the
// per-user overlay (Spec.UserSettings[username][key]) first; for v1/v2 it
// falls back to Spec.Settings[key].
func getSettingsKeyMap(app *appv1alpha1.Application, username, key string) map[string]map[string]string {
	r := make(map[string]map[string]string)
	settings := app.EffectiveSettings(username)
	data := settings[key]
	if data == "" {
		return r
	}
	if err := json.Unmarshal([]byte(data), &r); err != nil {
		klog.Warningf("provider: unmarshal settings %q for app %s/%s (user %q) failed: %v", key, app.Namespace, app.Name, username, err)
		return make(map[string]map[string]string)
	}
	return r
}

func parseAllowCIDRs(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	cidrs := make([]string, 0, len(parts))
	for _, part := range parts {
		cidr := strings.TrimSpace(part)
		if cidr == "" {
			continue
		}
		cidrs = append(cidrs, cidr)
	}
	sort.Strings(cidrs)
	return cidrs
}

func isSSLConfigMap(cm *corev1.ConfigMap, namespacePrefix string) bool {
	return strings.HasPrefix(cm.Namespace, namespacePrefix) && cm.Name == nameSSLConfigMapName
}

func isCustomDomainCertConfigMap(cm *corev1.ConfigMap) bool {
	return strings.Contains(cm.Name, applicationThirdPartyDomainCertKeySuffix)
}

func isFileServerPod(pod *corev1.Pod) bool {
	return pod.Labels["app"] == "files"
}

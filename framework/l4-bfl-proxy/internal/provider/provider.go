package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	appv1alpha1 "github.com/beclab/Olares/framework/app-service/api/app.bytetrade.io/v1alpha1"
	iamv1alpha2 "github.com/beclab/api/iam/v1alpha2"
	"github.com/beclab/l4-bfl-proxy/internal/message"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	ctrlcache "sigs.k8s.io/controller-runtime/pkg/cache"
)

const (
	mapKey           = "default"
	dnsRetry         = time.Second
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
)

type Config struct {
	UserNamespacePrefix string
	BFLServicePort      int
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
	userInformer, err := p.cache.GetInformer(ctx, &iamv1alpha2.User{})
	if err != nil {
		return fmt.Errorf("get user informer: %w", err)
	}

	appInformer, err := p.cache.GetInformer(ctx, &appv1alpha1.Application{})
	if err != nil {
		return fmt.Errorf("get app informer: %w", err)
	}

	handler := cache.ResourceEventHandlerFuncs{
		AddFunc:    func(_ interface{}) { p.notifyChanged() },
		UpdateFunc: func(_, _ interface{}) { p.notifyChanged() },
		DeleteFunc: func(_ interface{}) { p.notifyChanged() },
	}
	if _, err := userInformer.AddEventHandler(handler); err != nil {
		return fmt.Errorf("add user event handler failed: %w", err)
	}
	if _, err := appInformer.AddEventHandler(handler); err != nil {
		return fmt.Errorf("add app event handler failed: %w", err)
	}

	klog.Info("provider: informers and event handlers registered")
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
	klog.Info("provider: stopped")
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
	rawApps, err := p.getAppsFromCache(ctx)
	if err != nil {
		return err
	}

	users, dnsFailures, err := p.listUsers(ctx, rawApps)
	if err != nil {
		klog.Errorf("provider: list users: %v", err)
		return err
	}

	// When DNS fails for some users, merge stale data from the previous
	// snapshot so those users' filter chains are not removed from Envoy.
	if dnsFailures > 0 {
		users = p.mergeStaleUsers(users)
	}

	snapshot := &message.Resources{
		Users: users,
		Apps:  p.buildAppInfos(rawApps),
	}
	snapshot.Sort()

	if old, ok := p.resources.Load(mapKey); ok && old.Equal(snapshot) {
		klog.V(4).Info("provider: snapshot unchanged, skipping publish")
	} else {
		p.resources.Store(mapKey, snapshot)
		klog.Infof("provider: published snapshot with %d users and %d apps", len(snapshot.Users), len(snapshot.Apps))
	}

	if dnsFailures > 0 {
		klog.V(2).Infof("provider: %d user(s) pending DNS resolution, retrying in %s", dnsFailures, dnsRetry)
		go func() {
			select {
			case <-ctx.Done():
			case <-time.After(dnsRetry):
				p.notifyChanged()
			}
		}()
	}
	return nil
}

// mergeStaleUsers fills in users whose DNS resolution failed with their data
// from the previous snapshot, preventing Envoy from removing their filter chains.
func (p *Provider) mergeStaleUsers(current []*message.UserInfo) []*message.UserInfo {
	old, ok := p.resources.Load(mapKey)
	if !ok || old == nil {
		return current
	}

	present := make(map[string]bool)
	for _, u := range current {
		present[u.Name] = true
	}

	for _, stale := range old.Users {
		if !present[stale.Name] {
			klog.V(2).Infof("provider: retaining stale data for user %q (DNS pending)", stale.Name)
			current = append(current, stale)
		}
	}
	return current
}

func (p *Provider) buildAppInfos(appList []appv1alpha1.Application) []*message.AppInfo {
	var result []*message.AppInfo
	for _, app := range appList {
		entrances := make([]message.EntranceInfo, 0, len(app.Spec.Entrances))
		for _, e := range app.Spec.Entrances {
			entrances = append(entrances, message.EntranceInfo{
				Name:      e.Name,
				AuthLevel: e.AuthLevel,
			})
		}

		ports := make([]message.PortInfo, 0, len(app.Spec.Ports))
		for _, sp := range app.Spec.Ports {
			ports = append(ports, message.PortInfo{
				Name:       sp.Name,
				Host:       sp.Host,
				Port:       sp.Port,
				ExposePort: sp.ExposePort,
				Protocol:   sp.Protocol,
			})
		}

		result = append(result, &message.AppInfo{
			Name:      app.Spec.Name,
			Appid:     app.Spec.Appid,
			Owner:     app.Spec.Owner,
			Entrances: entrances,
			Ports:     ports,
		})
	}
	return result
}

func (p *Provider) listUsers(ctx context.Context, rawApps []appv1alpha1.Application) ([]*message.UserInfo, int, error) {
	publicAppIDs, publicCustomDomainApps, _, customDomainAppsWithUsers := p.listApplicationDetails(rawApps)

	userList, err := p.getUsersFromCache(ctx)
	if err != nil {
		return nil, 0, err
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
		if (publicAppIDs == nil && publicCustomDomainApps == nil) || denied != "1" {
			return r
		}
		for _, appID := range publicAppIDs {
			r = append(r, appID+"."+zone)
		}
		r = append(r, publicCustomDomainApps...)
		return r
	}

	var result []*message.UserInfo
	var dnsFailures int

	for _, user := range userList {
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

		if !isEphemeral {
			did = getAnnotation(&user, userAnnotationDid)
			zone = getAnnotation(&user, userAnnotationZone)
			localDomainIP = getAnnotation(&user, userLocalDomainIPDns)
			accLevel = getAnnotation(&user, userLauncherAccessLevel)
			allowCIDR = getAnnotation(&user, userLauncherAllowCIDR)
			serverNameDomains = []string{zone, user.Name + ".olares.local"}
			denyAllStatus = getAnnotation(&user, userDenyAllPolicy)
			allowedDomains = getPublicAccessDomain(zone, publicAppIDs, publicCustomDomainApps, denyAllStatus)

			if userCustomDomains, ok := customDomainAppsWithUsers[user.Name]; ok && len(userCustomDomains) > 0 {
				serverNameDomains = append(serverNameDomains, userCustomDomains...)
			}
		} else {
			creator := getAnnotation(&user, userAnnotationCreator)
			creatorUser := getUserByName(creator)
			if creatorUser == nil {
				klog.Warningf("provider: ephemeral user %q has no creator", user.Name)
				continue
			}
			did = getAnnotation(creatorUser, userAnnotationDid)
			zone = getAnnotation(creatorUser, userAnnotationZone)
			accLevel = getAnnotation(creatorUser, userLauncherAccessLevel)
			allowCIDR = getAnnotation(creatorUser, userLauncherAllowCIDR)
			denyAllStatus = getAnnotation(creatorUser, userDenyAllPolicy)
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

		denyAll, _ := strconv.Atoi(denyAllStatus)

		svcName := fmt.Sprintf("bfl.%s-%s", p.cfg.UserNamespacePrefix, user.Name)
		addr, err := lookupHostAddr(svcName)
		if err != nil {
			klog.V(2).Infof("provider: user %q lookup host: %v, will retry", user.Name, err)
			dnsFailures++
			continue
		}

		cidrs := parseAllowCIDRs(allowCIDR)
		sort.Strings(allowedDomains)
		sort.Strings(serverNameDomains)

		info := &message.UserInfo{
			Name:              user.Name,
			Namespace:         fmt.Sprintf("%s-%s", p.cfg.UserNamespacePrefix, user.Name),
			Did:               did,
			Zone:              zone,
			IsEphemeral:       isEphemeral,
			BFLHost:           addr,
			BFLPort:           p.cfg.BFLServicePort,
			AccessLevel:       accessLevel,
			AllowCIDRs:        cidrs,
			DenyAll:           denyAll == 1,
			AllowedDomains:    allowedDomains,
			ServerNameDomains: serverNameDomains,
			LocalDomainIP:     localDomainIP,
			CreateTimestamp:   user.CreationTimestamp.Unix(),
		}
		result = append(result, info)
	}

	return result, dnsFailures, nil
}

func (p *Provider) listApplicationDetails(appList []appv1alpha1.Application) ([]string, []string, []string, map[string][]string) {
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
		if len(app.Spec.Entrances) == 0 {
			continue
		}

		var customDomains []string
		var customDomainsPrefix []string
		entranceCount := len(app.Spec.Entrances)
		owner := app.Spec.Owner
		customDomainEntrancesMap := getSettingsKeyMap(&app, settingsCustomDomain)

		for index, entrance := range app.Spec.Entrances {
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

func (p *Provider) getAppsFromCache(ctx context.Context) ([]appv1alpha1.Application, error) {
	var appList appv1alpha1.ApplicationList
	if err := p.cache.List(ctx, &appList); err != nil {
		klog.Errorf("provider: list apps from cache: %v", err)
		return nil, fmt.Errorf("list apps from cache failed: %v", err)
	}
	apps := appList.Items
	return apps, nil
}

func (p *Provider) getUsersFromCache(ctx context.Context) ([]iamv1alpha2.User, error) {
	var userList iamv1alpha2.UserList
	if err := p.cache.List(ctx, &userList); err != nil {
		klog.Errorf("provider: list users from cache: %v", err)
		return nil, fmt.Errorf("list apps from cache failed: %v", err)
	}
	users := userList.Items
	return users, nil
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

func getSettingsKeyMap(app *appv1alpha1.Application, key string) map[string]map[string]string {
	r := make(map[string]map[string]string)
	if app.Spec.Settings == nil {
		return r
	}
	data := app.Spec.Settings[key]
	if data == "" {
		return r
	}
	if err := json.Unmarshal([]byte(data), &r); err != nil {
		klog.Warningf("provider: unmarshal settings %q for app %s/%s failed: %v", key, app.Namespace, app.Name, err)
		return make(map[string]map[string]string)
	}
	return r
}

func lookupHostAddr(svc string) (string, error) {
	addrs, err := net.LookupHost(svc)
	if err != nil {
		return "", err
	}
	if len(addrs) == 0 {
		return "", fmt.Errorf("svc %s: no address resolved", svc)
	}
	return addrs[0], nil
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

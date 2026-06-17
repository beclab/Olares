package watchers

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"net"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"bytetrade.io/web3os/tapr/pkg/app/application"
	"github.com/coredns/corefile-migration/migration/corefile"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

const (
	labelDNSPassthrough     = "gateway.olares.io/dns-passthrough"
	labelSRRAppID           = "gateway.olares.io/appid"
	labelSRREntrance        = "gateway.olares.io/entrance"
	labelAppShared          = "app.bytetrade.io/app-shared"
	labelAppSharedTrue      = "true"
	labelNSAppName          = "applications.app.bytetrade.io/name"
	labelNSShared           = "bytetrade.io/ns-shared"
	labelNSInstallUser      = "applications.app.bytetrade.io/install_user"
	labelNSNamespace        = "bytetrade.io/namespace"
	appGatewayNamespace     = "os-gateway"
	appGatewayDataService   = "app-gateway-data"
	srrRouteModeGateway     = "gateway"
	corefileSizeWarnBytes   = 800 * 1024
	corefileSizeRejectBytes = 950 * 1024
	defaultPodCIDR          = "10.233.64.0/18"
)

var calicoIPPoolGVR = schema.GroupVersionResource{
	Group:    "crd.projectcalico.org",
	Version:  "v1",
	Resource: "ippools",
}

var sharedRouteRegistryGVR = schema.GroupVersionResource{
	Group:    "gateway.olares.io",
	Version:  "v1alpha1",
	Resource: "sharedrouteregistries",
}

// RegenerateCorefile rebuilds the CoreDNS Corefile from current cluster state.
//
// behavior: every invocation re-reads ClusterConfig.spec.inClusterGatewayEnabled
// and applies or removes Shared in-cluster templates accordingly.
//
// requirement: this function is intentionally event-driven by existing SRR/User/DNS
// watcher paths. ClusterConfig changes take effect on the next regeneration event
// (delayed linkage), because sys-event does not register a dedicated ClusterConfig watcher.
func RegenerateCorefile(ctx context.Context, kubeClient kubernetes.Interface, dynamicClient dynamic.Interface) error {
	corefileConfigMap, err := kubeClient.CoreV1().ConfigMaps("kube-system").Get(ctx, "coredns", metav1.GetOptions{})
	if err != nil {
		klog.Error("get coredns configmap error, ", err)
		return err
	}

	corefileData := corefileConfigMap.Data["Corefile"]
	file, err := corefile.New(corefileData)
	if err != nil {
		klog.Error("parse corefile error, ", err)
		return err
	}

	if len(file.Servers) < 1 {
		klog.Warning("invalid corefile configuration")
		return nil
	}

	defaultsServer := file.Servers[0]
	var defaultPlugins []*corefile.Plugin

	// put the hosts plugin before other plugins, especially the forward plugin
	defaultPlugins = append(defaultPlugins, &corefile.Plugin{
		Name: "hosts",
		Args: []string{"/node-etc/hosts"},
		Options: []*corefile.Option{
			{
				Name: "ttl",
				Args: []string{"30"},
			},
			{
				Name: "fallthrough",
			},
		},
	})

	for _, p := range defaultsServer.Plugins {
		switch p.Name {
		case "errors", "health", "ready", "kubernetes", "prometheus", "forward", "cache", "loop", "reload", "loadbalance":
			defaultPlugins = append(defaultPlugins, p)
		}
	}
	defaultPlugins = normalizeReloadInDefaultPlugins(defaultPlugins)

	userList, err := dynamicClient.Resource(schema.GroupVersionResource{
		Group:    "iam.kubesphere.io",
		Version:  "v1alpha2",
		Resource: "users",
	}).List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.Error("get userlist error, ", err)
		return err
	}

	nodeList, err := kubeClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.Error("get nodelist error, ", err)
		return err
	}

	var masterNodeIp string
	for _, node := range nodeList.Items {
		if _, ok := node.Labels["node-role.kubernetes.io/control-plane"]; ok {
			for _, addr := range node.Status.Addresses {
				if addr.Type == "InternalIP" {
					masterNodeIp = addr.Address
					break
				}
			}
		}
	}

	var templatesPlugins []*corefile.Plugin
	var inclusterTemplatesPlugins []*corefile.Plugin
	var localTemplatesPlugins []*corefile.Plugin
	var localDomainTemplatesPlugins []*corefile.Plugin

	addUserTemplates := func(zone, ip string, plugins []*corefile.Plugin) []*corefile.Plugin {
		newOptions := []*corefile.Option{
			{
				Name: "match",
				Args: []string{fmt.Sprintf("\"\\w*\\.?(%s\\.)$\"", zone)},
			},
			{
				Name: "answer",
				Args: []string{fmt.Sprintf("\"{{ .Name }} 60 IN A %s\"", ip)},
			},
			{
				Name: "fallthrough",
				Args: []string{},
			},
		}
		anyOptions := []*corefile.Option{
			{
				Name: "rcode",
				Args: []string{"NOERROR"},
			},
		}
		userTemplateArgs := []string{"IN", "A", zone}
		userTemplateAnyArgs := []string{"IN", "ANY", zone}

		plugins = append(plugins, &corefile.Plugin{
			Name:    "template",
			Args:    userTemplateArgs,
			Options: newOptions,
		})

		plugins = append(plugins, &corefile.Plugin{
			Name:    "template",
			Args:    userTemplateAnyArgs,
			Options: anyOptions,
		})

		return plugins
	} // func addUserTemplates

	ingressIp, err := getUserIngressIP(ctx, kubeClient)
	if err != nil {
		klog.Error("get user ingress ip error, ", err)
		return err
	}

	for _, u := range userList.Items {
		userzone := u.GetAnnotations()[UserAnnotationZoneKey]
		if userzone == "" {
			klog.Info("user ", u.GetName(), " has no zone annotation, skip corefile update")
			continue
		}

		ip, err := getUserLocalIp(&u)
		if err != nil {
			klog.Error("get user local ip error, ", err)
			return err
		}
		if ip == nil || ip.String() == "" {
			klog.Info("user ", u.GetName(), " has no valid local ip, skip corefile update")
			continue
		}

		if ingressIp == "" {
			klog.Info("user ", u.GetName(), " has no valid ingress ip, skip corefile update")
			continue
		}

		templatesPlugins = addUserTemplates(userzone, ip.String(), templatesPlugins)
		inclusterTemplatesPlugins = addUserTemplates(userzone, ingressIp, inclusterTemplatesPlugins)

		if masterNodeIp == "" {
			klog.Info("no master node ip found, skip adding local domain dns record")
			continue
		}

		username := u.GetName()
		userLocalZone := fmt.Sprintf("%s.olares.local", username)
		localTemplatesPlugins = addUserTemplates(userzone, masterNodeIp, localTemplatesPlugins)
		localDomainTemplatesPlugins = addUserTemplates(userLocalZone, masterNodeIp, localDomainTemplatesPlugins)
	}

	// Degrade on failure: without the app-gateway-data ClusterIP the shared
	// in-cluster templates are skipped, while the rest of the Corefile still
	// regenerates. A missing Service (chart not installed) keeps behavior
	// identical to a cluster without app-gateway.
	gatewayDataIP, gatewayDataIPErr := appGatewayDataClusterIP(ctx, kubeClient)
	if gatewayDataIPErr != nil {
		gatewayDataIP = ""
		klog.V(2).Infof("skip shared incluster templates, app-gateway-data ClusterIP unavailable: %v", gatewayDataIPErr)
	}

	var sharedInclusterTemplatePlugins []*corefile.Plugin
	if inClusterGatewayEnabled(ctx, dynamicClient) {
		srrEntrances, err := sharedInclusterEntrancesFromCluster(ctx, kubeClient, dynamicClient)
		if err != nil {
			// degrade: skip shared templates, keep regenerating the rest.
			// A transient SRR list failure (e.g. RBAC lag, informer thrash)
			// must not freeze the whole Corefile. Leave the shared template
			// plugins nil so user wildcard and other zones still update this
			// round; the shared enhancement converges on the next reconcile.
			klog.Errorf("degrade: skip shared incluster templates, list SRR error: %v", err)
		} else if gatewayDataIP != "" {
			sharedInclusterTemplatePlugins = buildSharedInclusterTemplates(srrEntrances, gatewayDataIP)
		}
	} else {
		klog.V(2).Info("skip shared incluster CoreDNS templates: inClusterGatewayEnabled=false")
	}

	v2DirectSharedTemplatePlugins, v2Err := v2DirectSharedTemplatesFromCluster(ctx, kubeClient, dynamicClient, userList.Items)
	if v2Err != nil {
		klog.Errorf("degrade: skip v2 direct shared templates, scan applications error: %v", v2Err)
	}

	var adguardIp string
	pods, err := kubeClient.CoreV1().Pods("").List(ctx, metav1.ListOptions{LabelSelector: "applications.app.bytetrade.io/name=adguardhome"})
	if err != nil {
		klog.Error("get adguardhome pod error, ", err)
	} else {
		if len(pods.Items) > 0 {
			adguardIp = pods.Items[0].Status.PodIP
		}
	}

	podCIDR := detectPodCIDR(ctx, kubeClient, dynamicClient)
	clusterCIDR := clusterCIDRFromPod(podCIDR)
	klog.Infof("CoreDNS view expr: podCIDR=%s clusterCIDR=%s masterNodeIp=%s", podCIDR, clusterCIDR, masterNodeIp)
	inclusterExpr, vpnExpr := buildViewExprs(podCIDR, clusterCIDR, masterNodeIp, adguardIp)

	inclusterView := &corefile.Plugin{
		Name: "view",
		Args: []string{"incluster"},
		Options: []*corefile.Option{
			{
				Name: "expr",
				Args: []string{inclusterExpr},
			},
		},
	}

	vpnView := &corefile.Plugin{
		Name: "view",
		Args: []string{"vpn"},
		Options: []*corefile.Option{
			{
				Name: "expr",
				Args: []string{vpnExpr},
			},
		},
	}

	// Incluster template order (dual-track):
	//   v3 SRR exact templates → user-zone wildcards → v2 legacy shared templates.
	// v3 must stay before wildcards so options.shared apps keep gateway FQDN precedence.
	// v2 follows release-1.12.5 placement after wildcards.
	inclusterPluginsWithSharedTemplates := inclusterTemplatesPlugins
	if len(sharedInclusterTemplatePlugins) > 0 || len(v2DirectSharedTemplatePlugins) > 0 {
		var inclusterPluginChain []*corefile.Plugin
		inclusterPluginChain = append(inclusterPluginChain, sharedInclusterTemplatePlugins...)
		inclusterPluginChain = append(inclusterPluginChain, inclusterTemplatesPlugins...)
		inclusterPluginChain = append(inclusterPluginChain, v2DirectSharedTemplatePlugins...)
		inclusterPluginsWithSharedTemplates = inclusterPluginChain
	}
	inclusterPlugins := append(append([]*corefile.Plugin{}, defaultPlugins...), inclusterPluginsWithSharedTemplates...)

	inclusterServer := &corefile.Server{
		DomPorts: defaultsServer.DomPorts,
		Plugins:  append([]*corefile.Plugin{inclusterView}, inclusterPlugins...),
	}

	vpnServer := &corefile.Server{
		DomPorts: defaultsServer.DomPorts,
		Plugins:  append([]*corefile.Plugin{vpnView}, append(defaultPlugins, templatesPlugins...)...),
	}

	otherServer := &corefile.Server{
		DomPorts: defaultsServer.DomPorts,
		Plugins: append(defaultPlugins,
			append(localTemplatesPlugins, localDomainTemplatesPlugins...)...),
	}

	servers := []*corefile.Server{inclusterServer, vpnServer, otherServer}

	nxdomainServer, err := buildBlockLocalSearchServer()
	if err != nil {
		klog.Error("build NXDOMAIN server block error, ", err)
		return err
	}
	if nxdomainServer != nil {
		servers = append(servers, nxdomainServer)
	}

	file.Servers = servers

	newCorefileData := file.ToString()
	newCorefileSize := len(newCorefileData)
	if newCorefileSize >= corefileSizeRejectBytes {
		err := fmt.Errorf("regenerated Corefile size %d exceeds reject threshold %d", newCorefileSize, corefileSizeRejectBytes)
		klog.Error(err)
		return err
	}
	if newCorefileSize >= corefileSizeWarnBytes {
		klog.Warningf(
			"regenerated Corefile size %d exceeds warn threshold %d (reject threshold=%d)",
			newCorefileSize,
			corefileSizeWarnBytes,
			corefileSizeRejectBytes,
		)
	}
	corefileConfigMap.Data["Corefile"] = newCorefileData

	_, err = kubeClient.CoreV1().ConfigMaps("kube-system").Update(ctx, corefileConfigMap, metav1.UpdateOptions{})
	if err != nil {
		klog.Error("update coredns configmap error, ", err)
		return err
	}

	klog.Info("coredns corefile regenerated successfully")
	return nil
}

func UpsertCorefile(data, userzone, ip string) (string, error) {
	file, err := corefile.New(data)
	if err != nil {
		klog.Error("parse corefile error, ", err)
		return "", err
	}

	if len(file.Servers) != 1 {
		klog.Warning("invalid corefile configuration")
		return data, nil
	}

	var newPlugins []*corefile.Plugin
	found := false
	newOptions := []*corefile.Option{
		{
			Name: "match",
			Args: []string{fmt.Sprintf("\"\\w*\\.?(%s\\.)$\"", userzone)},
		},
		{
			Name: "answer",
			Args: []string{fmt.Sprintf("\"{{ .Name }} 60 IN A %s\"", ip)},
		},
		{
			Name: "fallthrough",
			Args: []string{},
		},
	}
	anyOptions := []*corefile.Option{
		{
			Name: "rcode",
			Args: []string{"NOERROR"},
		},
	}
	userTemplateArgs := []string{"IN", "A", userzone}
	userTemplateAnyArgs := []string{"IN", "ANY", userzone}

	for _, p := range file.Servers[0].Plugins {
		// only care about template plugins
		if p.Name != "template" {
			newPlugins = append(newPlugins, p)
			continue
		}

		if len(p.Args) != 3 {
			// the template is not added by us, keep it
			klog.Info(p.Args)
			newPlugins = append(newPlugins, p)
			continue
		}

		// update query type A with new options
		if p.Args[2] == userTemplateArgs[2] && p.Args[1] == userTemplateArgs[1] {
			found = true
			p.Options = newOptions
			newPlugins = append(newPlugins, p)
		} else if p.Args[2] == userTemplateAnyArgs[2] && p.Args[1] == userTemplateAnyArgs[1] {
			// update query type ANY with ANY options
			p.Options = anyOptions
			newPlugins = append(newPlugins, p)
		} else {
			// another user's template, keep it
			for _, o := range p.Options {
				switch o.Name {
				case "match", "answer":
					// fix args to one string
					o.Args = []string{fmt.Sprintf("\"%s\"", strings.Join(o.Args, " "))}
				}
			}
			newPlugins = append(newPlugins, p)
		}
	}

	if !found {
		newPlugins = append(newPlugins, &corefile.Plugin{
			Name:    "template",
			Args:    userTemplateArgs,
			Options: newOptions,
		})

		newPlugins = append(newPlugins, &corefile.Plugin{
			Name:    "template",
			Args:    userTemplateAnyArgs,
			Options: anyOptions,
		})
	}

	file.Servers[0].Plugins = newPlugins

	return file.ToString(), nil
}

func RemoveTemplateFromCorefile(data, userzone string) (string, error) {
	file, err := corefile.New(data)
	if err != nil {
		klog.Error("parse corefile error, ", err)
		return "", err
	}

	if len(file.Servers) != 1 {
		klog.Warning("invalid corefile configuration")
		return data, nil
	}

	var newPlugins []*corefile.Plugin
	userTemplateArgs := []string{"IN", "A", userzone}
	for _, p := range file.Servers[0].Plugins {
		// only care about template plugins
		if p.Name != "template" {
			newPlugins = append(newPlugins, p)
			continue
		}

		if len(p.Args) != 3 {
			// the template is not added by us, keep it
			klog.Info(p.Args)
			newPlugins = append(newPlugins, p)
			continue
		}

		if p.Args[2] == userTemplateArgs[2] {
			// remove the template plugin
			continue
		}
	}

	file.Servers[0].Plugins = newPlugins

	return file.ToString(), nil
}

func subDNSSplit(n int64) map[string]net.IP {
	subDNSMap := make(map[string]net.IP)
	log2n := int(math.Ceil(math.Log2(float64(n))))
	alignedN := 1 << log2n
	_, ipNet, _ := net.ParseCIDR("100.64.0.0/10")

	baseIP := ipNet.IP.To4()
	originalMaskLen, _ := ipNet.Mask.Size()

	newMaskLen := originalMaskLen + log2n
	ipsPerSubnet := 1 << (32 - newMaskLen)

	for i := 0; i < alignedN; i++ {
		offset := uint32(i * ipsPerSubnet)
		subnetIP := make(net.IP, 4)
		copy(subnetIP, baseIP)
		for j := 3; j >= 0 && offset > 0; j-- {
			subnetIP[j] += byte(offset & 0xFF)
			offset >>= 8
		}
		firstUsableIP := make(net.IP, 4)
		copy(firstUsableIP, subnetIP)
		firstUsableIP[3]++
		index := strconv.FormatInt(int64(i), 10)
		subDNSMap[index] = firstUsableIP
	}
	return subDNSMap
}

func getUserLocalIp(user *unstructured.Unstructured) (net.IP, error) {
	userIndex, ok := user.GetAnnotations()[UserIndexAna]
	if !ok || userIndex == "" {
		klog.Infof("can not find user index from annotations")
		return nil, nil
	}

	userMaxStr := os.Getenv("OLARES_MAX_USERS")
	if userMaxStr == "" {
		userMaxStr = "1024"
	}
	userMax, err := strconv.ParseInt(userMaxStr, 10, 64)
	if err != nil {
		klog.Infof("parse user index failed %v", err)
		return nil, err
	}
	localIp := subDNSSplit(userMax)[userIndex]
	if localIp == nil || localIp.String() == "" {
		return nil, fmt.Errorf("invalid ip address %v", localIp)
	}
	klog.Infof("localIp: %v", localIp)

	return localIp, nil
}

func getUserIngressIP(ctx context.Context, kubeClient kubernetes.Interface) (string, error) {
	pods, err := kubeClient.CoreV1().Pods("os-network").List(ctx, metav1.ListOptions{
		LabelSelector: "app=l4-bfl-proxy",
	})
	if err != nil {
		klog.Error("get l4 pod error, ", err)
		return "", err
	}
	if len(pods.Items) == 0 {
		return "", errors.New("no l4 proxy pod found")
	}
	pod := pods.Items[0]

	return pod.Status.PodIP, nil
}

func getNonClusterLocalSearchDomains() ([]string, error) {
	data, err := os.ReadFile("/etc/resolv.conf")
	if err != nil {
		return nil, fmt.Errorf("failed to read /etc/resolv.conf: %w", err)
	}

	var domains []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "search") {
			continue
		}
		for _, d := range strings.Fields(line)[1:] {
			if !strings.HasSuffix(d, "cluster.local") && d != "local" {
				domains = append(domains, d)
			}
		}
	}
	return domains, nil
}

func buildBlockLocalSearchServer() (*corefile.Server, error) {
	domains, err := getNonClusterLocalSearchDomains()
	if err != nil {
		return nil, err
	}
	if len(domains) == 0 {
		return nil, nil
	}

	var domPorts []string
	for _, d := range domains {
		domPorts = append(domPorts, d+":53")
	}

	klog.Infof("adding NXDOMAIN server block for search domains: %v", domains)
	return &corefile.Server{
		DomPorts: domPorts,
		Plugins: []*corefile.Plugin{
			{
				Name: "template",
				Args: []string{"ANY", "ANY"},
				Options: []*corefile.Option{
					{
						Name: "rcode",
						Args: []string{"NXDOMAIN"},
					},
				},
			},
		},
	}, nil
}

const UserAnnotationZoneKey = "bytetrade.io/zone"
const UserAnnotationLocalDomainDNSRecord = "bytetrade.io/local-domain-dns-record"
const UserIndexAna = "bytetrade.io/user-index"

// SharedInclusterEntrance identifies one gateway-mode SRR host pattern that
// may be rewritten inside the cluster.
type SharedInclusterEntrance struct {
	AppID          string
	EntranceName   string
	EntranceID     string
	PlatformDomain string
	HostPattern    string
}

const (
	hostPatternSharedExact    = "shared-exact"
	hostPatternViewerWildcard = "viewer-wildcard"
)

// matchRegex returns an anchored host regex for one SRR host pattern:
// - shared exact host:  ^<id>\.shared\.<base>\.$
// - application logical: ^<id>\.[^.]+\.<base>\.$
func (e SharedInclusterEntrance) matchRegex() string {
	entranceID, platformDomain, hostPatternType, ok := parseLogicalHostPattern(e.HostPattern)
	if !ok {
		platformDomain = strings.ToLower(strings.TrimSpace(strings.TrimSuffix(e.PlatformDomain, ".")))
		entranceID = strings.ToLower(strings.TrimSpace(e.EntranceID))
		hostPatternType = hostPatternSharedExact
	}
	if platformDomain == "" || entranceID == "" || hostPatternType == "" {
		return ""
	}
	escapedEntranceID := strings.ReplaceAll(entranceID, ".", `\.`)
	escapedPlatformDomain := strings.ReplaceAll(platformDomain, ".", `\.`)
	switch hostPatternType {
	case hostPatternSharedExact:
		return `"^` + escapedEntranceID + `\.shared\.` + escapedPlatformDomain + `\.$"`
	case hostPatternViewerWildcard:
		return `"^` + escapedEntranceID + `\.[^.]+\.` + escapedPlatformDomain + `\.$"`
	default:
		return ""
	}
}

// buildSharedInclusterTemplates builds CoreDNS `template` plugin instances
// that map every registered gateway-mode SRR host pattern to the app-gateway
// data plane ClusterIP.
//
// rationale: CoreDNS's plugin.cfg orders `template` before `hosts`, so the
// per-user wildcard `template IN A <userzone> { match "\w*\.?(<userzone>\.)$" }`
// would shadow any matching `hosts` entry for `<hash>.shared.<platformDomain>`.
// We therefore emit exact-FQDN `template` instances anchored at the root zone
// (`IN A .`) that match the literal FQDN with a `^…\.$` anchored regex and
// answer with the gateway ClusterIP. `fallthrough` is set so unrelated names
// continue down the chain to the wildcard templates / forward.
//
// requirement: only FQDNs derived from SRR hostPatterns may be rewritten.
// behavior: deterministic sorted ordering by match regex; empty input returns nil.
// test: table-driven unit tests in corefile_incluster_test.go.
func buildSharedInclusterTemplates(entrances []SharedInclusterEntrance, gatewayDataIP string) []*corefile.Plugin {
	ip := net.ParseIP(strings.TrimSpace(gatewayDataIP))
	if ip == nil || ip.To4() == nil {
		return nil
	}
	gatewayDataIP = ip.String()

	seen := make(map[string]struct{})
	var matches []string
	for _, ent := range entrances {
		match := ent.matchRegex()
		if match == "" {
			continue
		}
		if _, ok := seen[match]; ok {
			continue
		}
		seen[match] = struct{}{}
		matches = append(matches, match)
	}
	if len(matches) == 0 {
		return nil
	}
	sort.Strings(matches)

	plugins := make([]*corefile.Plugin, 0, len(matches))
	for _, match := range matches {
		answerArg := `"{{ .Name }} 15 IN A ` + gatewayDataIP + `"`
		plugins = append(plugins, &corefile.Plugin{
			Name: "template",
			Args: []string{"IN", "A", "."},
			Options: []*corefile.Option{
				{Name: "match", Args: []string{match}},
				{Name: "answer", Args: []string{answerArg}},
				{Name: "fallthrough"},
			},
		})
	}
	return plugins
}

func appGatewayDataClusterIP(ctx context.Context, kubeClient kubernetes.Interface) (string, error) {
	svc, err := kubeClient.CoreV1().Services(appGatewayNamespace).Get(ctx, appGatewayDataService, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	if svc.Spec.ClusterIP == "" || svc.Spec.ClusterIP == "None" {
		return "", fmt.Errorf("service %s/%s has no ClusterIP", appGatewayNamespace, appGatewayDataService)
	}
	return svc.Spec.ClusterIP, nil
}

func namespacesWithDNSPassthrough(ctx context.Context, kubeClient kubernetes.Interface) (map[string]struct{}, error) {
	nsList, err := kubeClient.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	out := make(map[string]struct{})
	for i := range nsList.Items {
		if nsList.Items[i].Labels[labelDNSPassthrough] == "true" {
			out[nsList.Items[i].Name] = struct{}{}
		}
	}
	return out, nil
}

func parseLogicalHostPattern(pattern string) (entranceID, platformDomain, hostPatternType string, ok bool) {
	pattern = strings.ToLower(strings.TrimSpace(pattern))
	if pattern == "" {
		return "", "", "", false
	}
	pattern = strings.TrimSuffix(pattern, ".")

	const sharedMarker = ".shared."
	if idx := strings.Index(pattern, sharedMarker); idx > 0 && len(pattern) > idx+len(sharedMarker) {
		entranceID = strings.TrimSpace(pattern[:idx])
		if strings.Contains(entranceID, ".") || entranceID == "" || strings.Contains(entranceID, "*") {
			return "", "", "", false
		}
		platformDomain = strings.TrimSpace(strings.TrimSuffix(pattern[idx+len(sharedMarker):], "."))
		if platformDomain == "" {
			return "", "", "", false
		}
		return entranceID, platformDomain, hostPatternSharedExact, true
	}

	const viewerWildcardMarker = ".*."
	if idx := strings.Index(pattern, viewerWildcardMarker); idx > 0 && len(pattern) > idx+len(viewerWildcardMarker) {
		entranceID = strings.TrimSpace(pattern[:idx])
		if strings.Contains(entranceID, ".") || entranceID == "" || strings.Contains(entranceID, "*") {
			return "", "", "", false
		}
		platformDomain = strings.TrimSpace(strings.TrimSuffix(pattern[idx+len(viewerWildcardMarker):], "."))
		if platformDomain == "" || strings.Contains(platformDomain, "*") {
			return "", "", "", false
		}
		return entranceID, platformDomain, hostPatternViewerWildcard, true
	}
	return "", "", "", false
}

func sharedInclusterEntrancesFromSRRItems(
	srrItems []unstructured.Unstructured,
	passthrough map[string]struct{},
) []SharedInclusterEntrance {
	if len(srrItems) == 0 {
		return nil
	}
	var entrances []SharedInclusterEntrance
	for i := range srrItems {
		srr := &srrItems[i]
		if _, skip := passthrough[srr.GetNamespace()]; skip {
			continue
		}
		routeMode, found, _ := unstructured.NestedString(srr.Object, "spec", "routeMode")
		if found && routeMode != "" && routeMode != srrRouteModeGateway {
			continue
		}
		labels := srr.GetLabels()
		appid := strings.ToLower(strings.TrimSpace(labels[labelSRRAppID]))
		entranceName := strings.ToLower(strings.TrimSpace(labels[labelSRREntrance]))
		if appid == "" || entranceName == "" {
			continue
		}
		patterns, found, err := unstructured.NestedStringSlice(srr.Object, "spec", "hostPatterns")
		if err != nil || !found || len(patterns) == 0 {
			continue
		}
		entranceID := ""
		platformDomain := ""
		hostPattern := ""
		for _, pattern := range patterns {
			id, domain, _, ok := parseLogicalHostPattern(pattern)
			if !ok {
				continue
			}
			entranceID = id
			platformDomain = domain
			hostPattern = strings.ToLower(strings.TrimSpace(strings.TrimSuffix(pattern, ".")))
			break
		}
		if platformDomain == "" || entranceID == "" || hostPattern == "" {
			continue
		}
		entrances = append(entrances, SharedInclusterEntrance{
			AppID:          appid,
			EntranceName:   entranceName,
			EntranceID:     entranceID,
			PlatformDomain: platformDomain,
			HostPattern:    hostPattern,
		})
	}
	return entrances
}

func sharedInclusterEntrancesFromCluster(
	ctx context.Context,
	kubeClient kubernetes.Interface,
	dynamicClient dynamic.Interface,
) ([]SharedInclusterEntrance, error) {
	passthrough, err := namespacesWithDNSPassthrough(ctx, kubeClient)
	if err != nil {
		return nil, err
	}
	srrList, err := dynamicClient.Resource(sharedRouteRegistryGVR).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return sharedInclusterEntrancesFromSRRItems(srrList.Items, passthrough), nil
}

func normalizeReloadInDefaultPlugins(defaultPlugins []*corefile.Plugin) []*corefile.Plugin {
	found := false
	for _, p := range defaultPlugins {
		if p == nil || p.Name != "reload" {
			continue
		}
		p.Args = []string{"5s"}
		found = true
	}
	if !found {
		defaultPlugins = append(defaultPlugins, &corefile.Plugin{
			Name: "reload",
			Args: []string{"5s"},
		})
	}
	return defaultPlugins
}

// appIDFromApplication mirrors appcfg.AppName(name).GetAppID() without a framework import.
func appIDFromApplication(app *application.Application) string {
	if app.Spec.IsSysApp {
		return app.Spec.Name
	}
	sum := md5.Sum([]byte(app.Spec.Name))
	return hex.EncodeToString(sum[:])[:8]
}

// legacySharedEntrancePrefix matches release-1.12.5 v2 shared CoreDNS ID base.
func legacySharedEntrancePrefix(appid string) string {
	sum := md5.Sum([]byte(appid + "shared"))
	return hex.EncodeToString(sum[:])[:8]
}

// legacySharedEntranceID always appends the loop index (single entrance uses …0).
func legacySharedEntranceID(appid string, entranceIndex int) string {
	return fmt.Sprintf("%s%d", legacySharedEntrancePrefix(appid), entranceIndex)
}

func isV3SharedApp(app *application.Application) bool {
	return app.GetLabels()[labelAppShared] == labelAppSharedTrue
}

func platformDomainFromUserZone(userZone string) (string, bool) {
	tokens := strings.Split(userZone, ".")
	if len(tokens) < 2 {
		return "", false
	}
	return strings.Join(tokens[1:], "."), true
}

func escapeRegexLabel(s string) string {
	return strings.ReplaceAll(s, ".", `\.`)
}

func legacySharedNamespacesForApp(app *application.Application, nsList []corev1.Namespace) []*corev1.Namespace {
	var sharedNs []*corev1.Namespace
	for i := range nsList {
		ns := &nsList[i]
		refAppName := ns.Labels[labelNSAppName]
		sharedNamespace := ns.Labels[labelNSShared]
		installedUser := ns.Labels[labelNSInstallUser]
		if refAppName == app.Spec.Name && sharedNamespace == "true" && installedUser == app.Spec.Owner {
			sharedNs = append(sharedNs, ns)
		}
	}
	return sharedNs
}

func buildV2DirectSharedTemplate(prefix string, entranceIndex int, sharedZone, clusterIP string) *corefile.Plugin {
	legacyID := fmt.Sprintf("%s%d", prefix, entranceIndex)
	fqdn := fmt.Sprintf("%s.%s", legacyID, sharedZone)
	match := fmt.Sprintf(`"%s%d.?(%s\.)$"`, prefix, entranceIndex, sharedZone)
	return &corefile.Plugin{
		Name: "template",
		Args: []string{"IN", "A", fqdn},
		Options: []*corefile.Option{
			{Name: "match", Args: []string{match}},
			{Name: "answer", Args: []string{fmt.Sprintf(`"{{ .Name }} 60 IN A %s"`, clusterIP)}},
			{Name: "fallthrough"},
		},
	}
}

func v2DirectSharedTemplatesFromCluster(
	ctx context.Context,
	kubeClient kubernetes.Interface,
	dynamicClient dynamic.Interface,
	users []unstructured.Unstructured,
) ([]*corefile.Plugin, error) {
	if len(users) == 0 {
		klog.Info("no users found, skip adding shared entrance dns records")
		return nil, nil
	}

	zone := users[0].GetAnnotations()[UserAnnotationZoneKey]
	if len(zone) == 0 {
		klog.Info("no zone annotation found in user, skip adding shared entrance dns records")
		return nil, nil
	}
	tokens := strings.Split(zone, ".")
	if len(tokens) < 2 {
		klog.Info("invalid zone annotation found in user, skip adding shared entrance dns records")
		return nil, nil
	}
	tokens[0] = "shared"
	sharedZone := strings.Join(tokens, ".")

	appList, err := dynamicClient.Resource(application.GVR).List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.Error("get applications error, ", err)
		return nil, err
	}

	nsList, err := kubeClient.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.Error("list namespaces error, ", err)
		return nil, err
	}

	var plugins []*corefile.Plugin
	for i := range appList.Items {
		var app application.Application
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(appList.Items[i].Object, &app); err != nil {
			klog.Error("convert obj error, ", err)
			continue
		}
		if len(app.Spec.SharedEntrances) == 0 {
			continue
		}
		if isV3SharedApp(&app) {
			continue
		}

		sharedNs := legacySharedNamespacesForApp(&app, nsList.Items)
		prefix := legacySharedEntrancePrefix(app.Spec.Appid)
		for entranceIndex, entrance := range app.Spec.SharedEntrances {
			for _, ns := range sharedNs {
				svc, err := kubeClient.CoreV1().Services(ns.Name).Get(ctx, entrance.Host, metav1.GetOptions{})
				if err != nil {
					klog.Error("get shared entrance service error, ", err)
					continue
				}
				entranceIP := svc.Spec.ClusterIP
				if entranceIP == "" {
					klog.Info("shared entrance has no ingress ip, skip corefile update")
					continue
				}
				plugins = append(plugins, buildV2DirectSharedTemplate(prefix, entranceIndex, sharedZone, entranceIP))
			}
		}
	}
	if len(plugins) == 0 {
		return nil, nil
	}
	return plugins, nil
}

func buildViewExprs(podCIDR, clusterCIDR, masterNodeIp, adguardIp string) (inclusterExpr, vpnExpr string) {
	inclusterExpr = fmt.Sprintf("incidr(client_ip(), '%s')", podCIDR)
	if adguardIp != "" {
		inclusterExpr = fmt.Sprintf("( %s && client_ip() != '%s' )", inclusterExpr, adguardIp)
	}
	vpnExpr = fmt.Sprintf(
		"incidr(client_ip(), '100.64.0.0/16') || client_ip() == '%s' || incidr(client_ip(), '%s')",
		masterNodeIp, clusterCIDR,
	)
	return inclusterExpr, vpnExpr
}

func clusterCIDRFromPod(podCIDR string) string {
	ip, _, err := net.ParseCIDR(podCIDR)
	if err != nil {
		return clusterCIDRFromPod(defaultPodCIDR)
	}
	v4 := ip.To4()
	if v4 == nil {
		return clusterCIDRFromPod(defaultPodCIDR)
	}
	return fmt.Sprintf("%d.%d.0.0/16", v4[0], v4[1])
}

func detectPodCIDR(ctx context.Context, kubeClient kubernetes.Interface, dynamicClient dynamic.Interface) string {
	if cidr, ok := podCIDRFromIptablesProbe(); ok {
		klog.V(2).Infof("detectPodCIDR: iptables KUBE-SERVICES %s", cidr)
		return cidr
	}
	if cidr, ok := podCIDRFromCalicoIPPools(ctx, dynamicClient); ok {
		klog.V(2).Infof("detectPodCIDR: Calico IPPool %s", cidr)
		return cidr
	}
	if cidr, ok := podCIDRFromKubeClusterCIDRArg(ctx, kubeClient); ok {
		klog.V(2).Infof("detectPodCIDR: kube --cluster-cidr %s", cidr)
		return cidr
	}
	klog.V(2).Infof("detectPodCIDR: fallback %s", defaultPodCIDR)
	return defaultPodCIDR
}

// podCIDRFromIptablesProbe is swapped in unit tests to avoid host iptables dependence.
var podCIDRFromIptablesProbe = podCIDRFromIptables

func podCIDRFromIptables() (string, bool) {
	out, err := exec.Command("iptables", "-t", "nat", "-S", "KUBE-SERVICES").CombinedOutput()
	if err != nil {
		klog.V(4).Infof("detectPodCIDR: iptables unavailable: %v", err)
		return "", false
	}
	return parsePodCIDRFromKubeServicesIPTables(string(out))
}

func parsePodCIDRFromKubeServicesIPTables(output string) (string, bool) {
	for _, line := range strings.Split(output, "\n") {
		if !strings.Contains(line, "KUBE-SERVICES") || !strings.Contains(line, "KUBE-MARK-MASQ") {
			continue
		}
		fields := strings.Fields(line)
		for i := 0; i < len(fields)-2; i++ {
			if fields[i] != "!" || fields[i+1] != "-s" {
				continue
			}
			cidr := fields[i+2]
			if _, _, err := net.ParseCIDR(cidr); err == nil {
				return cidr, true
			}
		}
	}
	return "", false
}

func podCIDRFromCalicoIPPools(ctx context.Context, dynamicClient dynamic.Interface) (string, bool) {
	if dynamicClient == nil {
		return "", false
	}
	list, err := dynamicClient.Resource(calicoIPPoolGVR).List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.V(4).Infof("detectPodCIDR: list Calico IPPools failed: %v", err)
		return "", false
	}
	for _, item := range list.Items {
		cidr, found, err := unstructured.NestedString(item.Object, "spec", "cidr")
		if err != nil || !found || cidr == "" {
			continue
		}
		if _, _, err := net.ParseCIDR(cidr); err == nil {
			return cidr, true
		}
	}
	return "", false
}

func podCIDRFromKubeClusterCIDRArg(ctx context.Context, kubeClient kubernetes.Interface) (string, bool) {
	if kubeClient == nil {
		return "", false
	}
	pods, err := kubeClient.CoreV1().Pods("kube-system").List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.V(4).Infof("detectPodCIDR: list kube-system pods failed: %v", err)
		return "", false
	}
	for _, pod := range pods.Items {
		for _, c := range pod.Spec.Containers {
			if cidr, ok := clusterCIDRArgFromCommand(c.Command, c.Args); ok {
				return cidr, true
			}
			if cidr, ok := clusterCIDRFromK3SArgsEnv(c.Env); ok {
				return cidr, true
			}
		}
		for _, c := range pod.Spec.InitContainers {
			if cidr, ok := clusterCIDRArgFromCommand(c.Command, c.Args); ok {
				return cidr, true
			}
		}
	}
	return "", false
}

func clusterCIDRFromK3SArgsEnv(env []corev1.EnvVar) (string, bool) {
	for _, e := range env {
		if e.Name != "K3S_ARGS" || e.Value == "" {
			continue
		}
		for _, token := range strings.Fields(e.Value) {
			if strings.HasPrefix(token, "--cluster-cidr=") {
				cidr := strings.TrimPrefix(token, "--cluster-cidr=")
				if _, _, err := net.ParseCIDR(cidr); err == nil {
					return cidr, true
				}
			}
		}
	}
	return "", false
}

func clusterCIDRArgFromCommand(command, args []string) (string, bool) {
	all := append(append([]string{}, command...), args...)
	for i, arg := range all {
		if strings.HasPrefix(arg, "--cluster-cidr=") {
			cidr := strings.TrimPrefix(arg, "--cluster-cidr=")
			if _, _, err := net.ParseCIDR(cidr); err == nil {
				return cidr, true
			}
		}
		if arg == "--cluster-cidr" && i+1 < len(all) {
			if _, _, err := net.ParseCIDR(all[i+1]); err == nil {
				return all[i+1], true
			}
		}
	}
	return "", false
}

// CorefileSRRSubscriber regenerates CoreDNS when SharedRouteRegistry changes.
type CorefileSRRSubscriber struct {
	*Subscriber
	kubeClient    kubernetes.Interface
	dynamicClient dynamic.Interface
}

func (s *CorefileSRRSubscriber) HandleEvent() cache.ResourceEventHandler {
	enqueue := func(obj interface{}) {
		s.Watchers.Enqueue(EnqueueObj{
			Subscribe: s,
			Obj:       obj,
			Action:    UPDATE,
		})
	}
	return cache.ResourceEventHandlerFuncs{
		AddFunc:    enqueue,
		UpdateFunc: func(_, newObj interface{}) { enqueue(newObj) },
		DeleteFunc: enqueue,
	}
}

func (s *CorefileSRRSubscriber) Do(ctx context.Context, obj interface{}, action Action) error {
	_ = obj
	_ = action
	return RegenerateCorefile(ctx, s.kubeClient, s.dynamicClient)
}

// RegisterCorefileSRRWatcher lists SharedRouteRegistry and triggers RegenerateCorefile on changes.
func RegisterCorefileSRRWatcher(w *Watchers, kubeClient kubernetes.Interface, dynamicClient dynamic.Interface) error {
	sub := &CorefileSRRSubscriber{
		Subscriber:    NewSubscriber(w),
		kubeClient:    kubeClient,
		dynamicClient: dynamicClient,
	}
	return AddToWatchers[unstructured.Unstructured](w, sharedRouteRegistryGVR, sub.HandleEvent())
}

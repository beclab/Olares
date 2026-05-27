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
	"sort"
	"strconv"
	"strings"

	"github.com/coredns/corefile-migration/migration/corefile"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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
	appGatewayNamespace     = "app-gateway"
	appGatewayDataService   = "app-gateway-data"
	srrRouteModeGateway     = "gateway"
)

var sharedRouteRegistryGVR = schema.GroupVersionResource{
	Group:    "gateway.olares.io",
	Version:  "v1alpha1",
	Resource: "sharedrouteregistries",
}

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

	gatewayDataIP, err := appGatewayDataClusterIP(ctx, kubeClient)
	if err != nil {
		klog.Error("get app-gateway-data ClusterIP error, ", err)
		return err
	}

	var sharedInclusterTemplatePlugins []*corefile.Plugin
	if inClusterGatewayEnabled(ctx, dynamicClient) {
		srrEntrances, err := sharedInclusterEntrancesFromCluster(ctx, kubeClient, dynamicClient, userList)
		if err != nil {
			klog.Error("list shared incluster entrances from SRR error, ", err)
			return err
		}
		if gatewayDataIP != "" {
			sharedInclusterTemplatePlugins = buildSharedInclusterTemplates(srrEntrances, gatewayDataIP)
		}
	} else {
		klog.V(2).Info("skip shared incluster CoreDNS templates: inClusterGatewayEnabled=false")
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

	inclusterExpr := "incidr(client_ip(), '10.233.0.0/16')"
	if adguardIp != "" {
		inclusterExpr = fmt.Sprintf("( %s && client_ip() != '%s' )", inclusterExpr, adguardIp)
	}

	vpnExpr := fmt.Sprintf("incidr(client_ip(), '100.64.0.0/16') || client_ip() == '%s'", masterNodeIp)

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

	// CoreDNS plugin chain orders `template` before `hosts`, and the user-zone
	// wildcard template (e.g. \w*\.?brucedai\.olares\.com\.$) would shadow any
	// `hosts` entry for shared FQDNs. Shared mappings are therefore emitted as
	// exact-match `template` instances inserted BEFORE the user wildcards in
	// the incluster server block, with fallthrough so other names still hit
	// the wildcards.
	inclusterPluginsWithSharedTemplates := inclusterTemplatesPlugins
	if len(sharedInclusterTemplatePlugins) > 0 {
		inclusterPluginsWithSharedTemplates = append(
			append([]*corefile.Plugin{}, sharedInclusterTemplatePlugins...),
			inclusterTemplatesPlugins...,
		)
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

// SharedInclusterEntrance identifies one Shared entrance FQDN that may be rewritten
// inside the cluster. Callers must expand SRR hostPatterns with per-viewer FQDNs
// before invoking buildSharedInclusterHosts.
type SharedInclusterEntrance struct {
	AppID          string
	EntranceName   string
	Viewer         string
	PlatformDomain string
}

// sharedEntranceHostPrefix returns the stable 8-hex label for a shared entrance
// (md5(appid + ":shared:" + entranceName)[:8]), matching app-service appcfg.
func sharedEntranceHostPrefix(appid, entranceName string) string {
	appid = strings.ToLower(strings.TrimSpace(appid))
	entranceName = strings.ToLower(strings.TrimSpace(entranceName))
	sum := md5.Sum([]byte(appid + ":shared:" + entranceName))
	return hex.EncodeToString(sum[:])[:8]
}

// fqdn returns the exact host name for this shared entrance and viewer.
func (e SharedInclusterEntrance) fqdn() string {
	viewer := strings.ToLower(strings.TrimSpace(e.Viewer))
	platformDomain := strings.ToLower(strings.TrimSpace(strings.TrimSuffix(e.PlatformDomain, ".")))
	if viewer == "" || platformDomain == "" {
		return ""
	}
	prefix := sharedEntranceHostPrefix(e.AppID, e.EntranceName)
	if prefix == "" {
		return ""
	}
	return prefix + "." + viewer + "." + platformDomain
}

// buildSharedInclusterTemplates builds CoreDNS `template` plugin instances
// that map every registered Shared entrance FQDN to the app-gateway data
// plane ClusterIP.
//
// rationale: CoreDNS's plugin.cfg orders `template` before `hosts`, so the
// per-user wildcard `template IN A <userzone> { match "\w*\.?(<userzone>\.)$" }`
// would shadow any matching `hosts` entry for `<hash>.<viewer>.<platformDomain>`.
// We therefore emit exact-FQDN `template` instances anchored at the root zone
// (`IN A .`) that match the literal FQDN with a `^…\.$` anchored regex and
// answer with the gateway ClusterIP. `fallthrough` is set so unrelated names
// continue down the chain to the wildcard templates / forward.
//
// requirement: only FQDNs derived from Shared entrances may be rewritten;
// per-user single-entrance hostnames must never be matched by regex.
// behavior: deterministic sorted ordering by FQDN; empty input returns nil.
// test: table-driven unit tests in corefile_incluster_test.go.
func buildSharedInclusterTemplates(entrances []SharedInclusterEntrance, gatewayDataIP string) []*corefile.Plugin {
	ip := net.ParseIP(strings.TrimSpace(gatewayDataIP))
	if ip == nil || ip.To4() == nil {
		return nil
	}
	gatewayDataIP = ip.String()

	seen := make(map[string]struct{})
	var hosts []string
	for _, ent := range entrances {
		host := ent.fqdn()
		if host == "" {
			continue
		}
		if _, ok := seen[host]; ok {
			continue
		}
		seen[host] = struct{}{}
		hosts = append(hosts, host)
	}
	if len(hosts) == 0 {
		return nil
	}
	sort.Strings(hosts)

	plugins := make([]*corefile.Plugin, 0, len(hosts))
	for _, h := range hosts {
		matchArg := `"^` + strings.ReplaceAll(h, ".", `\.`) + `\.$"`
		answerArg := `"{{ .Name }} 60 IN A ` + gatewayDataIP + `"`
		plugins = append(plugins, &corefile.Plugin{
			Name: "template",
			Args: []string{"IN", "A", "."},
			Options: []*corefile.Option{
				{Name: "match", Args: []string{matchArg}},
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

func viewersFromUserList(userList *unstructured.UnstructuredList) []string {
	if userList == nil {
		return nil
	}
	seen := make(map[string]struct{})
	var viewers []string
	for i := range userList.Items {
		name := strings.ToLower(strings.TrimSpace(userList.Items[i].GetName()))
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		viewers = append(viewers, name)
	}
	sort.Strings(viewers)
	return viewers
}

func parseLogicalHostPattern(pattern string) (hash8, platformDomain string, ok bool) {
	pattern = strings.ToLower(strings.TrimSpace(pattern))
	const marker = ".*."
	idx := strings.Index(pattern, marker)
	if idx != 8 || len(pattern) <= idx+len(marker) {
		return "", "", false
	}
	hash8 = pattern[:8]
	for _, c := range hash8 {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return "", "", false
		}
	}
	platformDomain = strings.TrimSuffix(pattern[idx+len(marker):], ".")
	if platformDomain == "" {
		return "", "", false
	}
	return hash8, platformDomain, true
}

func sharedInclusterEntrancesFromSRRItems(
	srrItems []unstructured.Unstructured,
	passthrough map[string]struct{},
	viewers []string,
) []SharedInclusterEntrance {
	if len(srrItems) == 0 || len(viewers) == 0 {
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
		expectedHash := sharedEntranceHostPrefix(appid, entranceName)
		platformDomain := ""
		for _, pattern := range patterns {
			hash8, domain, ok := parseLogicalHostPattern(pattern)
			if !ok || hash8 != expectedHash {
				continue
			}
			platformDomain = domain
			break
		}
		if platformDomain == "" {
			continue
		}
		for _, viewer := range viewers {
			entrances = append(entrances, SharedInclusterEntrance{
				AppID:          appid,
				EntranceName:   entranceName,
				Viewer:         viewer,
				PlatformDomain: platformDomain,
			})
		}
	}
	return entrances
}

func sharedInclusterEntrancesFromCluster(
	ctx context.Context,
	kubeClient kubernetes.Interface,
	dynamicClient dynamic.Interface,
	userList *unstructured.UnstructuredList,
) ([]SharedInclusterEntrance, error) {
	passthrough, err := namespacesWithDNSPassthrough(ctx, kubeClient)
	if err != nil {
		return nil, err
	}
	viewers := viewersFromUserList(userList)
	if len(viewers) == 0 {
		return nil, nil
	}
	srrList, err := dynamicClient.Resource(sharedRouteRegistryGVR).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return sharedInclusterEntrancesFromSRRItems(srrList.Items, passthrough, viewers), nil
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

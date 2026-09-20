package systemcomponents

// Config carries the few namespaces a vendored build is allowed to relocate.
// Empty fields fall back to the stock Olares namespaces, which is what the
// daemon and every non vendored build use.
type Config struct {
	// MeshNamespace holds the Linkerd control plane. See
	// framework/app-gateway/pkg/config.LinkerdNamespace.
	MeshNamespace string

	// GatewayNamespace holds the Envoy Gateway control plane and its data
	// plane. See framework/app-gateway/pkg/config.Namespace.
	GatewayNamespace string
}

func (c Config) withDefaults() Config {
	if c.MeshNamespace == "" {
		c.MeshNamespace = NamespaceOsMesh
	}
	if c.GatewayNamespace == "" {
		c.GatewayNamespace = NamespaceOsGateway
	}
	return c
}

// Default returns every component of a stock Olares installation.
func Default() []Component {
	return All(Config{})
}

// All returns every component that makes up an Olares installation.
//
// Keeping this list explicit rather than deriving it from the installed Helm
// releases is deliberate: a component that was never created is exactly the
// failure this check exists to catch, and several workloads (velero, citus,
// kvrocks, the CNI and the monitoring stack) are not Helm managed at all.
//
// When a component is added to or removed from the install wizard, it has to be
// added here too.
func All(cfg Config) []Component {
	cfg = cfg.withDefaults()

	var components []Component
	components = append(components, clusterInfrastructure()...)
	components = append(components, gpu()...)
	components = append(components, kubeSphere()...)
	components = append(components, osPlatform()...)
	components = append(components, osFramework()...)
	components = append(components, osProtected()...)
	components = append(components, gateway(cfg.GatewayNamespace)...)
	components = append(components, mesh(cfg.MeshNamespace)...)
	components = append(components, network()...)
	components = append(components, perUser()...)
	return components
}

// clusterInfrastructure covers what the CLI's network and storage plugins
// install plus what the Kubernetes distribution ships.
//
// Only CoreDNS is required. Olares always deploys calico and multus, and takes
// the local volume provisioner from the k3s and kubeadm cluster phases, but the
// macOS minikube cluster has none of them and brings its own networking and
// storage instead. The other CNIs inherited from kubekey are left out because
// no installer flag can select them.
func clusterInfrastructure() []Component {
	return []Component{
		{Namespace: NamespaceKubeSystem, Kind: Deployment, Name: "coredns"},
		{Namespace: NamespaceKubeSystem, Kind: Deployment, Name: "openebs-localpv-provisioner", Presence: Optional},
		{Namespace: NamespaceKubeSystem, Kind: Deployment, Name: "calico-kube-controllers", Presence: Optional},
		{Namespace: NamespaceKubeSystem, Kind: DaemonSet, Name: "calico-node", Presence: Optional},
		{Namespace: NamespaceKubeSystem, Kind: DaemonSet, Name: "kube-multus-ds", Presence: Optional},
	}
}

// gpu covers the accelerator plugins, which depend on the hardware present at
// install time: the hami stack and gpu-scheduler on NVIDIA hosts, the Intel GPU
// plugin with its node feature discovery workloads, and the AMD device plugin.
// All are optional here; the dedicated GPU pipeline in pkg/gpu is what verifies
// that a GPU the user actually asked for came up.
func gpu() []Component {
	return []Component{
		{Namespace: NamespaceKubeSystem, Kind: Deployment, Name: "hami-scheduler", Presence: Optional},
		{Namespace: NamespaceKubeSystem, Kind: Deployment, Name: "hami-webui", Presence: Optional},
		{Namespace: NamespaceKubeSystem, Kind: DaemonSet, Name: "hami-device-plugin", Presence: Optional},
		{Namespace: NamespaceKubeSystem, Kind: DaemonSet, Name: "hami-nvidia-dcgm-exporter", Presence: Optional},
		{Namespace: NamespaceOsGpu, Kind: DaemonSet, Name: "gpu-scheduler", Presence: Optional},

		{Namespace: NamespaceKubeSystem, Kind: DaemonSet, Name: "intel-gpu-plugin", Presence: Optional},
		{Namespace: NamespaceKubeSystem, Kind: Deployment, Name: "nfd-master", Presence: Optional},
		{Namespace: NamespaceKubeSystem, Kind: Deployment, Name: "nfd-gc", Presence: Optional},
		{Namespace: NamespaceKubeSystem, Kind: DaemonSet, Name: "nfd-worker", Presence: Optional},

		{Namespace: NamespaceKubeSystem, Kind: DaemonSet, Name: "amdgpu-device-plugin", Presence: Optional},
	}
}

func kubeSphere() []Component {
	return []Component{
		{Namespace: NamespaceKubesphereSystem, Kind: Deployment, Name: "ks-apiserver"},
		{Namespace: NamespaceKubesphereMonitoringSystem, Kind: Deployment, Name: "prometheus-operator"},
		{Namespace: NamespaceKubesphereMonitoringSystem, Kind: Deployment, Name: "kube-state-metrics"},
		{Namespace: NamespaceKubesphereMonitoringSystem, Kind: StatefulSet, Name: "prometheus-k8s"},
		{Namespace: NamespaceKubesphereMonitoringSystem, Kind: DaemonSet, Name: "node-exporter"},
	}
}

// osPlatform is the os-platform Helm release, plus the two StatefulSets the
// tapr middleware operator creates from the PGCluster and RedixCluster CRs.
// Those two carry no Helm annotation and no owner reference, which is why they
// cannot be discovered from the release manifest.
//
// The jfsnotify workloads are templated only when the storage backend is
// JuiceFS.
func osPlatform() []Component {
	return []Component{
		{Namespace: NamespaceOsPlatform, Kind: StatefulSet, Name: "citus"},
		{Namespace: NamespaceOsPlatform, Kind: StatefulSet, Name: "kvrocks"},
		{Namespace: NamespaceOsPlatform, Kind: StatefulSet, Name: "nats"},
		{Namespace: NamespaceOsPlatform, Kind: Deployment, Name: "lldap"},
		{Namespace: NamespaceOsPlatform, Kind: Deployment, Name: "tapr-middleware"},
		{Namespace: NamespaceOsPlatform, Kind: Deployment, Name: "tapr-s3"},
		{Namespace: NamespaceOsPlatform, Kind: Deployment, Name: "argo"},
		{Namespace: NamespaceOsPlatform, Kind: Deployment, Name: "workflow-controller"},
		{Namespace: NamespaceOsPlatform, Kind: Deployment, Name: "kubeblocks"},
		{Namespace: NamespaceOsPlatform, Kind: Deployment, Name: "opa"},
		{Namespace: NamespaceOsPlatform, Kind: Deployment, Name: "otel-opentelemetry-operator"},

		{Namespace: NamespaceOsPlatform, Kind: DaemonSet, Name: "jfsnotify-daemon", Presence: Optional},
		{Namespace: NamespaceOsPlatform, Kind: Deployment, Name: "jfsnotify-proxy", Presence: Optional},
	}
}

// osFramework is the os-framework Helm release. tapr-sysevent is templated by
// the os-platform chart but lands in this namespace, and velero is installed
// directly by the CLI rather than through Helm, on Linux hosts only.
func osFramework() []Component {
	return []Component{
		{Namespace: NamespaceOsFramework, Kind: StatefulSet, Name: AppServiceName},
		{Namespace: NamespaceOsFramework, Kind: Deployment, Name: "authelia-backend"},
		{Namespace: NamespaceOsFramework, Kind: Deployment, Name: "backup"},
		{Namespace: NamespaceOsFramework, Kind: Deployment, Name: "market-deployment"},
		{Namespace: NamespaceOsFramework, Kind: Deployment, Name: "monitoring-server-deployment"},
		{Namespace: NamespaceOsFramework, Kind: Deployment, Name: "notifications-server"},
		{Namespace: NamespaceOsFramework, Kind: Deployment, Name: "seafile"},
		{Namespace: NamespaceOsFramework, Kind: Deployment, Name: "search3"},
		{Namespace: NamespaceOsFramework, Kind: Deployment, Name: "search3-validation"},
		{Namespace: NamespaceOsFramework, Kind: Deployment, Name: "system-provider"},
		{Namespace: NamespaceOsFramework, Kind: Deployment, Name: "tapr-sysevent"},
		{Namespace: NamespaceOsFramework, Kind: Deployment, Name: "vault-server"},
		{Namespace: NamespaceOsFramework, Kind: DaemonSet, Name: "image-service"},
		{Namespace: NamespaceOsFramework, Kind: DaemonSet, Name: "files"},
		{Namespace: NamespaceOsFramework, Kind: DaemonSet, Name: "download"},
		{Namespace: NamespaceOsFramework, Kind: DaemonSet, Name: "osnodeinit-daemon"},
		{Namespace: NamespaceOsFramework, Kind: DaemonSet, Name: "search3monitor"},

		{Namespace: NamespaceOsFramework, Kind: Deployment, Name: "velero", Presence: Optional},
	}
}

func osProtected() []Component {
	return []Component{
		{Namespace: NamespaceOsProtected, Kind: Deployment, Name: "infisical-deployment"},
		{Namespace: NamespaceOsProtected, Kind: Deployment, Name: "integration"},
	}
}

// gateway covers the Envoy Gateway control plane and the data plane it creates.
// The data plane Deployment is named after a hash of the Gateway resource
// (envoy-os-gateway-app-gateway-f5ccc039 on one cluster), so it is matched by
// label instead of by name.
func gateway(namespace string) []Component {
	return []Component{
		{Namespace: namespace, Kind: Deployment, Name: "envoy-gateway"},
		{
			Namespace: namespace, Kind: Deployment,
			Selector: "app.kubernetes.io/component=proxy,app.kubernetes.io/managed-by=envoy-gateway",
		},
	}
}

// mesh returns the Linkerd control plane components.
func mesh(namespace string) []Component {
	return []Component{
		{Namespace: namespace, Kind: Deployment, Name: "linkerd-destination"},
		{Namespace: namespace, Kind: Deployment, Name: "linkerd-identity"},
		{Namespace: namespace, Kind: Deployment, Name: "linkerd-proxy-injector"},
		{Namespace: namespace, Kind: Deployment, Name: "linkerd-pki-guardian"},
	}
}

// network holds the edge proxy BFL applies once the user completes the network
// wizard, into constants.OSSystemNamespace unless L4_PROXY_NAMESPACE overrides
// it. It cannot be required, or the install gate would wait for a step the user
// has not taken yet.
func network() []Component {
	return []Component{
		{Namespace: NamespaceOsNetwork, Kind: Deployment, Name: "l4-bfl-proxy", Presence: Optional},
	}
}

// perUser is instantiated once for every provisioned Olares user. Only the
// workloads listed here are checked: everything else in a user's namespaces is
// an installed app, whose health is the app's own concern and must not fail a
// system readiness check.
//
// Three of them come and go with the user's own state. The wizard is deleted by
// BFL as soon as the user finishes activation, so it only exists on an
// installation nobody has logged into yet. The reverse proxy agent is applied
// while FRP or a Cloudflare tunnel is on and removed again when the user
// switches back to a public IP. jfsnotify-proxy needs the JuiceFS backend.
func perUser() []Component {
	space := NamespaceUserSpacePrefix + UserPlaceholder
	system := NamespaceUserSystemPrefix + UserPlaceholder
	return []Component{
		{Namespace: space, Kind: StatefulSet, Name: LauncherName},
		{Namespace: space, Kind: Deployment, Name: "authelia-deployment"},
		{Namespace: space, Kind: Deployment, Name: "headscale"},
		{Namespace: space, Kind: Deployment, Name: "tailscale"},
		{Namespace: space, Kind: Deployment, Name: "olares-app-deployment"},

		{Namespace: space, Kind: Deployment, Name: "wizard", Presence: Optional},
		{Namespace: space, Kind: Deployment, Name: "reverse-proxy-agent", Presence: Optional},

		{Namespace: system, Kind: Deployment, Name: "system-server"},
		{Namespace: system, Kind: Deployment, Name: "tapr-images"},
		{Namespace: system, Kind: Deployment, Name: "jfsnotify-proxy", Presence: Optional},
	}
}

const (
	// AppServiceName is the workload app-service runs as. The install pipeline
	// waits for it on its own before continuing, because everything installed
	// afterwards is reconciled by it.
	AppServiceName = "app-service"

	// LauncherName is the workload BFL runs as, one per user.
	LauncherName = "bfl"
)

// AppService returns the app-service component on its own.
func AppService() []Component {
	return []Component{{Namespace: NamespaceOsFramework, Kind: StatefulSet, Name: AppServiceName}}
}

// Mesh returns the Linkerd control plane components on their own.
func Mesh(namespace string) []Component {
	if namespace == "" {
		namespace = NamespaceOsMesh
	}
	return mesh(namespace)
}

// Launcher returns the BFL component for a single user.
func Launcher(user string) []Component {
	return []Component{{
		Namespace: NamespaceUserSpacePrefix + user,
		Kind:      StatefulSet,
		Name:      LauncherName,
	}}
}

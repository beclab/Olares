package minikube

type Minikube struct {
	Invalid interface{} `json:"invalid"`
	Valid   []Profile
}

type Profile struct {
	Name              string `json:"Name"`
	Status            string `json:"Status"`
	Config            Config `json:"Config"`
	Active            bool   `json:"Active"`
	ActiveKubeContext bool   `json:"ActiveKubeContext"`
}

type Config struct {
	Name               string `json:"Name"`
	KeepContext        bool
	EmbedCerts         bool
	Memory             int
	CPUs               int
	DiskSize           int64
	Driver             string
	HyperkitVpnKitSock string
	HostOnlyCIDR       string
	SSHUser            string
	SSHKey             string
	SSHPort            int
	KubernetesConfig   KubernetesConfig `json:"KubernetesConfig"`
	Nodes              []Node           `json:"Nodes"`
}

type KubernetesConfig struct {
	KubernetesVersion string
	ClusterName       string
	Namespace         string
	APIServerHAVIP    string
	APIServerName     string
	DNSDomain         string
	ContainerRuntime  string
	CRISocket         string
	NetworkPlugin     string
	ServiceCIDR       string
}

type Node struct {
	Name              string `json:"Name"`
	IP                string `json:"IP"`
	Port              int    `json:"Port"`
	KubernetesVersion string `json:"KubernetesVersion"`
	ContainerRuntime  string `json:"ContainerRuntime"`
	ControlPlane      bool   `json:"ControlPlane"`
	Worker            bool   `json:"Worker"`
}

package nets

type HostsItem struct {
	IP   string `json:"ip"`
	Host string `json:"host"`
}

const MASTER_NODE_HOSTNAME = "master-node"

var (
	internalHostsItem []string = []string{
		".cluster.local",
		"dockerhub.kubekey.local",
		"lb.kubesphere.local",
		"localhost",
		MASTER_NODE_HOSTNAME,
	}
)

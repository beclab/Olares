package nets

type HostsItem struct {
	IP   string `json:"ip"`
	Host string `json:"host"`
}

var (
	internalHostsItem []string = []string{
		".cluster.local",
		"dockerhub.kubekey.local",
		"lb.kubesphere.local",
		"localhost",
	}
)

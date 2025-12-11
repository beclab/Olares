package watchers

import "testing"

func TestCorefile(t *testing.T) {
	data := `
    .:53 {
        errors
        health
        ready
        kubernetes cluster.local in-addr.arpa ip6.arpa {
          pods insecure
          fallthrough in-addr.arpa ip6.arpa
        }
        hosts /node-etc/hosts {
          ttl 30
          reload 10s
          fallthrough
        }
        prometheus :9153
        forward . /etc/resolv.conf
        cache 30
        loop
        reload
        loadbalance
        template IN A cloud007.olares.cn {
            match "\w*\.?(cloud007.olares.cn\.)$"
            answer "{{ .Name }} 3600 IN A 172.16.0.4" 
            fallthrough
        }        
    }
`

	data, err := UpsertCorefile(data, "cloud007.olares.cn", "192.168.0.17")
	if err != nil {
		t.Fail()
	}

	t.Log(data)
}

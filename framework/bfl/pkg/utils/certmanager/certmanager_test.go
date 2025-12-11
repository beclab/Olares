package certmanager

import (
	"testing"

	"bytetrade.io/web3os/bfl/pkg/utils"

	"k8s.io/utils/pointer"
)

var cm = NewCertManager("d1x5")

func TestCertManager(t *testing.T) {
	t.Log("generate cert, and wait for it complete")
	err := cm.GenerateCert()
	if err != nil {
		t.Fatal(err)
	}

	var c *ResponseCert

	t.Log("download cert")
	c, err = cm.DownloadCert()
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("got cert: %v", utils.PrettyJSON(c))

	t.Log("add zone dns record")
	err = cm.AddDNSRecord(nil, pointer.String("frp-bj.snowinning.com"))
	if err != nil {
		t.Fatal(err)
	}

	t.Log("delete zone dns record")
	err = cm.DeleteDNSRecord()
	if err != nil {
		t.Fatal(err)
	}

	t.Log("all done")
}

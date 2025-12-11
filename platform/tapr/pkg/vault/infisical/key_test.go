package infisical

import (
	"context"
	"testing"

	"k8s.io/klog/v2"
)

func TestDecrypt(t *testing.T) {
	mongo := &MongoClient{
		User:     "infisical",
		Password: "ha5lXRdMXDPeWbEr",
		Database: "infisical",
		Addr:     "54.241.136.45:32133",
	}

	_, err := mongo.GetUser(context.Background(), "test@localhost.local")
	if err != nil {
		klog.Error(err)
		t.Fail()
		return
	}

	pk, err := DecryptUserPrivateKeyHelper(nil, "testInfisical1")
	if err != nil {
		t.Log(err)
		t.Fail()
	} else {
		t.Log("user's private key is ", pk)
	}
}

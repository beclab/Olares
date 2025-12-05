package kubesphere_test

import (
	"context"
	"testing"

	"bytetrade.io/web3os/app-service/pkg/kubesphere"
	"k8s.io/klog/v2"
)

func TestGetUserZone(t *testing.T) {
	//config := &rest.Config{
	//	Host:     "52.2.5.188:32480",
	//	Username: "liuyu",
	//	Password: "Test123456",
	//}

	zone, err := kubesphere.GetUserZone(context.TODO(), "liuyu")
	if err != nil {
		klog.Error(err)
		t.FailNow()
	}

	klog.Info("Get user zone, ", zone)
}

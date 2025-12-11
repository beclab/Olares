package infisical

import (
	"context"
	"testing"

	"k8s.io/klog/v2"
)

func TestMongo(t *testing.T) {
	mongo := &MongoClient{
		User:     "infisical",
		Password: "ha5lXRdMXDPeWbEr",
		Database: "infisical",
		Addr:     "54.241.136.45:32133",
	}

	id, err := mongo.GetUser(context.Background(), "test@localhost.local")
	if err != nil {
		klog.Error(err)
		t.Fail()
	} else {
		klog.Info("user id is ", id.ID.Hex())
	}
}

func TestMongoConnect(t *testing.T) {
	mongo := &MongoClient{
		User:     "infisical",
		Password: "ha5lXRdMXDPeWbEr",
		Database: "infisical",
		Addr:     "54.241.136.45:32133",
	}
	err := mongo.TryConnect()
	if err != nil {
		klog.Error(err)
		t.Fail()
	} else {
		klog.Info("connected")
	}
}

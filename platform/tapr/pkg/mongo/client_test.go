package mongo

import (
	"bytetrade.io/web3os/tapr/pkg/apis/apr/v1alpha1"
	"context"
	"testing"
)

func TestCreateUser(t *testing.T) {
	mongo := &MongoClient{
		User:     "root",
		Password: "1CKcF6QDMub8Zy1u",
		Database: "admin",
		Addr:     "54.241.136.45:32133",
	}

	ctx := context.Background()
	err := mongo.Connect(ctx)
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	defer mongo.Close(ctx)
	err = mongo.CreateOrUpdateUserWithDatabase(ctx, "newUser", "pwd123", []v1alpha1.MongoDatabase{{Name: "testdb1"}})
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	t.Log("success")
}

func TestDropUser(t *testing.T) {
	mongo := &MongoClient{
		User:     "root",
		Password: "1CKcF6QDMub8Zy1u",
		Database: "admin",
		Addr:     "54.241.136.45:32133",
	}

	ctx := context.Background()
	err := mongo.Connect(ctx)
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	defer mongo.Close(ctx)
	err = mongo.DropUserAndDatabase(ctx, "newUser", []v1alpha1.MongoDatabase{{Name: "testdb1"}})
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	t.Log("success")
}

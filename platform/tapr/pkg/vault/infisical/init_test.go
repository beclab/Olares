package infisical

import (
	"context"
	"os"
	"testing"
)

func TestCreateUser(t *testing.T) {
	mongo := &MongoClient{
		User:     "infisical",
		Password: "ha5lXRdMXDPeWbEr",
		Database: "infisical",
		Addr:     "54.241.136.45:32133",
	}

	os.Chdir("../../..")

	err := InsertKsUserToMongo(context.TODO(), mongo, "liuyu", "l@l111.com", "Test123456")
	if err != nil {
		t.Log("insert error, ", err)
		t.Fail()
	} else {
		t.Log("success")
	}
}

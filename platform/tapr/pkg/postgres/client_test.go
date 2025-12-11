package postgres

import (
	"context"
	"encoding/json"
	"testing"
)

func TestClient(t *testing.T) {
	c, e := newClient("postgres://postgres:password@54.241.136.45:32432/postgres?sslmode=disable", nil)
	if e != nil {
		t.Log(e)
		t.Fail()
	}

	// res, err := c.DB.Exec("create database if not exists test")
	// if err != nil {
	// 	t.Log(err)
	// 	t.Fail()
	// } else {
	// 	t.Log(res.RowsAffected())
	// }

	res, err := c.findDatabase(context.Background(), "test")
	if err != nil {
		t.Log(err)
		t.Fail()
	} else {
		t.Log("fetch database: ", res)
	}

}

func TestJson(t *testing.T) {
	s := `["a","b"]`
	var a []*string

	err := json.Unmarshal([]byte(s), &a)
	if err != nil {
		panic(err)
	}

	t.Log(a)
}

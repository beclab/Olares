package cache

import (
	"testing"

	"olares.com/backup-server/pkg/util"
)

var c = New()

var user = &struct {
	Name string
	Age  int
}{
	"shark",
	33,
}

func TestCacheSet(t *testing.T) {
	c.Set("a", "11")

	v, ok := c.GetMustString("a")
	t.Logf("v: %v, ok: %v", v, ok)

	c.Set("a", "22")
	v, ok = c.GetMustString("a")
	t.Logf("v: %v, ok: %v", v, ok)

	c.Delete("a")
	v, ok = c.GetMustString("a")
	t.Logf("v: %v, ok: %v", v, ok)

	c.Set("user", user)
	t.Logf("user: %s", util.PrettyJSON(user))
	user.Name = "zl"
	t.Logf("user: %s", util.PrettyJSON(user))

	u, ok := c.Get("user")
	t.Logf("ok: %v, user: %s", ok, util.PrettyJSON(u))
}

func TestDelete(t *testing.T) {
	c.Delete("b")
}

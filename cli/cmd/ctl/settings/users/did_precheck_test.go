package users

import "testing"

func TestDidGateBase(t *testing.T) {
	t.Parallel()
	if got := didGateBase("u@foo.olares.com"); got != "https://api.olares.com/did" {
		t.Fatalf("en: %s", got)
	}
	if got := didGateBase("u@foo.olares.cn"); got != "https://api.olares.cn/did" {
		t.Fatalf("cn: %s", got)
	}
}

func TestDomainSuffixFromOlaresID(t *testing.T) {
	t.Parallel()
	s, err := domainSuffixFromOlaresID("alice@foo.olares.cn")
	if err != nil || s != "foo.olares.cn" {
		t.Fatalf("got %q err %v", s, err)
	}
	if _, err := domainSuffixFromOlaresID("nondomain"); err == nil {
		t.Fatal("expected error")
	}
}

package cluster

import (
	"context"
	"errors"
	"testing"
	"time"

	sysv1alpha1 "github.com/beclab/api/api/sys.bytetrade.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func domainTestClient(t *testing.T, domain string) client.Client {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := sysv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	terminus := &sysv1alpha1.Terminus{ObjectMeta: metav1.ObjectMeta{Name: "terminus"}}
	if domain != "" {
		terminus.Spec.Settings = map[string]string{"domainName": domain}
	}
	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(terminus).Build()
}

func TestGetDomainName(t *testing.T) {
	for _, domain := range []string{"olares.com", "owner.org.olares.com", "example.org", "olares.local", "", " \t", "."} {
		t.Run(domain, func(t *testing.T) {
			got, err := GetDomainName(context.Background(), domainTestClient(t, domain))
			want := domain
			if domain == " \t" || domain == "." {
				want = ""
			}
			if got != want || (err != nil) != (want == "") {
				t.Fatalf("got (%q, %v), want %q", got, err, want)
			}
		})
	}
	got, err := GetDomainName(context.Background(), domainTestClient(t, "  Owner.Example.CN. \n"))
	if err != nil || got != "owner.example.cn" {
		t.Fatalf("normalized domain=(%q, %v)", got, err)
	}
}

type failingDomainReader struct {
	client.Reader
	err error
}

func (r failingDomainReader) Get(context.Context, client.ObjectKey, client.Object, ...client.GetOption) error {
	return r.err
}

func TestGetDomainNameReadErrors(t *testing.T) {
	c := domainTestClient(t, "example.org")
	var terminus sysv1alpha1.Terminus
	if err := c.Get(context.Background(), client.ObjectKey{Name: "terminus"}, &terminus); err != nil {
		t.Fatal(err)
	}
	if err := c.Delete(context.Background(), &terminus); err != nil {
		t.Fatal(err)
	}
	if _, err := GetDomainName(context.Background(), c); err == nil {
		t.Fatal("expected missing Terminus error")
	}
	for _, want := range []error{errors.New("API unavailable"), context.Canceled} {
		if _, err := GetDomainName(context.Background(), failingDomainReader{err: want}); !errors.Is(err, want) {
			t.Fatalf("got %v, want wrapped %v", err, want)
		}
	}
}

func TestGetPlatformDomainCacheAndRetry(t *testing.T) {
	ResetPlatformDomainForTest()
	original := newDomainReader
	t.Cleanup(func() { newDomainReader = original; ResetPlatformDomainForTest() })
	c := domainTestClient(t, "owner.example.org")
	calls := 0
	failInit := true
	newDomainReader = func() (client.Reader, error) {
		calls++
		if failInit {
			return nil, errors.New("config unavailable")
		}
		return c, nil
	}
	if got := GetPlatformDomain(context.Background()); got != "" {
		t.Fatalf("got %q on initialization failure", got)
	}
	failInit = false
	if got := GetPlatformDomain(context.Background()); got != "owner.example.org" {
		t.Fatalf("got %q after retry", got)
	}
	var terminus sysv1alpha1.Terminus
	if err := c.Get(context.Background(), client.ObjectKey{Name: "terminus"}, &terminus); err != nil {
		t.Fatal(err)
	}
	terminus.Spec.Settings["domainName"] = "new.example.cn"
	if err := c.Update(context.Background(), &terminus); err != nil {
		t.Fatal(err)
	}
	if got := GetPlatformDomain(context.Background()); got != "owner.example.org" || calls != 2 {
		t.Fatalf("cache got %q, calls %d", got, calls)
	}
	defaultCache.mu.Lock()
	defaultCache.expiresAt = time.Now().Add(-time.Second)
	defaultCache.mu.Unlock()
	if got := GetPlatformDomain(context.Background()); got != "new.example.cn" {
		t.Fatalf("got %q after expiry", got)
	}
	ResetPlatformDomainForTest()
	newDomainReader = func() (client.Reader, error) { return failingDomainReader{err: errors.New("API unavailable")}, nil }
	if got := GetPlatformDomain(context.Background()); got != "" {
		t.Fatalf("got %q on read failure", got)
	}
	newDomainReader = func() (client.Reader, error) { return c, nil }
	if got := GetPlatformDomain(context.Background()); got != "new.example.cn" {
		t.Fatalf("got %q after read retry", got)
	}
}

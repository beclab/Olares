package controllers

import (
	"context"
	"os"
	"testing"

	sysv1alpha1 "github.com/beclab/api/api/sys.bytetrade.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestInitializeSystemEnvDomainRegion(t *testing.T) {
	for _, tc := range []struct {
		domain, want string
		missing      bool
	}{
		{domain: "olares.com", want: "https://api.olares.com"},
		{domain: " Owner.Example.CN. ", want: "https://api.olares.cn"},
		{domain: "olares.local", want: "https://api.olares.com"},
		{domain: ""},
		{missing: true},
	} {
		t.Run(tc.domain, func(t *testing.T) {
			t.Setenv("OLARES_SYSTEM_REMOTE_SERVICE", "initial")
			t.Setenv("OLARES_USER_HUGGINGFACE_SERVICE", "initial")
			scheme := runtime.NewScheme()
			if err := sysv1alpha1.AddToScheme(scheme); err != nil {
				t.Fatal(err)
			}
			se := &sysv1alpha1.SystemEnv{ObjectMeta: metav1.ObjectMeta{Name: "remote-service"}, EnvVarSpec: sysv1alpha1.EnvVarSpec{EnvName: "OLARES_SYSTEM_REMOTE_SERVICE"}}
			ue := &sysv1alpha1.UserEnv{ObjectMeta: metav1.ObjectMeta{Name: "huggingface-service"}, EnvVarSpec: sysv1alpha1.EnvVarSpec{EnvName: "OLARES_USER_HUGGINGFACE_SERVICE"}}
			objects := []client.Object{se, ue}
			if !tc.missing {
				terminus := &sysv1alpha1.Terminus{ObjectMeta: metav1.ObjectMeta{Name: "terminus"}}
				terminus.Spec.Settings = map[string]string{"domainName": tc.domain}
				objects = append(objects, terminus)
			}
			c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()
			err := InitializeSystemEnvProcessEnv(context.Background(), c)
			if tc.want == "" {
				if err == nil {
					t.Fatal("expected domain read error")
				}
				if got := os.Getenv(se.EnvName); got != "initial" {
					t.Fatalf("failed initialization changed env to %q", got)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got := os.Getenv(se.EnvName); got != tc.want {
				t.Fatalf("env=%q, want %q", got, tc.want)
			}
			var updated sysv1alpha1.SystemEnv
			if err := c.Get(context.Background(), client.ObjectKeyFromObject(se), &updated); err != nil {
				t.Fatal(err)
			}
			var updatedUserEnv sysv1alpha1.UserEnv
			if err := c.Get(context.Background(), client.ObjectKeyFromObject(ue), &updatedUserEnv); err != nil {
				t.Fatal(err)
			}
			wantUserDefault := "https://huggingface.co/"
			if tc.want == "https://api.olares.cn" {
				wantUserDefault = "https://hf-mirror.com"
			}
			if updatedUserEnv.Default != wantUserDefault {
				t.Fatalf("user default=%q, want %q", updatedUserEnv.Default, wantUserDefault)
			}
			if updated.Default != tc.want {
				t.Fatalf("default=%q, want %q", updated.Default, tc.want)
			}
		})
	}
}

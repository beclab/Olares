package controllers

import (
	"reflect"
	"testing"

	sysv1alpha1 "github.com/beclab/api/api/sys.bytetrade.io/v1alpha1"
)

// TestAppEnvReferencesCoversSecrets guards the discovery step that makes live
// propagation work. The SystemEnv/UserEnv controllers use this to decide which
// AppEnvs to flag for sync when a variable is rotated. It originally scanned
// only Envs, which meant an app consuming a rotated credential as a secret was
// never flagged and silently kept the stale value until someone reinstalled.
func TestAppEnvReferencesCoversSecrets(t *testing.T) {
	envRef := &sysv1alpha1.AppEnv{
		Envs: []sysv1alpha1.AppEnvVar{{
			EnvVarSpec: sysv1alpha1.EnvVarSpec{EnvName: "APP_CDN"},
			ValueFrom:  &sysv1alpha1.ValueFrom{EnvName: "OLARES_SYSTEM_CDN_SERVICE"},
		}},
	}
	secretRef := &sysv1alpha1.AppEnv{
		Secrets: []sysv1alpha1.AppSecretVar{{
			Name:      "smtp-password",
			ValueFrom: &sysv1alpha1.ValueFrom{EnvName: "OLARES_USER_SMTP_PASSWORD"},
		}},
	}
	both := &sysv1alpha1.AppEnv{Envs: envRef.Envs, Secrets: secretRef.Secrets}

	cases := []struct {
		name    string
		appEnv  *sysv1alpha1.AppEnv
		envName string
		want    bool
	}{
		{"env reference matches", envRef, "OLARES_SYSTEM_CDN_SERVICE", true},
		{"env reference does not match", envRef, "OLARES_UNRELATED", false},
		{"secret reference matches", secretRef, "OLARES_USER_SMTP_PASSWORD", true},
		{"secret reference does not match", secretRef, "OLARES_UNRELATED", false},
		{"mixed, matches via env", both, "OLARES_SYSTEM_CDN_SERVICE", true},
		{"mixed, matches via secret", both, "OLARES_USER_SMTP_PASSWORD", true},
		{"empty appenv", &sysv1alpha1.AppEnv{}, "OLARES_ANYTHING", false},
		{
			"declaration without valueFrom is not a reference",
			&sysv1alpha1.AppEnv{Secrets: []sysv1alpha1.AppSecretVar{{Name: "orphan"}}},
			"OLARES_ANYTHING",
			false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := appEnvReferences(tc.appEnv, tc.envName); got != tc.want {
				t.Errorf("appEnvReferences(%q) = %v, want %v", tc.envName, got, tc.want)
			}
		})
	}
}

// TestAppSecretVarHasNoValueField is a structural guard on the privacy property
// the design depends on: the resolved value must live only in the Kubernetes
// Secret. If a value-bearing field is ever added to AppSecretVar, the AppEnv CR
// becomes a place where plaintext secrets can accumulate, which is exactly what
// declaring a variable as a secret is meant to avoid.
func TestAppSecretVarHasNoValueField(t *testing.T) {
	forbidden := map[string]struct{}{
		"Value": {}, "Default": {}, "Data": {}, "Secret": {},
	}

	typ := reflect.TypeOf(sysv1alpha1.AppSecretVar{})
	for i := 0; i < typ.NumField(); i++ {
		if _, bad := forbidden[typ.Field(i).Name]; bad {
			t.Errorf("AppSecretVar gained a value-bearing field %q; resolved secret "+
				"values must never be persisted on the AppEnv CR", typ.Field(i).Name)
		}
	}
}

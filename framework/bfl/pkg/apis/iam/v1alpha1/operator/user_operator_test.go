package operator

import (
	"testing"

	"bytetrade.io/web3os/bfl/pkg/constants"
	iamV1alpha2 "github.com/beclab/api/iam/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestLocalUserDomainWithoutOwner(t *testing.T) {
	for _, tt := range []struct {
		name        string
		annotations map[string]string
		zone        string
	}{
		{
			name: "cli-created admin without onboarding marker",
			annotations: map[string]string{
				constants.AnnotationUserCreator:   "cli",
				constants.UserAnnotationOwnerRole: constants.RoleAdmin,
			},
		},
		{
			name: "user without creator or domain",
		},
		{
			name: "cli-created admin with its own domain",
			annotations: map[string]string{
				constants.AnnotationUserCreator:   "cli",
				constants.UserAnnotationOwnerRole: constants.RoleAdmin,
				constants.UserAnnotationZoneKey:   "alice.example.com",
			},
			zone: "alice.example.com",
		},
		{
			name: "user with its own domain",
			annotations: map[string]string{
				constants.AnnotationUserCreator: "alice",
				constants.UserAnnotationZoneKey: "bob.example.com",
			},
			zone: "bob.example.com",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			// No Kubernetes client: these paths must not look up an owner or creator.
			op := &UserOperator{}
			user := &iamV1alpha2.User{ObjectMeta: metav1.ObjectMeta{
				Name: "alice", Annotations: tt.annotations,
			}}
			if zone := op.GetUserZone(user); zone != tt.zone {
				t.Fatalf("GetUserZone = %q, want %q", zone, tt.zone)
			}
			ephemeral, zone, err := op.GetUserDomainType(user)
			if err != nil || ephemeral || zone != tt.zone {
				t.Fatalf("GetUserDomainType = (%t, %q, %v), want (false, %q, nil)", ephemeral, zone, err, tt.zone)
			}
		})
	}
}

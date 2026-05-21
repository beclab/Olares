package routecontrol

import (
	"testing"

	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestReferenceGrantName(t *testing.T) {
	srr := &srrv1alpha1.SharedRouteRegistry{
		ObjectMeta: metav1.ObjectMeta{Namespace: "ollamav2-brucedai"},
	}
	got := referenceGrantName(srr)
	if got != "allow-httproute-ollamav2-brucedai" {
		t.Fatalf("referenceGrantName() = %q", got)
	}
}

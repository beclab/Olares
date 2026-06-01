package routecontrol

import "testing"

func TestDefaultInClusterStrongIdentityServicePort_value(t *testing.T) {
	const opaqueStart int32 = 19001
	const opaqueEnd int32 = 19003

	got := DefaultInClusterStrongIdentityServicePort
	if got != 8081 {
		t.Fatalf("DefaultInClusterStrongIdentityServicePort = %d, want %d", got, 8081)
	}
	if got == 0 {
		t.Fatalf("DefaultInClusterStrongIdentityServicePort must be non-zero")
	}
	forbidden := []int32{80, 443, 8080}
	for _, p := range forbidden {
		if got == p {
			t.Fatalf("DefaultInClusterStrongIdentityServicePort must not equal %d", p)
		}
	}
	if got >= opaqueStart && got <= opaqueEnd {
		t.Fatalf("DefaultInClusterStrongIdentityServicePort must not be in opaque range [%d,%d]", opaqueStart, opaqueEnd)
	}
}

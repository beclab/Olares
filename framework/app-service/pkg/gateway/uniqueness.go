package gateway

import (
	"context"
	"errors"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
)

// ErrHash8Collision is returned when a SharedEntrance's logical hostPattern is
// already owned by a different SRR somewhere in the cluster.
var ErrHash8Collision = errors.New("HASH8_COLLISION")

// CheckLogicalPatternUniqueness lists every SRR across the cluster and returns
// ErrHash8Collision if any registry other than (selfNS, selfName) already owns
// the same logical hostPattern.
func CheckLogicalPatternUniqueness(ctx context.Context, c client.Client,
	pattern, selfNS, selfName string) error {
	if pattern == "" {
		return fmt.Errorf("uniqueness: empty pattern")
	}
	list := &srrv1alpha1.SharedRouteRegistryList{}
	if err := c.List(ctx, list); err != nil {
		return fmt.Errorf("list SRRs: %w", err)
	}
	for i := range list.Items {
		o := &list.Items[i]
		if o.Namespace == selfNS && o.Name == selfName {
			continue
		}
		for _, h := range o.Spec.HostPatterns {
			if h == pattern {
				return fmt.Errorf("%w: pattern %q already owned by SRR %s/%s",
					ErrHash8Collision, pattern, o.Namespace, o.Name)
			}
		}
	}
	return nil
}

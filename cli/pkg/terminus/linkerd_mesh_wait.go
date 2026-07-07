package terminus

import (
	"context"
	"fmt"
	"strings"
	"time"

	agwconfig "github.com/beclab/Olares/framework/app-gateway/pkg/config"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	netv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	appGatewayMeshNPName         = "app-gateway-mesh-np"
	appGatewayMeshNPWaitTimeout  = 5 * time.Minute
	appGatewayMeshNPPollInterval = 5 * time.Second
)

var appGatewayMeshNPWaitNamespaces = []string{
	agwconfig.LinkerdNamespace(),
	agwconfig.Namespace(),
}

// WaitAppGatewayMeshNP waits for app-service to reconcile app-gateway-mesh-np in linkerd and os-gateway.
type WaitAppGatewayMeshNP struct {
	common.KubeAction
}

func (t *WaitAppGatewayMeshNP) Execute(_ connector.Runtime) error {
	config, err := ctrl.GetConfig()
	if err != nil {
		return err
	}
	k8sClient, err := client.New(config, client.Options{})
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), appGatewayMeshNPWaitTimeout)
	defer cancel()
	return waitAppGatewayMeshNP(ctx, k8sClient, appGatewayMeshNPWaitTimeout, appGatewayMeshNPPollInterval)
}

func waitAppGatewayMeshNP(ctx context.Context, c client.Client, timeout, pollInterval time.Duration) error {
	start := time.Now()
	for {
		missing, err := missingAppGatewayMeshNamespaces(ctx, c)
		if err != nil {
			return err
		}
		if len(missing) == 0 {
			return nil
		}
		if time.Since(start) >= timeout {
			return fmt.Errorf(
				"WaitAppGatewayMeshNP: timed out after %s waiting for NetworkPolicy %q reconciled by app-service; missing in namespace(s): %s",
				timeout, appGatewayMeshNPName, strings.Join(missing, ", "),
			)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(pollInterval):
		}
	}
}

func missingAppGatewayMeshNamespaces(ctx context.Context, c client.Client) ([]string, error) {
	var missing []string
	for _, ns := range appGatewayMeshNPWaitNamespaces {
		var np netv1.NetworkPolicy
		err := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: appGatewayMeshNPName}, &np)
		if err != nil {
			if apierrors.IsNotFound(err) {
				missing = append(missing, ns)
				continue
			}
			return nil, fmt.Errorf("WaitAppGatewayMeshNP: get app-gateway-mesh-np in namespace %s: %w", ns, err)
		}
	}
	return missing, nil
}

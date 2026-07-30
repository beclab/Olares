package terminus

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	sysv1alpha1 "github.com/beclab/Olares/framework/app-service/api/sys.bytetrade.io/v1alpha1"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	kruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/beclab/Olares/cli/pkg/common"
	cc "github.com/beclab/Olares/cli/pkg/core/common"
)

// JoinClusterScriptURL returns the CDN URL of the worker join bootstrap script
// for an Olares version.
//
// It follows the convention of every other versioned artifact on the CDN: the
// version lives in the file name, at the bucket root, rather than in a path
// segment. Unlike the release tarballs it carries no vendor repo path, because
// the script is vendor-agnostic -- it derives the vendor download path itself,
// from /etc/machine.info, before it fetches anything.
//
// The version is also baked into the published script, so a command built from
// this URL does not need to pass it through the environment as well.
func JoinClusterScriptURL(cdnService, olaresVersion string) string {
	base := strings.TrimRight(strings.TrimSpace(cdnService), "/")
	if base == "" {
		base = cc.DefaultOlaresCDNService
	}
	name := fmt.Sprintf("joincluster-v%s.sh", olaresVersion)
	scriptURL, err := url.JoinPath(base, name)
	if err != nil {
		// JoinPath only fails on an unparseable base, which CDN validation
		// rules out; keep a predictable value instead of propagating an error
		// into every message that embeds a URL.
		return base + "/" + name
	}
	return scriptURL
}

// ClusterCDNService returns the CDN endpoint this Olares cluster is configured
// to use, read from the OLARES_SYSTEM_CDN_SERVICE system environment variable.
//
// That variable is the authoritative, per-region setting (it is switched to the
// Chinese endpoint for .cn domains and is user-editable in Settings), and it
// lives only in the cluster, so it cannot be recovered from the local process
// environment. Callers that need to hand a CDN endpoint to another machine must
// resolve it from here rather than falling back to the compiled-in default.
//
// It returns an empty string, without an error, when the cluster does not
// declare the variable, so callers can fall back to their own configuration.
func ClusterCDNService(ctx context.Context) (string, error) {
	config, err := ctrl.GetConfig()
	if err != nil {
		return "", fmt.Errorf("failed to load kubeconfig: %w", err)
	}
	scheme := kruntime.NewScheme()
	if err := sysv1alpha1.AddToScheme(scheme); err != nil {
		return "", fmt.Errorf("failed to add systemenv scheme: %w", err)
	}
	c, err := ctrlclient.New(config, ctrlclient.Options{Scheme: scheme})
	if err != nil {
		return "", fmt.Errorf("failed to create kube client: %w", err)
	}
	resourceName, err := apputils.EnvNameToResourceName(common.ENV_OLARES_CDN_SERVICE)
	if err != nil {
		return "", fmt.Errorf("invalid system env name %s: %w", common.ENV_OLARES_CDN_SERVICE, err)
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var systemEnv sysv1alpha1.SystemEnv
	if err := c.Get(ctx, types.NamespacedName{Name: resourceName}, &systemEnv); err != nil {
		// A cluster from before the variable existed reports either a missing
		// object or a missing kind; neither is a failure for the caller.
		if apierrors.IsNotFound(err) || meta.IsNoMatchError(err) {
			return "", nil
		}
		return "", fmt.Errorf("failed to get system env %s: %w", resourceName, err)
	}
	return strings.TrimRight(strings.TrimSpace(systemEnv.GetEffectiveValue()), "/"), nil
}

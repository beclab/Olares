package v1

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"bytetrade.io/web3os/bfl/pkg/client/dynamic_client/users"
	"bytetrade.io/web3os/bfl/pkg/constants"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation"
)

var localUsername = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$`)

// Older local binding code persisted username + "@" when the domain was empty.
// Normalize only known local formats at the API boundary; keep cloud IDs intact.
func normalizeLocalOlaresID(olaresID string) string {
	name, domain, qualified := strings.Cut(olaresID, "@")
	if localUsername.MatchString(name) && (!qualified || domain == "" || domain == "olares.local") {
		return name
	}
	return olaresID
}

func isLocalOlaresID(olaresID string) bool {
	return localUsername.MatchString(normalizeLocalOlaresID(olaresID))
}

// The cluster's own name and zone select every user's service entrances.
// Without a zone, preserve the bare local name and existing LAN Vault URL.
func clusterConnectionURLs(localName, zone string, status constants.WizardStatus) (string, string, error) {
	if !localUsername.MatchString(localName) {
		return "", "", fmt.Errorf("machine local name is missing or invalid")
	}
	if zone == "" {
		vaultURL := ""
		if status == constants.WaitActivateVault {
			vaultURL = "http://vault." + localName + ".olares.local/server"
		}
		return localName, vaultURL, nil
	}
	zone = strings.TrimSuffix(zone, ".")
	clusterName := localName + "." + zone
	if zone == "" || len(validation.IsDNS1123Subdomain(clusterName)) != 0 {
		return "", "", fmt.Errorf("cluster zone is invalid")
	}
	scheme := "https"
	if strings.HasSuffix(clusterName, ".local") {
		scheme = "http"
	}
	vaultURL := ""
	if status == constants.WaitActivateVault {
		vaultURL = scheme + "://vault." + clusterName + "/server"
	}
	return clusterName, vaultURL, nil
}

func clusterConnectionInfo(ctx context.Context, olaresID string, status constants.WizardStatus) (string, string, error) {
	client, err := users.NewResourceTerminusClient()
	if err != nil {
		return "", "", err
	}
	terminus, err := client.Get(ctx, "terminus", metav1.GetOptions{})
	if err != nil {
		return "", "", err
	}
	// Older domain installations may predate the separate cluster local name.
	// Keep their existing info response working without inventing a host name.
	if terminus.Spec.LocalName == "" && !isLocalOlaresID(olaresID) {
		return "", "", nil
	}
	return clusterConnectionURLs(terminus.Spec.LocalName, terminus.Spec.Settings["domainName"], status)
}

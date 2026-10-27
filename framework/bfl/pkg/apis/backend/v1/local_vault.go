package v1

import (
	"context"
	"fmt"
	"regexp"

	"bytetrade.io/web3os/bfl/pkg/client/dynamic_client/users"
	"bytetrade.io/web3os/bfl/pkg/constants"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var localUsername = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$`)

// Local accounts keep using the machine LAN endpoint after a public domain is bound.
func clusterConnectionURLs(localName string, status constants.WizardStatus) (string, string, error) {
	if !localUsername.MatchString(localName) {
		return "", "", fmt.Errorf("machine local name is missing or invalid")
	}
	vaultURL := ""
	if status == constants.WaitActivateVault {
		vaultURL = "http://vault." + localName + ".olares.local/server"
	}
	return localName, vaultURL, nil
}

func clusterConnectionInfo(ctx context.Context, status constants.WizardStatus) (string, string, error) {
	client, err := users.NewResourceTerminusClient()
	if err != nil {
		return "", "", err
	}
	terminus, err := client.Get(ctx, "terminus", metav1.GetOptions{})
	if err != nil {
		return "", "", err
	}
	return clusterConnectionURLs(terminus.Spec.LocalName, status)
}

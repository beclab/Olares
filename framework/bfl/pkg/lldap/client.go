package lldap

import (
	"context"
	"fmt"

	"github.com/beclab/lldap-client/pkg/cache/memory"
	"github.com/beclab/lldap-client/pkg/client"
	"github.com/beclab/lldap-client/pkg/config"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
)

func getCredentialVal(secret *corev1.Secret, key string) (string, error) {
	if value, ok := secret.Data[key]; ok {
		return string(value), nil
	}
	return "", fmt.Errorf("can not find credentialval for key %s", key)
}

func New() (*client.Client, error) {
	cfg, err := ctrl.GetConfig()
	if err != nil {
		return nil, err
	}
	clientSet, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	secret, err := clientSet.CoreV1().Secrets("os-platform").Get(context.TODO(), "lldap-credentials", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	bindUsername, err := getCredentialVal(secret, "lldap-ldap-user-dn")
	if err != nil {
		return nil, err
	}
	bindPassword, err := getCredentialVal(secret, "lldap-ldap-user-pass")
	if err != nil {
		return nil, err
	}
	lldapClient, err := client.New(&config.Config{
		Host:       "http://lldap-service.os-platform:17170",
		Username:   bindUsername,
		Password:   bindPassword,
		TokenCache: memory.New(),
	})
	if err != nil {
		return nil, err
	}
	return lldapClient, nil
}

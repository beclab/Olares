package tapr

import (
	"context"
	"fmt"

	"bytetrade.io/web3os/app-service/pkg/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// CreateOrUpdateSecret creates or updates secret for middleware password.
func CreateOrUpdateSecret(config *rest.Config, appName, namespace string, middlewareType MiddlewareType) error {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}
	return createOrUpdateSecret(clientset, appName, namespace, middlewareType)
}

func createOrUpdateSecret(clientset kubernetes.Interface, appName, namespace string, middleware MiddlewareType) error {
	secretName := fmt.Sprintf("%s-%s-%s-password", appName, namespace, middleware)
	_, err := clientset.CoreV1().Secrets(namespace).Get(context.Background(), secretName, metav1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
	} else {
		return nil
	}

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
		Data: map[string][]byte{
			"password": []byte(utils.RandString()),
		},
		Type: v1.SecretTypeOpaque,
	}
	_, err = clientset.CoreV1().Secrets(namespace).Create(context.Background(), secret, metav1.CreateOptions{})
	if err != nil {
		if errors.IsAlreadyExists(err) {
			_, err = clientset.CoreV1().Secrets(namespace).Update(context.Background(), secret, metav1.UpdateOptions{})
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}

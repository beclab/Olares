package argo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

const (
	syncProviderKey           = "syncProvider"
	syncProviderConfigMapName = "sync-provider"
)

func jsonRawMarshal(t interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(t)
	return buffer.Bytes(), err
}

func saveProvidersToConfigMap(clientset *kubernetes.Clientset, namespace string, sliceData []map[string]interface{}) error {
	if len(sliceData) == 0 {
		return nil
	}

	data, err := convertMapSliceToMap(sliceData, syncProviderKey)
	if err != nil {
		return err
	}

	return storeConfigMap(clientset, namespace, syncProviderConfigMapName, data)
}

func convertMapSliceToMap(data []map[string]interface{}, key string) (map[string]string, error) {
	jsonData, err := jsonRawMarshal(data)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	result[key] = string(jsonData)

	return result, nil
}

func storeConfigMap(clientset *kubernetes.Clientset, namespace, name string, data map[string]string) error {
	configMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: data,
	}

	_, err := clientset.CoreV1().ConfigMaps(namespace).Create(context.Background(), configMap, metav1.CreateOptions{})
	if err != nil {
		if errors.IsAlreadyExists(err) {
			_, err = clientset.CoreV1().ConfigMaps(namespace).Update(context.Background(), configMap, metav1.UpdateOptions{})
			if err != nil {
				return err
			}
			return nil
		}
		return fmt.Errorf("failed to create ConfigMap: %v", err)
	}
	klog.Infof("Data=%v stored in ConfigMap=%s in namespace=%s", data, name, namespace)
	return nil
}

// GetProviderData get provider data from specified configmap.
func GetProviderData(clientset *kubernetes.Clientset, namespace string) ([]map[string]interface{}, error) {
	return readConfigMap(clientset, namespace, syncProviderConfigMapName)
}

func readConfigMap(clientset *kubernetes.Clientset, namespace, name string) ([]map[string]interface{}, error) {
	// Get the ConfigMap
	configMap, err := clientset.CoreV1().ConfigMaps(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get ConfigMap: %v", err)
	}

	// Convert string data to original format
	if data, ok := configMap.Data[syncProviderKey]; ok {
		klog.Infof("Data=%v get from ConfigMap=%s in namespace=%s", data, name, namespace)
		var resData []map[string]interface{}
		err = json.Unmarshal([]byte(data), &resData)
		if err != nil {
			return nil, fmt.Errorf("failed to Unmarshal: %v", err)
		}

		return resData, nil
	}

	return nil, nil
}

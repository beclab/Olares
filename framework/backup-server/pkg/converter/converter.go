package converter

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

var unstructuredConverter = runtime.DefaultUnstructuredConverter

func FromUnstructured(u map[string]interface{}, obj interface{}) error {
	return unstructuredConverter.FromUnstructured(u, obj)
}

func ToUnstructured(obj any) (map[string]interface{}, error) {
	return unstructuredConverter.ToUnstructured(obj)
}

func UnstructuredListTo(list *unstructured.UnstructuredList, obj any) error {
	return unstructuredConverter.FromUnstructured(list.UnstructuredContent(), obj)
}

func FromUnstructuredWithValidation(u map[string]interface{}, obj interface{}, returnUnknownFields bool) error {
	return unstructuredConverter.FromUnstructuredWithValidation(u, obj, returnUnknownFields)
}

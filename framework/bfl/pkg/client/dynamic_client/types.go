package dynamic_client

import (
	"k8s.io/apimachinery/pkg/runtime"
)

var UnstructuredConverter = runtime.DefaultUnstructuredConverter

var ToUnstructured = UnstructuredConverter.ToUnstructured

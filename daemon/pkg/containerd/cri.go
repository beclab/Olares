package containerd

import (
	cdefaults "github.com/containerd/containerd/defaults"
	"go.opentelemetry.io/otel/trace/noop"
	criapi "k8s.io/cri-api/pkg/apis"
	criclient "k8s.io/cri-client/pkg"
	"k8s.io/klog/v2"
	"time"
)

var (
	criDefaultTimeout = 5 * time.Second
)

func NewCRIImageService() (criapi.ImageManagerService, error) {
	tp := noop.NewTracerProvider()
	logger := klog.Background()
	return criclient.NewRemoteImageService(cdefaults.DefaultAddress, criDefaultTimeout, tp, &logger)
}

func NewCRIRuntimeService() (criapi.RuntimeService, error) {
	tp := noop.NewTracerProvider()
	logger := klog.Background()
	return criclient.NewRemoteRuntimeService(cdefaults.DefaultAddress, criDefaultTimeout, tp, &logger)
}

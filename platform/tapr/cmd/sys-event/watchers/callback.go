package watchers

import (
	"context"
	"errors"
	"fmt"
	"time"

	aprv1 "bytetrade.io/web3os/tapr/pkg/apis/apr/v1alpha1"
	aprclientset "bytetrade.io/web3os/tapr/pkg/generated/clientset/versioned"
	"github.com/emicklei/go-restful"
	"github.com/go-resty/resty/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
)

const InvokeRetry = 10

type CallbackInvoker struct {
	AprClient *aprclientset.Clientset
	Retriable func(error) bool
}

// invoke callback with 'data' when 'filter' is true
func (s *CallbackInvoker) Invoke(ctx context.Context, filter func(cb *aprv1.SysEventRegistry) bool, data interface{}) (err error) {
	callbacks, err := s.AprClient.AprV1alpha1().SysEventRegistries("").List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.Error("list sys event callbacks error, ", err)
		return err
	}

	backoff := wait.Backoff{
		Duration: time.Second,
		Factor:   2,
		Jitter:   0.1,
		Steps:    InvokeRetry,
		Cap:      120 * time.Second,
	}

	for _, cb := range callbacks.Items {
		if filter(&cb) {
			if cb.Spec.Callback == "" {
				klog.Error("callback url is empty, ", cb.Name, ", ", cb.Namespace)
				return errors.New("callback url is empty")
			}

			retriable := func(e error) bool {
				return s.Retriable(e)
			}

			if err = retry.OnError(backoff,
				retriable,
				func() error {
					klog.Info("send event ", cb.Spec.Event, " to, ", cb.Name, ", ", cb.Spec.Callback)
					return s.sendEvent(ctx, &cb, data)
				}); err != nil {
				return err
			}
		}
	}

	klog.Info("success to send events to all callbacks")
	return nil
}

func (s *CallbackInvoker) sendEvent(ctx context.Context, cb *aprv1.SysEventRegistry, data interface{}) error {
	client := resty.New().SetTimeout(2 * time.Minute)

	// TODO: invoke with auth if callback url is in the user space
	res, err := client.R().
		SetHeader(restful.HEADER_ContentType, restful.MIME_JSON).
		SetBody(data).
		Post(cb.Spec.Callback)

	if err != nil {
		klog.Error("invoke callback error, ", err, ", ", cb.Name, ", ", cb.Namespace)
		return err
	}

	if res.StatusCode() >= 400 {
		klog.Error("invoke callback response error code, ", res.StatusCode(), ", ", cb.Name, ", ", cb.Namespace)
		if res.StatusCode() == 493 {
			return fmt.Errorf("[%s] response forbidden, canceled", cb.Name)
		}
		return fmt.Errorf("invoke callback [%s] response error", cb.Name)
	}

	klog.Info("success to invoke callback, ", cb.Name, ", ", cb.Namespace, ", ", string(res.Body()))

	return nil
}

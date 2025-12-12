package serviceproxy

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	apiv1alpha1 "bytetrade.io/web3os/system-server/pkg/apiserver/v1alpha1/api"
	"bytetrade.io/web3os/system-server/pkg/constants"
	prodiverregistry "bytetrade.io/web3os/system-server/pkg/providerregistry/v1alpha1"
	"bytetrade.io/web3os/system-server/pkg/utils"

	"github.com/emicklei/go-restful/v3"
	"github.com/go-resty/resty/v2"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

type Dispatcher struct {
	registry  *prodiverregistry.Registry
	workqueue workqueue.RateLimitingInterface
	serverCtx context.Context
}

func NewDispatcher(ctx context.Context, registry *prodiverregistry.Registry) *Dispatcher {
	dispatcher := &Dispatcher{
		registry:  registry,
		workqueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "ContentWatcher"),
		serverCtx: ctx,
	}

	go dispatcher.runWorker()
	return dispatcher
}

func (d *Dispatcher) DoWatch(req *DispatchRequest) {
	d.workqueue.Add(req)
}

func (d *Dispatcher) runWorker() {
	for d.processNextWorkItem() {
		select {
		case <-d.serverCtx.Done():
			return
		default:
		}
	}
}

func (d *Dispatcher) processNextWorkItem() bool {
	obj, shutdown := d.workqueue.Get()

	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer d.workqueue.Done(obj)

		return d.dispatch(obj)
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

func (d *Dispatcher) dispatch(obj interface{}) error {
	request, ok := obj.(*DispatchRequest)
	if !ok {
		return fmt.Errorf("invalid dispatch obj, %s, %v", reflect.TypeOf(obj), obj)
	}

	klog.Info("dispatch request, ", utils.PrettyJSON(request))

	watchers, err := d.registry.GetWatchers(d.serverCtx,
		request.DataType,
		request.Group,
		request.Version,
	)

	if err != nil {
		return err
	}

	klog.Info("find watchers, ", len(watchers))

	for _, w := range watchers {
		for _, cb := range w.Spec.Callbacks {
			filtered, err := watchFilter(request.Data, cb.Filters)
			if err != nil {
				klog.Error("watcher filter error, ", err)
				continue
			}

			if cb.Op == request.Op && filtered {
				var url string
				if strings.HasPrefix(w.Spec.Endpoint, "http://") ||
					strings.HasPrefix(w.Spec.Endpoint, "https://") {
					url = fmt.Sprintf("%s%s", w.Spec.Endpoint, cb.URI)
				} else {
					url = fmt.Sprintf("http://%s%s", w.Spec.Endpoint, cb.URI)
				}

				klog.Info("watcher url: ", url)

				client := resty.New()

				resp, err := client.SetTimeout(2*time.Second).R().
					SetHeader(restful.HEADER_ContentType, restful.MIME_JSON).
					SetHeader(apiv1alpha1.BackendTokenHeader, constants.Nonce).
					SetBody(request).
					Post(url)

				if err != nil {
					return fmt.Errorf("invoke watcher err: %s", err.Error())
				}

				if resp.StatusCode() >= 400 {
					return fmt.Errorf("invoke watcher err: code %d, %s", resp.StatusCode(), string(resp.Body()))
				}

			}
		}
	}

	return nil
}

func watchFilter(data any, filter map[string][]string) (bool, error) {
	if filter == nil {
		return true, nil
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return false, err
	}

	var filterData map[string]interface{}
	err = json.Unmarshal(jsonData, &filterData)
	if err != nil {
		return false, err
	}

	for q, v := range filter {
		if f, ok := filterData[q]; ok {
			var field string

			switch f := f.(type) {
			// just app metadata can be filtered
			case string:
				field = f
			case float64:
				field = fmt.Sprintf("%f", f)
			}

			found := false
			for _, s := range v {
				if s == field {
					found = true
				}
			}

			if !found {
				return false, nil
			}
		}

	}

	return true, nil
}

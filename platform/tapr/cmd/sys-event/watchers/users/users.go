package users

import (
	"context"
	"errors"
	"time"

	"bytetrade.io/web3os/tapr/cmd/sys-event/watchers"
	aprv1 "bytetrade.io/web3os/tapr/pkg/apis/apr/v1alpha1"
	aprclientset "bytetrade.io/web3os/tapr/pkg/generated/clientset/versioned"
	"bytetrade.io/web3os/tapr/pkg/kubesphere"
	"github.com/emicklei/go-restful"
	"github.com/go-resty/resty/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
)

type watcher struct {
	dynamicClient dynamic.Interface
	ctx           context.Context
	aprClient     *aprclientset.Clientset
	cacheEvent    map[string]map[string]runtime.Object
	eventWatchers *watchers.Watchers
	subscriber    *Subscriber
	activingUsers map[string]string
}

const InvokeRetry = 10

const UserTerminusWizardStatus = "bytetrade.io/wizard-status"

type WizardStatus string

const (
	Completed WizardStatus = "completed"
)

var schemeGroupVersionResource = schema.GroupVersionResource{Group: "iam.kubesphere.io", Version: "v1alpha2", Resource: "users"}

func NewWatcher(ctx context.Context, kubeconfig *rest.Config,
	w *watchers.Watchers, n *watchers.Notification) *watcher {
	kubeClient := kubernetes.NewForConfigOrDie(kubeconfig)
	dynamicClient := dynamic.NewForConfigOrDie(kubeconfig)
	return &watcher{
		ctx:           ctx,
		dynamicClient: dynamicClient,
		aprClient:     aprclientset.NewForConfigOrDie(kubeconfig),
		cacheEvent:    make(map[string]map[string]runtime.Object),
		eventWatchers: w,
		subscriber:    &Subscriber{tasks: []task{&Notify{notification: n}, &UserDomain{kubeClient: kubeClient, dynamicClient: dynamicClient}}},
		activingUsers: make(map[string]string),
	}
}

func (w *watcher) Start() {
	go func() {
		for {
			if w.ctx.Err() != nil {
				klog.Info("user watcher canceled, ", w.ctx.Err())
				return
			}

			var (
				userWatcherOK     bool = false
				callbackWatcherOK bool = false
				event             watch.Event
				userWatcher       watch.Interface
				callbackWatcher   watch.Interface
				err               error
			)

			klog.Info("start to watch user event")
			if !userWatcherOK {
				userWatcher, err = w.dynamicClient.Resource(schemeGroupVersionResource).Watch(w.ctx, metav1.ListOptions{})
				if err != nil {
					klog.Error("watch user error, ", err)
					continue
				}
			}

			if !callbackWatcherOK {
				callbackWatcher, err = w.aprClient.AprV1alpha1().SysEventRegistries("").Watch(w.ctx, metav1.ListOptions{})
				if err != nil {
					klog.Error("watch callback error, ", err)
					continue
				}
			}

		eventSelect:
			for {
				select {
				case <-w.ctx.Done():
					klog.Info("stop user watcher")
					userWatcher.Stop()
					callbackWatcher.Stop()
					return

				case event, userWatcherOK = <-userWatcher.ResultChan():
					if !userWatcherOK {
						klog.Error("user watcher broken")
						break eventSelect
					}

					userData, ok := event.Object.(*unstructured.Unstructured)
					if !ok {
						klog.Error("unsupported user event object")
						continue
					}

					if w.isDupEvent(&event) {
						klog.Info("duplicate user event, ignore")
						continue
					}

					var user kubesphere.User
					err := runtime.DefaultUnstructuredConverter.FromUnstructured(userData.Object, &user)
					if err != nil {
						klog.Error("convert event object to user error, ", err)
						continue
					}

					switch event.Type {
					case watch.Added:
						if watchers.ValidWatchDuration(&user.ObjectMeta) {
							w.eventWatchers.Enqueue(watchers.EnqueueObj{
								Obj:       &user,
								Action:    watchers.ADD,
								Subscribe: w.subscriber,
							})
						}

						if err := w.invokeUserCreatedCB(nil, &user); err != nil {
							klog.Error("invoke user create callback error, ", err, user.Name)
							continue
						}

						klog.Info("invoke user create callback succeed, ", user.Name)

					case watch.Deleted:
						w.eventWatchers.Enqueue(watchers.EnqueueObj{
							Obj:       &user,
							Action:    watchers.DELETE,
							Subscribe: w.subscriber,
						})

						if err := w.invokeUserDeletedCB(nil, &user); err != nil {
							klog.Error("invoke user delete callback error, ", err, user.Name)
							continue
						}

						klog.Info("invoke user delete callback succeed, ", user.Name)

					case watch.Modified:
						w.eventWatchers.Enqueue(watchers.EnqueueObj{
							Obj:       &user,
							Action:    watchers.UPDATE,
							Subscribe: w.subscriber,
						})

						if status, ok := user.Annotations[UserTerminusWizardStatus]; ok {
							switch status {
							case string(Completed):
								if s, found := w.activingUsers[user.Name]; found {
									if s != string(Completed) {
										delete(w.activingUsers, user.Name)
										if err := w.invokeUserActivedCB(nil, &user); err != nil {
											klog.Error("invoke user active callback error, ", err, user.Name)
											continue
										}

										klog.Info("invoke user active callback succeed, ", user.Name)
									}
								}
							default:
								w.activingUsers[user.Name] = status
							}
						}

					default:
						klog.Info("ignore user event, ", event.Type)
					}

					w.eventRecord(&event)

				case event, callbackWatcherOK = <-callbackWatcher.ResultChan():
					if !callbackWatcherOK {
						klog.Error("callback watcher broken")
						break eventSelect
					}

					callback, ok := event.Object.(*aprv1.SysEventRegistry)
					if !ok {
						klog.Error("unsupported callback event object")
						continue
					}

					if w.isDupEvent(&event) {
						klog.Info("duplicate sys event, ignore")
						continue
					}

					if callback.Spec.Type != aprv1.Subscriber || callback.Spec.Event != aprv1.UserCreate {
						continue
					}

					switch event.Type {
					case watch.Added:
						klog.Info("a new callback is registered, ", callback.Name, ", ", callback.Namespace)

						userDataes, err := w.dynamicClient.Resource(schemeGroupVersionResource).List(w.ctx, metav1.ListOptions{})
						if err != nil {
							klog.Error("list user error, ", err)
							continue
						}

						// when a new sys event callback is registered, all users in os will send to it to initialize
						for _, userData := range userDataes.Items {
							var user kubesphere.User
							err := runtime.DefaultUnstructuredConverter.FromUnstructured(userData.Object, &user)
							if err != nil {
								klog.Error("convert user data to user error, ", err)
								continue
							}

							if err := w.invokeUserCreatedCB(&aprv1.SysEventRegistryList{Items: []aprv1.SysEventRegistry{*callback}}, &user); err != nil {
								klog.Error("invoke user create callback error, ", err, user.Name)
								continue
							}

						} // end of user loop

					default:
						klog.Info("ignore sys event, ", event.Type)
					} // end of switch

					w.eventRecord(&event)
				}
			}
		}
	}()
}

func (w *watcher) invokeCallbacks(callbacks *aprv1.SysEventRegistryList, trigger func(cb *aprv1.SysEventRegistry) bool, action func(cb *aprv1.SysEventRegistry) error) (err error) {
	if callbacks == nil {
		callbacks, err = w.aprClient.AprV1alpha1().SysEventRegistries("").List(w.ctx, metav1.ListOptions{})
		if err != nil {
			klog.Error("list sys event callbacks error, ", err)
			return err
		}
	}

	backoff := wait.Backoff{
		Duration: time.Second,
		Factor:   2,
		Jitter:   0.1,
		Steps:    InvokeRetry,
		Cap:      120 * time.Second,
	}

	for _, cb := range callbacks.Items {
		if trigger(&cb) {
			if cb.Spec.Callback == "" {
				klog.Error("callback url is empty, ", cb.Name, ", ", cb.Namespace)
				return errors.New("callback url is empty")
			}

			if err = retry.OnError(backoff, func(err error) bool { return true }, func() error {
				return action(&cb)
			}); err != nil {
				return err
			}
		}
	}

	klog.Info("success to send events to all callbacks")
	return nil
}

func (w *watcher) invokeUserCreatedCB(callbacks *aprv1.SysEventRegistryList, user *kubesphere.User) error {
	return w.invokeCallbacks(callbacks,
		func(cb *aprv1.SysEventRegistry) bool {
			return cb.Spec.Type == aprv1.Subscriber && cb.Spec.Event == aprv1.UserCreate
		},
		func(cb *aprv1.SysEventRegistry) error {
			klog.Info("start to send event to [", cb.Spec.Type, ", ", cb.Spec.Event, "], ", cb.Spec.Callback)

			postUserInfo := &struct {
				Name  string `json:"name"`
				Role  string `json:"role"`
				Email string `json:"email"`
			}{
				Name:  user.Name,
				Role:  user.Annotations["bytetrade.io/owner-role"],
				Email: user.Spec.Email,
			}
			client := resty.New().SetTimeout(2 * time.Second)
			res, err := client.R().
				SetHeader(restful.HEADER_ContentType, restful.MIME_JSON).
				SetBody(postUserInfo).
				Post(cb.Spec.Callback)

			if err != nil {
				klog.Error("invoke callback error, ", err, ", ", cb.Name, ", ", cb.Namespace)
				return err
			}

			if res.StatusCode() >= 400 {
				klog.Error("invoke callback response error code, ", res.StatusCode(), ", ", cb.Name, ", ", cb.Namespace)
				return errors.New("invoke callback response error")
			}

			klog.Info("success to invoke callback, ", cb.Name, ", ", cb.Namespace, ", ", string(res.Body()))

			return nil
		},
	)
}

func (w *watcher) invokeUserDeletedCB(callbacks *aprv1.SysEventRegistryList, user *kubesphere.User) error {
	return w.invokeCallbacks(callbacks,
		func(cb *aprv1.SysEventRegistry) bool {
			return cb.Spec.Type == aprv1.Subscriber && cb.Spec.Event == aprv1.UserDelete
		},
		func(cb *aprv1.SysEventRegistry) error {
			klog.Info("start to send event to [", cb.Spec.Type, ", ", cb.Spec.Event, "], ", cb.Spec.Callback)

			postUserInfo := &struct {
				Name  string `json:"name"`
				Email string `json:"email"`
			}{
				Name: user.Name,
				// Role:  user.Annotations["bytetrade.io/owner-role"],
				Email: user.Spec.Email,
			}
			client := resty.New().SetTimeout(2 * time.Second)
			res, err := client.R().
				SetHeader(restful.HEADER_ContentType, restful.MIME_JSON).
				SetBody(postUserInfo).
				Post(cb.Spec.Callback)

			if err != nil {
				klog.Error("invoke callback error, ", err, ", ", cb.Name, ", ", cb.Namespace)
				return err
			}

			if res.StatusCode() >= 400 {
				klog.Error("invoke callback response error code, ", res.StatusCode(), ", ", cb.Name, ", ", cb.Namespace)
				return errors.New("invoke callback response error")
			}

			klog.Info("success to invoke callback, ", cb.Name, ", ", cb.Namespace, ", ", string(res.Body()))

			return nil
		},
	)
}

func (w *watcher) invokeUserActivedCB(callbacks *aprv1.SysEventRegistryList, user *kubesphere.User) error {
	return w.invokeCallbacks(callbacks,
		func(cb *aprv1.SysEventRegistry) bool {
			return cb.Spec.Type == aprv1.Subscriber && cb.Spec.Event == aprv1.UserActive
		},
		func(cb *aprv1.SysEventRegistry) error {
			klog.Info("start to send event to [", cb.Spec.Type, ", ", cb.Spec.Event, "], ", cb.Spec.Callback)

			postUserInfo := &struct {
				Name  string `json:"name"`
				Email string `json:"email"`
			}{
				Name: user.Name,
				// Role:  user.Annotations["bytetrade.io/owner-role"],
				Email: user.Spec.Email,
			}
			client := resty.New().SetTimeout(2 * time.Second)
			res, err := client.R().
				SetHeader(restful.HEADER_ContentType, restful.MIME_JSON).
				SetBody(postUserInfo).
				Post(cb.Spec.Callback)

			if err != nil {
				klog.Error("invoke callback error, ", err, ", ", cb.Name, ", ", cb.Namespace)
				return err
			}

			if res.StatusCode() >= 400 {
				klog.Error("invoke callback response error code, ", res.StatusCode(), ", ", cb.Name, ", ", cb.Namespace)
				return errors.New("invoke callback response error")
			}

			klog.Info("success to invoke callback, ", cb.Name, ", ", cb.Namespace, ", ", string(res.Body()))

			return nil
		},
	)
}

func (w *watcher) eventRecord(e *watch.Event) {
	uid := getUid(e)

	if _, ok := w.cacheEvent[uid]; !ok {
		w.cacheEvent[uid] = make(map[string]runtime.Object)
	}

	w.cacheEvent[uid][string(e.Type)] = e.Object
}

func (w *watcher) isDupEvent(e *watch.Event) bool {
	uid := getUid(e)
	if oldEvent, ok := w.cacheEvent[uid]; ok {
		if oldMeta, ok := oldEvent[string(e.Type)]; ok {
			return getResourceVersion(oldMeta) == getResourceVersion(e.Object)
		}

		return false
	}

	return false
}

func getUid(e *watch.Event) string {
	var uid string
	switch meta := e.Object.(type) {
	case *unstructured.Unstructured:
		uid = string(meta.GetUID())
	case *aprv1.SysEventRegistry:
		uid = string(meta.UID)
	}

	return uid
}

func getResourceVersion(o runtime.Object) string {
	var version string
	switch meta := o.(type) {
	case *unstructured.Unstructured:
		version = meta.GetResourceVersion()
	case *aprv1.SysEventRegistry:
		version = meta.ResourceVersion
	}

	return version
}

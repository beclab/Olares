package apiserver

import (
	"fmt"
	"sync"
	"time"

	"bytetrade.io/web3os/app-service/pkg/apiserver/api"
	"bytetrade.io/web3os/app-service/pkg/client/clientset"
	"bytetrade.io/web3os/app-service/pkg/constants"
	"bytetrade.io/web3os/app-service/pkg/utils"

	"github.com/emicklei/go-restful/v3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
)

var running bool = false
var switchLock sync.Mutex

func (h *Handler) disableGpuManagedMemory(req *restful.Request, resp *restful.Response) {
	if err := h.nvshareSwitch(req, false); err != nil {
		api.HandleError(resp, req, &errors.StatusError{
			ErrStatus: metav1.Status{Code: 400, Message: "operation failed, " + err.Error()},
		})

		return
	}
	resp.WriteAsJson(map[string]int{"code": 0})
}

func (h *Handler) enableGpuManagedMemory(req *restful.Request, resp *restful.Response) {
	if err := h.nvshareSwitch(req, true); err != nil {
		api.HandleError(resp, req, &errors.StatusError{
			ErrStatus: metav1.Status{Code: 400, Message: "operation failed, " + err.Error()},
		})

		return
	}
	resp.WriteAsJson(map[string]int{"code": 0})
}

func (h *Handler) nvshareSwitch(req *restful.Request, enable bool) error {
	client := req.Attribute(constants.KubeSphereClientAttribute).(*clientset.ClientSet)
	switchLock.Lock()
	defer switchLock.Unlock()

	if running {
		return fmt.Errorf("last operation is still running")
	}

	deployments, err := client.KubeClient.Kubernetes().AppsV1().Deployments("").List(req.Request.Context(), metav1.ListOptions{})
	if err != nil {
		klog.Error("list deployment error, ", err)
		return err
	}

	envValue := "0"
	if enable {
		envValue = "1"
	}

	for _, d := range deployments.Items {
		shouldUpdate := false
		for i, c := range d.Spec.Template.Spec.Containers {
			found := false
			for k := range c.Resources.Limits {
				if k == constants.NvshareGPU {
					found = true
					break
				}
			}

			if found {
				// a gpu request container
				addEnv := true
				for n, env := range d.Spec.Template.Spec.Containers[i].Env {
					if env.Name == constants.EnvNvshareManagedMemory {
						addEnv = false
						d.Spec.Template.Spec.Containers[i].Env[n].Value = envValue
						break
					}
				}

				if addEnv {
					d.Spec.Template.Spec.Containers[i].Env =
						append(d.Spec.Template.Spec.Containers[i].Env,
							corev1.EnvVar{Name: constants.EnvNvshareManagedMemory, Value: envValue})
				}

				shouldUpdate = true
			} // end found
		} // end of container loop

		if shouldUpdate {
			err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
				deployment, err := client.KubeClient.Kubernetes().AppsV1().Deployments(d.Namespace).
					Get(req.Request.Context(), d.Name, metav1.GetOptions{})

				if err != nil {
					return err
				}

				deployment.Spec.Template.Spec.Containers = d.Spec.Template.Spec.Containers

				_, err = client.KubeClient.Kubernetes().AppsV1().Deployments(d.Namespace).
					Update(req.Request.Context(), deployment, metav1.UpdateOptions{})

				return err
			})

			if err != nil {
				klog.Error("update deployment error, ", err, ", ", d.Name, ", ", d.Namespace)
				return err
			}
		} // should update
	} // end of deployment loop

	// update terminus
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		terminus, err := utils.GetTerminus(req.Request.Context(), h.ctrlClient)
		if err != nil {
			return err
		}

		terminus.Spec.Settings[constants.EnvNvshareManagedMemory] = envValue

		return h.ctrlClient.Update(req.Request.Context(), terminus)
	})

	if err != nil {
		klog.Error("update terminus error, ", err)

		return err
	}

	running = true
	// delay 30s, assume the all pods will be reload in 30s.
	delay := time.NewTimer(30 * time.Second)
	go func() {
		<-delay.C
		switchLock.Lock()
		defer switchLock.Unlock()

		running = false
	}()

	return nil
}

func (h *Handler) getManagedMemoryValue(req *restful.Request, resp *restful.Response) {
	terminus, err := utils.GetTerminus(req.Request.Context(), h.ctrlClient)
	if err != nil {
		klog.Error("get terminus value error, ", err)
		api.HandleError(resp, req, &errors.StatusError{
			ErrStatus: metav1.Status{Code: 400, Message: "get value error, " + err.Error()},
		})

		return
	}

	managed := true
	if v, ok := terminus.Spec.Settings[constants.EnvNvshareManagedMemory]; ok && v == "0" {
		managed = false
	}

	resp.WriteAsJson(&map[string]interface{}{
		"managed_memory": managed,
	},
	)
}

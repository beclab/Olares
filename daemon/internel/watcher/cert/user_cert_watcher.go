package cert

import (
	"context"
	"fmt"
	"time"

	"github.com/beclab/Olares/daemon/internel/watcher"
	"github.com/beclab/Olares/daemon/pkg/cluster/state"
	"github.com/beclab/Olares/daemon/pkg/utils"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	kubeErr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
)

var _ watcher.Watcher = &userCertWatcher{}

type userCertWatcher struct {
}

func NewCertWatcher() *userCertWatcher {
	return &userCertWatcher{}
}

// Watch implements watcher.Watcher.
func (u *userCertWatcher) Watch(ctx context.Context) {

	if state.CurrentState.TerminusState != state.TerminusRunning {
		return
	}

	kubeClient, err := utils.GetKubeClient()
	if err != nil {
		klog.Error("failed to get kube client, ", err)
		return
	}

	dynamicClient, err := utils.GetDynamicClient()
	if err != nil {
		klog.Error("failed to get dynamic client, ", err)
		return
	}

	users, err := utils.ListUsers(ctx, dynamicClient)
	if err != nil {
		klog.Error("failed to list users, ", err)
		return
	}

	for _, user := range users {
		namespace := fmt.Sprintf("user-space-%s", user.GetName())
		config, err := kubeClient.CoreV1().ConfigMaps(namespace).Get(ctx, "zone-ssl-config", metav1.GetOptions{})
		if err != nil {
			klog.Error("failed to get user config map, ", err, ", namespace: ", namespace)
			continue
		}

		if expired, ok := config.Data["expired_at"]; ok {
			expiredTime, err := time.Parse("2006-01-02T15:04:05Z", expired)
			if err != nil {
				klog.Error("failed to parse expired_at, ", err)
				continue
			}

			// Check if the certificate will expire within 10 days
			if expiredTime.Before(time.Now().Add(10 * 24 * time.Hour)) {
				klog.Info("user cert expired, ", user.GetName())
				err = createOrUpdateJob(ctx, kubeClient, namespace)
				if err != nil {
					klog.Error("failed to create or update job for user cert, ", err, ", namespace: ", namespace)
				} else {
					klog.Info("job created for user cert download, ", user.GetName(), ", namespace: ", namespace)
				}
			}
		}
	}
}

func createOrUpdateJob(ctx context.Context, kubeClient kubernetes.Interface, namespace string) error {
	currentJob, err := kubeClient.BatchV1().Jobs(namespace).Get(ctx, jobDownloadUserCert.Name, metav1.GetOptions{})
	if err != nil {
		if kubeErr.IsNotFound(err) {
			// Create the job if it does not exist
		} else {
			return fmt.Errorf("failed to get job: %w", err)
		}
	} else {
		// check the existing job
		if currentJob.Status.Succeeded == 0 || currentJob.Status.Failed > 0 {
			klog.Info("job is still running, skip creating a new one")
			return nil
		}

		// If the job exists and has completed, delete it before creating a new one
		klog.Info("delete existing job: ", currentJob.Name)
		err = kubeClient.BatchV1().Jobs(namespace).Delete(ctx, currentJob.Name, metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("failed to delete job: %w", err)
		}
	}

	job := jobDownloadUserCert.DeepCopy()
	job.Namespace = namespace
	_, err = kubeClient.BatchV1().Jobs(job.Namespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create job: %w", err)
	}

	klog.Info("Job created: ", job.Name)
	return nil
}

var jobDownloadUserCert = batchv1.Job{
	ObjectMeta: metav1.ObjectMeta{
		Name: "download-user-cert",
	},
	Spec: batchv1.JobSpec{
		BackoffLimit: ptr.To[int32](5),
		Template: corev1.PodTemplateSpec{
			Spec: corev1.PodSpec{
				RestartPolicy: corev1.RestartPolicyOnFailure,
				Containers: []corev1.Container{
					{
						Name:  "download-user-cert",
						Image: "busybox:1.28",
						Command: []string{"wget",
							"--header",
							"X-FROM-CRONJOB: true",
							"-qSO -",
							"http://bfl.user-space-pengpeng9/bfl/backend/v1/re-download-cert",
						},
					},
				},
			},
		},
	},
}

package pgclusterrestore

import (
	"context"
	"errors"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"

	"bytetrade.io/web3os/tapr/pkg/apis/apr/v1alpha1"
	"bytetrade.io/web3os/tapr/pkg/workload/citus"
	"k8s.io/klog/v2"
)

func (c *controller) handler(action Action, obj interface{}) error {
	restore, ok := obj.(*v1alpha1.PGClusterRestore)
	if !ok {
		return errors.New("invalid object, not pg cluster restore")
	}

	switch action {
	case ADD:

		switch restore.Status.State {
		case v1alpha1.RestoreStateError, v1alpha1.RestoreStateRejected, v1alpha1.RestoreStateReady:
			klog.Info("ignore restore with finished state")
			return nil
		}

		if restore.Spec.ClusterName == "" {
			err := errors.New("pg restore cluster name is empty")
			c.updateStatus(restore.DeepCopy(), v1alpha1.RestoreStateError, err)

			return err
		}

		klog.Info("find cluster to restore, ", restore.Spec.ClusterName, ", ", restore.Namespace)
		cluster, err := c.aprClientSet.AprV1alpha1().PGClusters(restore.Namespace).Get(c.ctx, restore.Spec.ClusterName, metav1.GetOptions{})
		if err != nil {
			klog.Error("find cluster to restore error, ", err, ", ", restore.Spec.ClusterName, ", ", restore.Namespace)
			return err
		}

		if cluster.Spec.AdminUser == "" {
			// get default user
			secret, err := c.k8sClientSet.CoreV1().Secrets(cluster.Namespace).Get(c.ctx, citus.CitusAdminSecretName, metav1.GetOptions{})
			if err != nil {
				return err
			}

			cluster.Spec.Password = v1alpha1.PasswordVar{
				Value: string(secret.Data["password"]),
			}
			cluster.Spec.AdminUser = string(secret.Data["user"])
		}

		admin := cluster.Spec.AdminUser
		password, err := cluster.Spec.Password.GetVarValue(c.ctx, c.k8sClientSet, restore.Namespace)
		if err != nil {
			klog.Error("find cluster admin user error, ", err, ", ", restore.Spec.ClusterName, ", ", restore.Namespace)
			return err
		}

		klog.Info("find cluster's backup file to restore, ", restore.Spec.BackupName, ", ", restore.Namespace)
		backup, err := c.aprClientSet.AprV1alpha1().PGClusterBackups(restore.Namespace).Get(c.ctx, restore.Spec.BackupName, metav1.GetOptions{})
		if err != nil {
			klog.Error("find cluster's backup error, ", err, ", ", restore.Spec.BackupName, ", ", restore.Namespace)
			return err
		}

		if backup.Status.State != v1alpha1.BackupStateReady {
			klog.Error("cluster's backup is not ready for restore, ", backup.Status.State, ", ", restore.Spec.BackupName, ", ", restore.Namespace)

			return errors.New("cluster's backup is not ready for restore")
		}

		if backup.Status.BackupPath == "" {
			err = errors.New("cluster's backup file is empty")
			c.updateStatus(restore, v1alpha1.RestoreStateRejected, err)
			klog.Error(err)
			return err
		}

		job := citus.RestoreJob.DeepCopy()
		job.Namespace = restore.Namespace
		currentJob, err := c.k8sClientSet.BatchV1().Jobs(job.Namespace).Get(c.ctx, job.Name, metav1.GetOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			klog.Error("get prev job error, ", err, ", ", job.Name, ", ", job.Namespace)
			return err
		}

		if err == nil {
			switch {
			case currentJob.Status.Active > 0:
				klog.Warning("there is a running restore job, do not do it repeat")
				_, err = c.updateStatus(restore, v1alpha1.RestoreStateRejected, errors.New("duplicate restore request"))
				return err
			case currentJob.Status.Succeeded > 0 || currentJob.Status.Failed > 0:
				klog.Info("remove prev finished restore job")
				err = c.k8sClientSet.BatchV1().Jobs(currentJob.Namespace).Delete(c.ctx, currentJob.Name, metav1.DeleteOptions{})
				if err != nil {
					klog.Info("delete prev restore job error, ", err, ", ", job.Name, ", ", job.Namespace)
					return err
				}
			} // end  of switch job status
		}

		klog.Info("create a new restore job, ", job.Name, ", ", job.Namespace)
		volumeMountPath := "/restore"
		jobEnv := []corev1.EnvVar{
			{
				Name:  "PG_HOST",
				Value: "citus-0.citus-headless",
			},
			{
				Name:  "PG_PORT",
				Value: "5432",
			},
			{
				Name:  "PGUSER",
				Value: admin,
			},
			{
				Name:  "PGPASSWORD",
				Value: password,
			},
			{
				Name:  "BACKUP_FILENAME",
				Value: volumeMountPath + backup.Status.BackupPath,
			},
		}

		if backup.Spec.VolumeSpec != nil {
			klog.Info("bind volume for job from backup cr", job.Name, ", ", job.Namespace)
			volumeName := "backup-data"

			job.Spec.Template.Spec.Volumes = []corev1.Volume{
				{
					Name:         volumeName,
					VolumeSource: backup.Spec.VolumeSpec.VolumeSource,
				},
			}

			job.Spec.Template.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{
				{
					Name:      volumeName,
					MountPath: volumeMountPath,
				},
			}

		} // end of if volume spec

		job.Spec.Template.Spec.Containers[0].Env = jobEnv
		_, err = c.k8sClientSet.BatchV1().Jobs(job.Namespace).Create(c.ctx, job, metav1.CreateOptions{})
		if err != nil {
			klog.Error("create restore job error, ", err, ", ", job.Name, ", ", job.Namespace)
			c.updateStatus(restore, v1alpha1.RestoreStateError, err)
			return err
		}

		restore, err = c.updateStatus(restore, v1alpha1.RestoreStateRunning, nil)
		if err != nil {
			klog.Error("update restore job running error")
			return err
		}

		klog.Info("waiting for restore job completed")
		err = c.waitForJobComplete(restore, job)
		if err != nil {
			return err
		}

		c.watchJobDelete(job)

	} // end of switch action
	return nil
}

func (c *controller) updateStatus(restore *v1alpha1.PGClusterRestore, state v1alpha1.RestoreState, errMsg error) (*v1alpha1.PGClusterRestore, error) {
	restore.Status.State = state
	if errMsg != nil {
		restore.Status.Error = errMsg.Error()
	}

	now := metav1.Now()
	switch state {
	case v1alpha1.RestoreStateError, v1alpha1.RestoreStateRejected, v1alpha1.RestoreStateReady:
		restore.Status.CompletedAt = &now
	case v1alpha1.RestoreStateRunning:
		restore.Status.StartAt = &now
	}

	r, err := c.aprClientSet.AprV1alpha1().PGClusterRestores(restore.Namespace).UpdateStatus(c.ctx, restore, metav1.UpdateOptions{})
	if err != nil {
		klog.Error("update restore status error, ", err, ", ", restore.Name, ", ", restore.Namespace)
		return nil, err
	}

	return r, nil
}

func (c *controller) waitForJobComplete(restore *v1alpha1.PGClusterRestore, job *batchv1.Job) error {
	return wait.PollWithContext(c.ctx, 5*time.Second, time.Hour,
		func(ctx context.Context) (done bool, err error) {
			j, err := c.k8sClientSet.BatchV1().Jobs(job.Namespace).Get(ctx, job.Name, metav1.GetOptions{})
			if err != nil {
				klog.Error("get restore job error, ", err, ", ", restore.Name, ", ", restore.Namespace)
				return false, err
			}

			switch {
			case j.Status.Active > 0:
				return false, nil
			case j.Status.Active == 0 && j.Status.Succeeded > 0:
				_, err := c.updateStatus(restore, v1alpha1.RestoreStateReady, nil)
				if err != nil {
					klog.Error("update restore job status error, ", err, ", ", restore.Name, ", ", restore.Namespace)
					return false, err
				}
				return true, nil
			case j.Status.Active == 0 && j.Status.Failed > 0:
				_, err := c.updateStatus(restore, v1alpha1.RestoreStateError, errors.New("restore job failed"))
				if err != nil {
					klog.Error("update restore job status error, ", err, ", ", restore.Name, ", ", restore.Namespace)
					return false, err
				}
				return true, nil
			}

			return false, nil
		},
	)
}

func (c *controller) watchJobDelete(job *batchv1.Job) {
	go func() {
		for {
			watcher, err := c.k8sClientSet.BatchV1().Jobs(job.Namespace).Watch(c.ctx, metav1.SingleObject(job.ObjectMeta))
			if err != nil {
				klog.Error("watch job error, ", err, ", ", job.Name, ", ", job.Namespace)
				return
			}

			select {
			case <-c.ctx.Done():
				watcher.Stop()
				return
			case event, ok := <-watcher.ResultChan():
				if !ok {
					time.Sleep(5 * time.Second)
					continue
				}

				switch event.Type {
				case watch.Deleted:

					watchedJob, ok := event.Object.(*batchv1.Job)
					if !ok {
						klog.Error("unexpected object")
						time.Sleep(5 * time.Second)
						continue
					}

					if watchedJob.Name != job.Name || watchedJob.Namespace != job.Namespace {
						klog.Error("watch object error, ", watchedJob.Name, ", ", watchedJob.Namespace)
						time.Sleep(5 * time.Second)
						continue
					}

					// job deleted, clean pod
					pods, err := c.k8sClientSet.CoreV1().Pods(job.Namespace).List(c.ctx, metav1.ListOptions{})
					if err != nil {
						klog.Error("list job pods error, ", err)
						return
					}

					for _, pod := range pods.Items {
						for _, o := range pod.ObjectMeta.OwnerReferences {
							if o.UID == job.UID {
								err = c.k8sClientSet.CoreV1().Pods(pod.Namespace).Delete(c.ctx, pod.Name, metav1.DeleteOptions{})
								if err != nil {
									klog.Warning("delete pod error, ", err, ", ", pod.Name, ", ", pod.Namespace)
								}
							}
						}
					}

					return
				default:
					continue
				}
			}
		}
	}()
}

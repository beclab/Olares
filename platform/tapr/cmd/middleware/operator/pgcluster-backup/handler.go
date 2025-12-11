package pgclusterbackup

import (
	"context"
	"errors"
	"os"
	"time"

	"bytetrade.io/web3os/tapr/pkg/apis/apr/v1alpha1"
	"bytetrade.io/web3os/tapr/pkg/workload/citus"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/klog/v2"
)

func (c *controller) handler(action Action, obj interface{}) error {
	backup, ok := obj.(*v1alpha1.PGClusterBackup)
	if !ok {
		return errors.New("invalid object, not pg cluster backup")
	}

	switch action {
	case ADD:
		switch backup.Status.State {
		case v1alpha1.BackupStateError, v1alpha1.BackupStateRejected, v1alpha1.BackupStateReady:
			klog.Info("ignore backup with finished state")
			return nil
		}
		if backup.Spec.ClusterName == "" {
			err := errors.New("pg backup cluster name is empty")
			c.updateStatus(backup, v1alpha1.BackupStateError, err)

			return err
		}

		klog.Info("find cluster to backup, ", backup.Spec.ClusterName, ", ", backup.Namespace)
		cluster, err := c.aprClientSet.AprV1alpha1().PGClusters(backup.Namespace).Get(c.ctx, backup.Spec.ClusterName, metav1.GetOptions{})
		if err != nil {
			klog.Error("find cluster to backup error, ", err, ", ", backup.Spec.ClusterName, ", ", backup.Namespace)
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
		password, err := cluster.Spec.Password.GetVarValue(c.ctx, c.k8sClientSet, backup.Namespace)
		if err != nil {
			klog.Error("find cluster admin user error, ", err, ", ", backup.Spec.ClusterName, ", ", backup.Namespace)
			return err
		}

		job := citus.BackupJob.DeepCopy()
		job.Namespace = backup.Namespace
		currentJob, err := c.k8sClientSet.BatchV1().Jobs(job.Namespace).Get(c.ctx, job.Name, metav1.GetOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			klog.Error("get prev job error, ", err, ", ", job.Name, ", ", job.Namespace)
			return err
		}

		if err == nil {
			switch {
			case currentJob.Status.Active > 0:
				klog.Warning("there is a running backup job, do not do it repeat")
				_, err = c.updateStatus(backup, v1alpha1.BackupStateRejected, errors.New("duplicate backup request"))
				return err
			case currentJob.Status.Succeeded > 0 || currentJob.Status.Failed > 0:
				klog.Info("remove prev finished backup job")
				err = c.k8sClientSet.BatchV1().Jobs(currentJob.Namespace).Delete(c.ctx, currentJob.Name, metav1.DeleteOptions{})
				if err != nil {
					klog.Info("delete prev backup job error, ", err, ", ", job.Name, ", ", job.Namespace)
					return err
				}

				// waiting for job deleted
				err = wait.PollWithContext(c.ctx, time.Second, time.Minute, func(ctx context.Context) (done bool, err error) {
					_, err = c.k8sClientSet.BatchV1().Jobs(job.Namespace).Get(ctx, currentJob.Name, metav1.GetOptions{})
					if err != nil {
						if apierrors.IsNotFound(err) {
							klog.Info("prev finished pg backup job deleted")
							return true, nil
						}

						return false, err
					}

					return false, nil
				})

				if err != nil {
					return err
				}

			} // end  of switch job status
		}

		klog.Info("create a new backup job, ", job.Name, ", ", job.Namespace)
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
		}

		if backup.Spec.VolumeSpec != nil {
			klog.Info("bind volume for job from backup cr", job.Name, ", ", job.Namespace)
			volumeName := "backup-data"
			volumeMountSubPath := "backup/" + getBackupSubPath()
			volumeMountPath := "/" + volumeMountSubPath
			backupFilePath := volumeMountPath + "/all.sql"

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
					SubPath:   volumeMountSubPath,
				},
			}

			jobEnv = append(jobEnv, corev1.EnvVar{
				Name:  "BACKUP_FILENAME",
				Value: backupFilePath,
			})

			backup.Status.BackupPath = backupFilePath
		} // end of if volume spec

		job.Spec.Template.Spec.Containers[0].Env = jobEnv

		_, err = c.k8sClientSet.BatchV1().Jobs(job.Namespace).Create(c.ctx, job, metav1.CreateOptions{})
		if err != nil {
			klog.Error("create backup job error, ", err, ", ", job.Name, ", ", job.Namespace)
			c.updateStatus(backup, v1alpha1.BackupStateError, err)
			return err
		}

		backup, err = c.updateStatus(backup, v1alpha1.BackupStateRunning, nil)
		if err != nil {
			klog.Errorf("update backup job running error %v", err)
			return err
		}

		klog.Info("waiting for pg backup job completed")
		err = c.waitForJobComplete(backup, job)
		if err != nil {
			klog.Errorf("waiting for pg backup job error %v", err)
			return err
		}

		c.watchJobDelete(job)

	case DELETE:
		// clean up the backup files
		if backup.Status.BackupPath != "" {
			return os.RemoveAll(backup.Status.BackupPath)
		}
	} // end of switch action

	return nil
}

func (c *controller) updateStatus(backup *v1alpha1.PGClusterBackup, state v1alpha1.BackupState, errMsg error) (*v1alpha1.PGClusterBackup, error) {
	backup.Status.State = state
	if errMsg != nil {
		backup.Status.Error = errMsg.Error()
	}

	now := metav1.Now()
	switch state {
	case v1alpha1.BackupStateError, v1alpha1.BackupStateRejected, v1alpha1.BackupStateReady:
		backup.Status.CompletedAt = &now
	case v1alpha1.BackupStateRunning:
		backup.Status.StartAt = &now
	}

	b, err := c.aprClientSet.AprV1alpha1().PGClusterBackups(backup.Namespace).UpdateStatus(c.ctx, backup, metav1.UpdateOptions{})
	if err != nil {
		klog.Error("update backup status error, ", err, ", ", backup.Name, ", ", backup.Namespace)
		return nil, err
	}

	return b, nil
}

func (c *controller) waitForJobComplete(backup *v1alpha1.PGClusterBackup, job *batchv1.Job) error {
	return wait.PollWithContext(c.ctx, 5*time.Second, time.Hour,
		func(ctx context.Context) (done bool, err error) {
			j, err := c.k8sClientSet.BatchV1().Jobs(job.Namespace).Get(ctx, job.Name, metav1.GetOptions{})
			if err != nil {
				klog.Error("get job error, ", err, ", ", backup.Name, ", ", backup.Namespace)
				return false, err
			}

			switch {
			case j.Status.Active > 0:
				return false, nil
			case j.Status.Active == 0 && j.Status.Succeeded > 0:
				_, err := c.updateStatus(backup, v1alpha1.BackupStateReady, nil)
				if err != nil {
					klog.Error("update backup job status error, ", err, ", ", backup.Name, ", ", backup.Namespace)
					return false, err
				}
				return true, nil
			case j.Status.Active == 0 && j.Status.Failed > 0:
				_, err := c.updateStatus(backup, v1alpha1.BackupStateError, errors.New("backup job failed"))
				if err != nil {
					klog.Error("update backup job status error, ", err, ", ", backup.Name, ", ", backup.Namespace)
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

		selectChan:
			for {
				select {
				case <-c.ctx.Done():
					watcher.Stop()
					return
				case event, ok := <-watcher.ResultChan():
					if !ok {
						watcher.Stop()
						break selectChan
					}

					switch event.Type {
					case watch.Deleted:
						watchedJob, ok := event.Object.(*batchv1.Job)
						if !ok {
							klog.Error("unexpected object")
							continue
						}

						if watchedJob.Name != job.Name || watchedJob.Namespace != job.Namespace {
							klog.Error("watch object error, ", watchedJob.Name, ", ", watchedJob.Namespace)
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
					}
				}
			} // end for chan select
		}
	}()
}

func getBackupSubPath() string {
	currentTime := time.Now()
	timeStr := currentTime.Format("2006-01-02_15-04-05")
	return timeStr
}

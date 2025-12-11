package handlers

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	sysv1 "olares.com/backup-server/pkg/apis/sys.bytetrade.io/v1"
	"olares.com/backup-server/pkg/client"
	"olares.com/backup-server/pkg/constant"
	"olares.com/backup-server/pkg/converter"
	"olares.com/backup-server/pkg/integration"
	"olares.com/backup-server/pkg/notify"
	"olares.com/backup-server/pkg/util"
	"olares.com/backup-server/pkg/util/log"
	"olares.com/backup-server/pkg/util/uuid"
)

var applicationsResource = schema.GroupVersionResource{Group: "app.bytetrade.io", Version: "v1alpha1", Resource: "applications"}

type BackupHandler struct {
	factory  client.Factory
	handlers Interface
}

func NewBackupHandler(f client.Factory, handlers Interface) *BackupHandler {
	return &BackupHandler{
		factory:  f,
		handlers: handlers,
	}
}

func (o *BackupHandler) CheckAppInstalled(appName string) bool {
	client, err := o.factory.DynamicClient()
	if err != nil {
		return false
	}

	obj, err := client.Resource(applicationsResource).Get(context.Background(), appName, metav1.GetOptions{})
	if err != nil || obj == nil {
		return false
	}
	return true
}

func (o *BackupHandler) DeleteBackup(ctx context.Context, backup *sysv1.Backup) error {
	if err := o.handlers.GetSnapshotHandler().DeleteSnapshots(ctx, backup.Name); err != nil {
		log.Errorf("delete backup %s snapshots error: %v", backup.Name, err)
		return err
	}

	locationConfig, _ := GetBackupLocationConfig(backup)
	location := locationConfig["location"]
	locationConfigName := locationConfig["name"]
	if location == constant.BackupLocationSpace.String() {
		spaceToken, err := integration.IntegrationManager().GetIntegrationSpaceToken(ctx, backup.Spec.Owner, locationConfigName)
		if err != nil {
			return err
		}

		_ = notify.NotifyStopBackup(ctx, constant.SyncServerURL, spaceToken.OlaresDid, spaceToken.AccessToken, backup.Name)
	}

	return o.delete(ctx, backup)
}

func (o *BackupHandler) UpdateBackupPolicy(ctx context.Context, backup *sysv1.Backup) error {
	return o.update(ctx, backup)
}

func (o *BackupHandler) Delete(ctx context.Context, backup *sysv1.Backup) error {
	backup.Spec.Deleted = true
	return o.update(ctx, backup)
}

func (o *BackupHandler) Enabled(ctx context.Context, backup *sysv1.Backup, data string) error {
	var enabled bool
	if data == constant.BackupResume {
		enabled = true
	} else {
		enabled = false
	}
	backup.Spec.BackupPolicy.Enabled = enabled
	return o.update(ctx, backup)
}

func (o *BackupHandler) UpdateNotifyState(ctx context.Context, backupId string, notified bool) error {
	backup, err := o.GetById(ctx, backupId)
	if err != nil {
		return err
	}
	backup.Spec.Notified = notified

	return o.update(ctx, backup)

}

func (o *BackupHandler) UpdateTotalSize(ctx context.Context, backup *sysv1.Backup, totalSize, restoreSize uint64, newLocation, newLocationData string) error {
	extra := backup.Spec.Extra
	if extra == nil {
		extra = make(map[string]string)
	}
	extra["size_updated"] = fmt.Sprintf("%d", time.Now().Unix())
	backup.Spec.Size = &totalSize
	backup.Spec.RestoreSize = &restoreSize
	backup.Spec.Notified = false
	backup.Spec.Extra = extra

	if newLocation != "" {
		backup.Spec.Location[newLocation] = newLocationData
	}

	return o.update(ctx, backup)
}

func (o *BackupHandler) ListBackups(ctx context.Context, owner string, offset, limit int64, labelSelector string, fieldSelector string) (*sysv1.BackupList, error) {
	c, err := o.factory.Sysv1Client()
	if err != nil {
		return nil, err
	}

	var listOpts = metav1.ListOptions{}
	if labelSelector != "" {
		listOpts.LabelSelector = labelSelector
	}
	if fieldSelector != "" {
		listOpts.FieldSelector = fieldSelector
	}

	backups, err := c.SysV1().Backups(constant.DefaultNamespaceOsFramework).List(ctx, listOpts)

	if err != nil {
		return nil, err
	}

	if backups == nil || backups.Items == nil || len(backups.Items) == 0 {
		return nil, fmt.Errorf("backups not exists")
	}

	var filteredItems []sysv1.Backup
	for _, item := range backups.Items {
		if !item.Spec.Deleted {
			filteredItems = append(filteredItems, item)
		}
	}

	backups.Items = filteredItems

	if len(backups.Items) == 0 {
		return nil, fmt.Errorf("no active backups exist")
	}

	sort.Slice(backups.Items, func(i, j int) bool {
		return !backups.Items[i].ObjectMeta.CreationTimestamp.Before(&backups.Items[j].ObjectMeta.CreationTimestamp)
	})

	return backups, nil
}

func (o *BackupHandler) GetById(ctx context.Context, id string) (*sysv1.Backup, error) {
	c, err := o.factory.Sysv1Client()
	if err != nil {
		return nil, err
	}

	return c.SysV1().Backups(constant.DefaultNamespaceOsFramework).Get(ctx, id, metav1.GetOptions{})
}

func (o *BackupHandler) GetByLabel(ctx context.Context, label string) (*sysv1.Backup, error) {
	var getCtx, cancel = context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	c, err := o.factory.Sysv1Client()
	if err != nil {
		return nil, err
	}
	backups, err := c.SysV1().Backups(constant.DefaultNamespaceOsFramework).List(getCtx, metav1.ListOptions{
		LabelSelector: label,
	})

	if err != nil {
		return nil, err
	}

	if backups == nil || backups.Items == nil || len(backups.Items) == 0 {
		return nil, apierrors.NewNotFound(sysv1.Resource("Backup"), label)
	}

	return &backups.Items[0], nil
}

func (o *BackupHandler) Create(ctx context.Context, owner string, backupName string, backupPath string, backupType string, backupAppTypeName string, backupSpec *sysv1.BackupSpec) (*sysv1.Backup, error) {
	backupName = strings.TrimSpace(backupName)
	var backupId = uuid.NewUUID()
RETRY:

	var labels = make(map[string]string)
	labels["owner"] = owner
	labels["name"] = util.MD5(backupName)
	labels["policy"] = util.MD5(backupSpec.BackupPolicy.TimesOfDay)
	labels["type"] = backupType

	if backupType == constant.BackupTypeFile {
		labels["path"] = util.MD5(backupPath)
	} else {
		labels["appTypeName"] = backupAppTypeName
	}

	var backup = &sysv1.Backup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      backupId,
			Namespace: constant.DefaultNamespaceOsFramework,
			Labels:    labels,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Backup",
			APIVersion: sysv1.SchemeGroupVersion.String(),
		},
		Spec: *backupSpec,
	}

	obj, err := converter.ToUnstructured(backup)
	if err != nil {
		return nil, err
	}

	res := unstructured.Unstructured{Object: obj}
	res.SetGroupVersionKind(backup.GroupVersionKind())

	dynamicClient, err := o.factory.DynamicClient()
	if err != nil {
		return nil, err
	}

	_, err = dynamicClient.Resource(constant.BackupGVR).Namespace(constant.DefaultNamespaceOsFramework).
		Apply(ctx, res.GetName(), &res, metav1.ApplyOptions{Force: true, FieldManager: constant.BackupController})

	if err != nil && apierrors.IsConflict(err) {
		goto RETRY
	} else if err != nil {
		return nil, errors.WithStack(err)
	}

	return backup, nil
}

func (o *BackupHandler) GetBackupIdForLabels(backups *sysv1.BackupList) []string {
	var labels []string

	for _, backup := range backups.Items {
		labels = append(labels, fmt.Sprintf("backup-id=%s", backup.Name))
	}
	return labels
}

func (o *BackupHandler) update(ctx context.Context, backup *sysv1.Backup) error {
	sc, err := o.factory.Sysv1Client()
	if err != nil {
		return err
	}

	var getCtx, cancel = context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

RETRY:
	_, err = sc.SysV1().Backups(constant.DefaultNamespaceOsFramework).Update(getCtx, backup, metav1.UpdateOptions{
		FieldManager: constant.BackupController,
	})

	if err != nil && apierrors.IsConflict(err) {
		log.Warnf("update backup %s spec retry", backup.Spec.Name)
		goto RETRY
	} else if err != nil {
		return errors.WithStack(fmt.Errorf("update backup error: %v", err))
	}

	return nil
}

func (o *BackupHandler) delete(ctx context.Context, backup *sysv1.Backup) error {
	sc, err := o.factory.Sysv1Client()
	if err != nil {
		return err
	}

	var getCtx, cancel = context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

RETRY:
	err = sc.SysV1().Backups(constant.DefaultNamespaceOsFramework).Delete(getCtx, backup.Name, metav1.DeleteOptions{})

	if err != nil && apierrors.IsConflict(err) {
		log.Warnf("delete backup %s spec retry", backup.Spec.Name)
		goto RETRY
	} else if err != nil {
		return errors.WithStack(fmt.Errorf("delete backup error: %v", err))
	}

	return nil
}

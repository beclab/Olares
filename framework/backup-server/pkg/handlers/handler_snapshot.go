package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
	sysv1 "olares.com/backup-server/pkg/apis/sys.bytetrade.io/v1"
	"olares.com/backup-server/pkg/client"
	"olares.com/backup-server/pkg/constant"
	"olares.com/backup-server/pkg/util"
	"olares.com/backup-server/pkg/util/log"
	"olares.com/backup-server/pkg/util/pointer"
	"olares.com/backup-server/pkg/util/uuid"
	"olares.com/backups-sdk/pkg/restic"
	backupssdkrestic "olares.com/backups-sdk/pkg/restic"
	backupssdkmodel "olares.com/backups-sdk/pkg/storage/model"
)

type SnapshotHandler struct {
	factory  client.Factory
	cron     *cron.Cron
	handlers Interface
	sync.Mutex
}

type backupJob struct {
	name string
	f    func()
}

func (b backupJob) Run() { b.f() }

func NewSnapshotHandler(f client.Factory, handlers Interface) *SnapshotHandler {
	c := cron.New()
	c.Start()

	return &SnapshotHandler{
		factory:  f,
		cron:     c,
		handlers: handlers,
	}
}

func (o *SnapshotHandler) UpdateProgress(ctx context.Context, snapshotId string, percent int) error {
	snapshot, err := o.GetById(ctx, snapshotId)
	if err != nil {
		return err
	}

	if snapshot == nil {
		return fmt.Errorf("snapshot %s not found", snapshotId)
	}

	if *snapshot.Spec.Phase != constant.Running.String() {
		return fmt.Errorf("snapshot %s is not Running, phase: %s", snapshotId, *snapshot.Spec.Phase)
	}

	snapshot.Spec.Progress = percent
	return o.update(ctx, snapshot)
}

func (o *SnapshotHandler) DeleteSnapshots(ctx context.Context, backupId string) error {
	var getCtx, cancel = context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	var labelSelector = fmt.Sprintf("backup-id=%s", backupId)

	c, err := o.factory.Sysv1Client()
	if err != nil {
		return err
	}

	return c.SysV1().Snapshots(constant.DefaultNamespaceOsFramework).DeleteCollection(getCtx, metav1.DeleteOptions{}, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
}

func (o *SnapshotHandler) UpdateNotifyResultState(ctx context.Context, snapshot *sysv1.Snapshot) error {
	extra := snapshot.Spec.Extra
	if extra == nil {
		return fmt.Errorf("snapshot %s extra is nil", snapshot.Name)
	}

	notifyState, ok := snapshot.Spec.Extra["push"]
	if !ok {
		return fmt.Errorf("snapshot %s extra push is nil", snapshot.Name)
	}

	var s *SnapshotNotifyState
	if err := json.Unmarshal([]byte(notifyState), &s); err != nil {
		return err
	}

	s.Result = true
	snapshot.Spec.Extra["push"] = util.ToJSON(s)

	return o.update(ctx, snapshot)
}

func (o *SnapshotHandler) UpdatePhase(ctx context.Context, snapshotId string, phase string, msg string) error {
	snapshot, err := o.GetById(ctx, snapshotId)
	if err != nil {
		return err
	}

	var now = pointer.Time()

	if phase == constant.Running.String() {
		snapshot.Spec.StartAt = now
	} else {
		snapshot.Spec.EndAt = now
	}

	snapshot.Spec.Phase = pointer.String(phase)
	snapshot.Spec.Message = pointer.String(msg)

	return o.update(ctx, snapshot)
}

func (o *SnapshotHandler) ListSnapshots(ctx context.Context, offset, limit int64, labelSelector string, fieldSelector string) (*sysv1.SnapshotList, error) {
	c, err := o.factory.Sysv1Client()
	if err != nil {
		return nil, err
	}

	var listOptions = metav1.ListOptions{}

	if labelSelector != "" {
		listOptions.LabelSelector = labelSelector
	}
	if fieldSelector != "" {
		listOptions.FieldSelector = fieldSelector
	}

	l, err := c.SysV1().Snapshots(constant.DefaultNamespaceOsFramework).List(ctx, listOptions)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if l == nil || l.Items == nil || len(l.Items) == 0 {
		return nil, nil
	}

	sort.Slice(l.Items, func(i, j int) bool {
		return !l.Items[i].ObjectMeta.CreationTimestamp.Before(&l.Items[j].ObjectMeta.CreationTimestamp)
	})
	return l, nil
}

func (o *SnapshotHandler) RemoveFromSchedule(ctx context.Context, backup *sysv1.Backup) {
	o.Lock()
	defer o.Unlock()
	entries := o.cron.Entries()
	for _, e := range entries {
		if e.Job.(backupJob).name == backup.Name {
			log.Infof("remove cron job, name: %s, id: %s", backup.Spec.Name, backup.Name)
			o.cron.Remove(e.ID)
		}
	}
}

func (o *SnapshotHandler) CreateSchedule(ctx context.Context, backup *sysv1.Backup, schedule string, paused bool) error {
	o.Lock()
	defer o.Unlock()
	log.Infof("create snapshot schedule, name: %s, frequency: %s, schedule: %s", backup.Spec.Name, backup.Spec.BackupPolicy.SnapshotFrequency, schedule)

	entries := o.cron.Entries()
	for _, e := range entries {
		if e.Job.(backupJob).name == backup.Name {
			log.Info("remove prev cron job to apply new one")
			o.cron.Remove(e.ID)
		}
	}

	_, err := o.cron.AddJob(schedule, backupJob{
		name: backup.Name,
		f: func() {
			log.Infof("prepare to create snapshot task, name: %s, id: %s", backup.Spec.Name, backup.Name)

			var location string
			for k := range backup.Spec.Location {
				location = k
				break
			}

			_, err := o.Create(ctx, backup, location)
			if err != nil {
				log.Error("create snapshot task error, ", err)
			}
		},
	})

	if err != nil {
		log.Error("add snapshot schedule error, ", err)
	}

	return err
}

func (o *SnapshotHandler) Create(ctx context.Context, backup *sysv1.Backup, location string) (*sysv1.Snapshot, error) {
	c, err := o.factory.Sysv1Client()
	if err != nil {
		return nil, err
	}
	var createAt = pointer.Time()
	var name = uuid.NewUUID()
	var phase = constant.Pending.String()
	var parseSnapshotType = ParseSnapshotType(constant.UnKnownBackup)
	var backupType = GetBackupType(backup)
	var pushState = &SnapshotNotifyState{
		Prepare:  false,
		Progress: false,
		Result:   false,
	}

	var snapshot = &sysv1.Snapshot{
		TypeMeta: metav1.TypeMeta{
			Kind:       constant.KindSnapshot,
			APIVersion: sysv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: constant.DefaultNamespaceOsFramework,
			Labels: map[string]string{
				"backup-id": backup.Name,
				"type":      backupType,
			},
		},
		Spec: sysv1.SnapshotSpec{
			BackupId:     backup.Name,
			Location:     location,
			SnapshotType: parseSnapshotType,
			CreateAt:     createAt,
			StartAt:      createAt,
			Progress:     0,
			Phase:        &phase,
			Extra: map[string]string{
				"push": util.ToJSON(pushState),
			},
		},
	}

	created, err := c.SysV1().Snapshots(constant.DefaultNamespaceOsFramework).Create(ctx, snapshot, metav1.CreateOptions{FieldManager: constant.SnapshotController})
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (o *SnapshotHandler) GetById(ctx context.Context, snapshotId string) (*sysv1.Snapshot, error) {
	var getCtx, cancel = context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	c, err := o.factory.Sysv1Client()
	if err != nil {
		return nil, err
	}

	return c.SysV1().Snapshots(constant.DefaultNamespaceOsFramework).Get(getCtx, snapshotId, metav1.GetOptions{})
}

func (o *SnapshotHandler) GetRunningSnapshot(ctx context.Context, backupId string) (bool, error) {
	// check exists
	var labelSelector = fmt.Sprintf("backup-id=%s", backupId)
	snapshots, err := o.ListSnapshots(ctx, 0, 0, labelSelector, "") // find all snapshots by backupName
	if err != nil {
		return false, err
	}

	if snapshots == nil || len(snapshots.Items) == 0 {
		return false, nil
	}

	for _, snapshot := range snapshots.Items {
		if util.ListContains([]string{constant.Running.String()}, *snapshot.Spec.Phase) {
			return true, nil
		}
	}

	return false, nil
}

func (o *SnapshotHandler) GetSnapshotType(ctx context.Context, backupId string) (snapshotType string, err error) {
	var snapshots *sysv1.SnapshotList
	var labelSelector = fmt.Sprintf("backup-id=%s", backupId)
	snapshotType = constant.FullyBackup

	snapshots, err = o.ListSnapshots(ctx, 0, 0, labelSelector, "") // find all snapshots by backupName
	if err != nil {
		return
	}

	if snapshots == nil || len(snapshots.Items) == 0 {
		return
	}

	for _, snapshot := range snapshots.Items {
		if util.ListContains([]string{constant.Completed.String()}, *snapshot.Spec.Phase) {
			snapshotType = constant.IncrementalBackup
			return
		}
	}

	return
}

func (o *SnapshotHandler) UpdateBackupResult(ctx context.Context, snapshot *sysv1.Snapshot) error {
	return o.update(ctx, snapshot)
}

func (o *SnapshotHandler) update(ctx context.Context, snapshot *sysv1.Snapshot) error {
	// todo
	sc, err := o.factory.Sysv1Client()
	if err != nil {
		return err
	}

	var getCtx, cancel = context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

RETRY:
	_, err = sc.SysV1().Snapshots(constant.DefaultNamespaceOsFramework).Update(getCtx, snapshot, metav1.UpdateOptions{
		FieldManager: constant.SnapshotController,
	})

	if err != nil && apierrors.IsConflict(err) {
		log.Warnf("update snapshot %s spec retry", snapshot.Name)
		goto RETRY
	} else if err != nil {
		return errors.WithStack(fmt.Errorf("update snapshot error: %v", err))
	}

	return nil
}

func (o *SnapshotHandler) SortSnapshotList(snapshots *restic.SnapshotList) (*sysv1.SnapshotList, string, string) {
	var first = snapshots.First()
	var backupTypeTag = GetBackupTypeFromTags(first.Tags)
	var backupTypeAppName = GetBackupTypeAppName(first.Tags)

	var result = &sysv1.SnapshotList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Snapshotlist",
			APIVersion: sysv1.SchemeGroupVersion.String(),
		},
		Items: []sysv1.Snapshot{},
	}

	for _, snapshot := range *snapshots {
		createAt, _ := time.Parse(time.RFC3339, snapshot.Summary.BackupStart)
		var path string
		if snapshot.Paths != nil && len(snapshot.Paths) > 0 {
			path = snapshot.Paths[0]
		}
		var item = sysv1.Snapshot{
			TypeMeta: metav1.TypeMeta{
				Kind:       constant.KindSnapshot,
				APIVersion: sysv1.SchemeGroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				CreationTimestamp: metav1.Time{Time: createAt},
			},
			Spec: sysv1.SnapshotSpec{
				SnapshotId: pointer.String(snapshot.Id),
				CreateAt:   &metav1.Time{Time: createAt},
				Location:   path,
				Size:       pointer.UInt64(uint64(snapshot.Summary.TotalBytesProcessed)),
			},
		}

		result.Items = append(result.Items, item)
	}

	return result, backupTypeTag, backupTypeAppName
}

// --

func (o *SnapshotHandler) GetOlaresId(owner string) (string, error) {
	dynamicClient, err := o.factory.DynamicClient()
	if err != nil {
		return "", errors.WithStack(fmt.Errorf("get dynamic client error: %v", err))
	}

	var backoff = wait.Backoff{
		Duration: 2 * time.Second,
		Factor:   2,
		Jitter:   0.1,
		Steps:    5,
	}

	var olaresName string
	if err := retry.OnError(backoff, func(err error) bool {
		return true
	}, func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		unstructuredUser, err := dynamicClient.Resource(constant.UsersGVR).Get(ctx, owner, metav1.GetOptions{})
		if err != nil {
			return errors.WithStack(fmt.Errorf("get user error: %v", err))
		}
		obj := unstructuredUser.UnstructuredContent()
		olaresName, _, err = unstructured.NestedString(obj, "spec", "email")
		if err != nil {
			return errors.WithStack(fmt.Errorf("get user nested string error: %v", err))
		}
		return nil
	}); err != nil {
		return "", errors.WithStack(err)
	}

	return olaresName, nil
}

func (o *SnapshotHandler) ParseSnapshotInfo(snapshot *sysv1.Snapshot) (*backupssdkmodel.StorageInfo, *backupssdkrestic.SummaryOutput) {
	var storageInfo *backupssdkmodel.StorageInfo
	if snapshot.Spec.Extra != nil {
		storage, ok := snapshot.Spec.Extra["storage"]
		if ok && len(storage) > 0 {
			json.Unmarshal([]byte(storage), &storageInfo)
		}
	}

	var backupOutput *backupssdkrestic.SummaryOutput
	if snapshot.Spec.ResticMessage != nil && *snapshot.Spec.ResticMessage != "" {
		json.Unmarshal([]byte(*snapshot.Spec.ResticMessage), &backupOutput)
	}

	return storageInfo, backupOutput
}

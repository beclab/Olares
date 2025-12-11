package worker

import (
	"olares.com/backup-server/pkg/storage"
)

type TaskInterface interface {
	GetTaskType() TaskType
	Run()
	IsCanceled() bool
	Cancel()
}

type BackupTask struct {
	*BaseTask
	backupId   string
	snapshotId string
	backup     *storage.StorageBackup
}

func (b *BackupTask) GetTaskType() TaskType {
	return b.taskType
}

func (b *BackupTask) IsCanceled() bool {
	return b.BaseTask.IsCanceled()
}

func (b *BackupTask) Run() {
	b.backup.RunBackup()
}

func (b *BackupTask) Cancel() {
	b.BaseTask.cancel()
}

func (b *BackupTask) GetBackupId() string {
	return b.backupId
}

func (b *BackupTask) GetSnapshotId() string {
	return b.snapshotId
}

type RestoreTask struct {
	*BaseTask
	restoreId string
	restore   *storage.StorageRestore
}

func (r *RestoreTask) GetTaskType() TaskType {
	return r.taskType
}

func (r *RestoreTask) IsCanceled() bool {
	return r.BaseTask.IsCanceled()
}

func (r *RestoreTask) Run() {
	r.restore.RunRestore()
}

func (r *RestoreTask) Cancel() {
	r.BaseTask.cancel()
}

func (r *RestoreTask) GetRestoreId() string {
	return r.restoreId
}

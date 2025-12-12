package worker

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	pond "github.com/alitto/pond/v2"
	"olares.com/backup-server/pkg/constant"
	"olares.com/backup-server/pkg/handlers"
	"olares.com/backup-server/pkg/storage"
	"olares.com/backup-server/pkg/util/log"
	"olares.com/backup-server/pkg/watchers/notification"
)

var workerPool *WorkerPool

type TaskType string

const (
	BackupTaskType  TaskType = "backup"
	RestoreTaskType TaskType = "restore"
)

type WorkerPool struct {
	ctx       context.Context
	cancel    context.CancelFunc
	handlers  handlers.Interface
	userPools sync.Map
}

type TaskPool struct {
	pool       pond.Pool
	tasks      sync.Map
	activeTask atomic.Pointer[any]
	owner      string
}

func (tp *TaskPool) setActiveTask(task interface{}) {
	tp.activeTask.Store(&task)
}

type UserPool struct {
	owner       string
	executePool *TaskPool
}

type BaseTask struct {
	ctx      context.Context
	cancel   context.CancelFunc
	id       string
	owner    string
	taskType TaskType
	canceled bool
	mutex    sync.RWMutex
}

func (t *BaseTask) IsCanceled() bool {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.canceled
}

func newPool() pond.Pool {
	return pond.NewPool(constant.MaxConcurrency, pond.WithContext(context.Background()), pond.WithNonBlocking(constant.NonBlocking))
}

func GetWorkerPool() *WorkerPool {
	return workerPool
}

func NewWorkerPool(ctx context.Context, handlers handlers.Interface) {
	ctx, cancel := context.WithCancel(ctx)

	workerPool = &WorkerPool{
		ctx:      ctx,
		cancel:   cancel,
		handlers: handlers,
	}
}

func (w *WorkerPool) GetOrCreateUserPool(owner string) *UserPool {
	if pool, ok := w.userPools.Load(owner); ok {
		userPool := pool.(*UserPool)
		return userPool
	}

	userPool := &UserPool{
		owner: owner,
		executePool: &TaskPool{
			pool:  newPool(),
			owner: owner,
		},
	}

	w.userPools.Store(owner, userPool)

	return userPool
}

func (w *WorkerPool) AddBackupTask(owner, backupId, snapshotId string) error {
	if err := w.ExistsTask(owner, backupId, snapshotId); err != nil {
		return err
	}

	if backupId == "" || snapshotId == "" {
		return fmt.Errorf("backup task parms invalid, backupId: %s, snapshotId: %s", backupId, snapshotId)
	}

	var taskType = BackupTaskType
	var taskId = fmt.Sprintf("%s_%s", backupId, snapshotId)
	var traceId = snapshotId

	var ctxTask, cancelTask = context.WithCancel(w.ctx)
	var ctxBase context.Context = context.WithValue(ctxTask, constant.TraceId, traceId)

	var newTask = &BaseTask{
		ctx:      ctxBase,
		cancel:   cancelTask,
		owner:    owner,
		id:       taskId,
		taskType: taskType,
	}

	var backup = &storage.StorageBackup{
		Ctx:              ctxBase,
		Handlers:         w.handlers,
		SnapshotId:       snapshotId,
		LastProgressTime: time.Now(),
	}

	var backupTask = &BackupTask{
		BaseTask:   newTask,
		backupId:   backupId,
		snapshotId: snapshotId,
		backup:     backup,
	}

	return w.addTask(owner, taskId, backupTask)
}

func (w *WorkerPool) AddRestoreTask(owner, restoreId string) error {
	if restoreId == "" {
		return fmt.Errorf("restore task parms invalid, restoreId: %s", restoreId)
	}

	var taskType = RestoreTaskType
	var taskId = restoreId
	var traceId = restoreId

	var ctxTask, cancelTask = context.WithCancel(w.ctx)
	var ctxBase context.Context = context.WithValue(ctxTask, constant.TraceId, traceId)

	var newTask = &BaseTask{
		ctx:      ctxBase,
		cancel:   cancelTask,
		owner:    owner,
		id:       taskId,
		taskType: taskType,
	}

	var restore = &storage.StorageRestore{
		Ctx:              ctxBase,
		Handlers:         w.handlers,
		RestoreId:        restoreId,
		LastProgressTime: time.Now(),
	}

	var restoreTask = &RestoreTask{
		BaseTask:  newTask,
		restoreId: restoreId,
		restore:   restore,
	}

	return w.addTask(owner, taskId, restoreTask)
}

func (w *WorkerPool) addTask(owner, taskId string, task TaskInterface) error {
	userPool := w.GetOrCreateUserPool(owner)
	taskPool := userPool.executePool
	taskType := task.GetTaskType()

	taskPool.tasks.Store(taskId, task)
	taskFn := func() {
		defer func() {
			taskPool.tasks.Delete(taskId)
			taskPool.setActiveTask(nil)

			if r := recover(); r != nil {
				log.Errorf("[worker] task panic: %v, owner: %s, taskId: %s, taskType: %s", r, owner, taskId, taskType)
			}
		}()

		if task.IsCanceled() {
			return
		}

		log.Infof("[worker] task start, owner: %s, taskId: %s, taskType: %s", owner, taskId, taskType)
		taskPool.setActiveTask(task)

		task.Run()
	}

	_, ok := taskPool.pool.TrySubmit(taskFn)
	if !ok {
		task.Cancel()
		taskPool.tasks.Delete(taskId)
		return fmt.Errorf("[worker] task try sumit failed, owner: %s, queuesize: %d, waitings: %d", owner, taskPool.pool.QueueSize(), taskPool.pool.WaitingTasks())
	}

	return nil
}

func (w *WorkerPool) ExistsTask(owner string, backupId, snapshotId string) error {
	poolObj, ok := w.userPools.Load(owner)
	if !ok {
		log.Warnf("[worker] no backup tasks found for owner: %s", owner)
		return nil
	}

	userPool := poolObj.(*UserPool)
	taskPool := userPool.executePool

	var snapshotExists bool

	taskPool.tasks.Range(func(key, value interface{}) bool {
		task := value.(TaskInterface)
		if task != nil {
			if task.GetTaskType() == BackupTaskType {
				backupTask := task.(*BackupTask)
				if backupTask.backupId == backupId && !backupTask.IsCanceled() {
					snapshotExists = true
					return false
				}
			}
		}
		return true
	})

	if snapshotExists {
		return fmt.Errorf(constant.MessageTaskExists)
	}

	return nil
}

func (w *WorkerPool) CancelBackup(owner, backupId string) {
	log.Infof("[worker] cancel backup, owner: %s, backupId: %s", owner, backupId)

	poolObj, ok := w.userPools.Load(owner)
	if !ok {
		log.Warnf("[worker] no backup tasks found for owner: %s", owner)
		return
	}

	userPool := poolObj.(*UserPool)
	taskPool := userPool.executePool

	taskPool.tasks.Range(func(key, value interface{}) bool {
		task := value.(TaskInterface)
		if task != nil {
			if task.GetTaskType() == BackupTaskType {
				backupTask := task.(*BackupTask)
				if backupTask.backupId == backupId {
					backupTask.BaseTask.canceled = true
				}
			}
		}

		return true
	})

	activeTask := taskPool.activeTask.Load()
	if activeTask != nil && *activeTask != nil {
		backupTask := (*activeTask).(*BackupTask)
		if backupTask.backupId == backupId {
			w.sendBackupCanceledEvent(backupTask.owner, backupTask.backupId, backupTask.snapshotId)
			backupTask.BaseTask.cancel()
		}
	}
}

func (w *WorkerPool) CancelSnapshot(owner, snapshotId string) {
	log.Infof("[worker] cancel snapshot, owner: %s, snapshotId: %s", owner, snapshotId)

	poolObj, ok := w.userPools.Load(owner)
	if !ok {
		log.Warnf("[worker] no backup tasks found for owner: %s", owner)
		return
	}

	userPool := poolObj.(*UserPool)
	taskPool := userPool.executePool

	taskPool.tasks.Range(func(key, value interface{}) bool {
		task := value.(TaskInterface)
		if task != nil {
			if task.GetTaskType() == BackupTaskType {
				backupTask := task.(*BackupTask)
				if backupTask.snapshotId == snapshotId {
					backupTask.BaseTask.canceled = true
				}
			}
		}

		return true
	})

	activeTask := taskPool.activeTask.Load()
	if activeTask != nil && *activeTask != nil {
		backupTask := (*activeTask).(*BackupTask)
		if backupTask.snapshotId == snapshotId {
			w.sendBackupCanceledEvent(backupTask.owner, backupTask.backupId, backupTask.snapshotId)
			backupTask.BaseTask.cancel()
		}
	}
}

func (w *WorkerPool) CancelRestore(owner, restoreId string) {

	log.Infof("[worker] cancel restore, owner: %s, restoreId: %s", owner, restoreId)

	poolObj, ok := w.userPools.Load(owner)
	if !ok {
		log.Warn("[worker] no restore tasks found for owner: %s", owner)
		return
	}

	userPool := poolObj.(*UserPool)
	taskPool := userPool.executePool

	taskPool.tasks.Range(func(key, value interface{}) bool {
		task := value.(TaskInterface)
		if task != nil {
			if task.GetTaskType() == RestoreTaskType {
				restoreTask := task.(*RestoreTask)
				if restoreTask.restoreId == restoreId {
					restoreTask.BaseTask.canceled = true
				}
			}
		}

		return true
	})

	activeTask := taskPool.activeTask.Load()
	if activeTask != nil && *activeTask != nil {
		restoreTask := (*activeTask).(*RestoreTask)
		if restoreTask.restoreId == restoreId {
			w.sendRestoreCanceledEvent(restoreTask.owner, restoreTask.restoreId)
			restoreTask.BaseTask.cancel()
		}
	}
}

func (w *WorkerPool) sendBackupCanceledEvent(owner string, backupId string, snapshotId string) {
	var data = map[string]interface{}{
		"id":       snapshotId,
		"type":     constant.MessageTypeBackup,
		"backupId": backupId,
		"endat":    time.Now().Unix(),
		"status":   constant.Canceled.String(),
		"message":  "",
	}
	notification.DataSender.Send(owner, data)
}

func (w *WorkerPool) sendRestoreCanceledEvent(owner string, restoreId string) {
	var data = map[string]interface{}{
		"id":      restoreId,
		"type":    constant.MessageTypeRestore,
		"endat":   time.Now().Unix(),
		"status":  constant.Canceled.String(),
		"message": "",
	}
	notification.DataSender.Send(owner, data)
}

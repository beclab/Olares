package userspace

import (
	"context"
	"sync"
)

const (
	UserCreating = "Creating"
	UserDeleting = "Deleting"
	UserCreated  = "Created"
	UserDeleted  = "Deleted"
	UserFailed   = "Failed"
)

type Manager struct {
	managerCtx    context.Context
	runningTasks  map[string]*Task
	completedTask map[string]*TaskResult // user's latest completed or failed task result
	cond          *sync.Cond
}

func NewManager(ctx context.Context) *Manager {
	um := &Manager{
		managerCtx: ctx,
	}
	um.completedTask = make(map[string]*TaskResult)
	um.runningTasks = make(map[string]*Task)
	um.cond = sync.NewCond(&sync.Mutex{})
	return um
}

//func (um *Manager) CreateUserApps(client *clientset.ClientSet, config *rest.Config, user string) (*Task, error) {
//	task, ok := um.runningTasks[user]
//	if ok {
//		if task.name != UserCreating {
//			//task.cancel()
//			return nil, errors.New("latest deleting task of user incompleted")
//		}
//		return task, nil
//	}
//
//	taskCtx, cancel := context.WithCancel(um.managerCtx)
//	task = &Task{
//		user:   user,
//		name:   UserCreating,
//		ctx:    taskCtx,
//		done:   make(chan struct{}),
//		cancel: cancel,
//	}
//
//	task.Action = func(ctx context.Context, user string) error {
//		creator := NewCreator(client, config, user)
//		desktop, wizard, err := creator.CreateUserApps(ctx)
//
//		if err == nil {
//			um.completedTask[user] = &TaskResult{
//				Name: task.name,
//				Values: []int32{
//					desktop,
//					wizard,
//				},
//				Status: UserCreated,
//			}
//		}
//		um.cond.L.Lock()
//		defer um.cond.L.Unlock()
//		delete(um.runningTasks, user)
//		if len(um.runningTasks) == 0 {
//			um.cond.Signal()
//		}
//		return err
//	}
//
//	task.Error = func(msg string, err error, args ...any) {
//		um.completedTask[user] = &TaskResult{
//			Name:  task.name,
//			Error: err,
//		}
//		klog.Error(msg, err, args)
//	}
//
//	task.Do()
//	um.runningTasks[user] = task
//
//	return task, nil
//}

//func (um *Manager) DeleteUserApps(client *clientset.ClientSet, config *rest.Config, user string) (*Task, error) {
//	task, ok := um.runningTasks[user]
//	if ok {
//		if task.name != UserDeleting {
//			task.cancel()
//		} else {
//			return task, nil
//		}
//	}
//
//	taskCtx, cancel := context.WithCancel(um.managerCtx)
//	task = &Task{
//		user:   user,
//		name:   UserDeleting,
//		ctx:    taskCtx,
//		done:   make(chan struct{}),
//		cancel: cancel,
//	}
//
//	task.Action = func(ctx context.Context, user string) error {
//		deleter := NewDeleter(client, config, user)
//		err := deleter.DeleteUserApps(ctx)
//
//		if err == nil {
//			um.completedTask[user] = &TaskResult{
//				Name:   task.name,
//				Status: UserDeleted,
//			}
//		}
//
//		delete(um.runningTasks, user)
//		return err
//	}
//
//	task.Error = func(msg string, err error, args ...any) {
//		um.completedTask[user] = &TaskResult{
//			Name:  task.name,
//			Error: err,
//		}
//		klog.Error(msg, err, args)
//	}
//
//	task.Do()
//	um.runningTasks[user] = task
//
//	return task, nil
//
//}

//func (um *Manager) TaskStatus(user string) *TaskResult {
//	task, ok := um.runningTasks[user]
//
//	if ok {
//		return &TaskResult{
//			Name:   task.name,
//			Status: task.name,
//		}
//	}
//
//	taskRes := um.completedTask[user]
//
//	return taskRes
//}

//func (um *Manager) HasRunningTask() bool {
//	return len(um.runningTasks) > 0
//}
//
//func (um *Manager) Wait() {
//	um.cond.L.Lock()
//	defer um.cond.L.Unlock()
//	um.cond.Wait()
//}

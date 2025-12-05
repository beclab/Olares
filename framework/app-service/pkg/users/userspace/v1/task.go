package userspace

import "context"

type Task struct {
	user   string
	name   string
	Action func(ctx context.Context, user string) error
	ctx    context.Context
	Error  func(msg string, err error, args ...any)
	done   chan struct{}
	cancel context.CancelFunc
}

type TaskResult struct {
	Name   string
	Values interface{}
	Error  error
	Status string
}

func (t *Task) Do() {
	go func() {
		err := t.Action(t.ctx, t.user)
		if err != nil {
			t.Error("Running task error, ", err, t.user, t.name)
		}

		close(t.done)
	}()
}

func (t *Task) Done() <-chan struct{} {
	return t.done
}

func (t *Task) Cancel() {
	t.cancel()
}

package appstate

import (
	"context"
	"sync"
	"time"

	appsv1 "github.com/beclab/Olares/framework/app-service/api/app.bytetrade.io/v1alpha1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type statefulAppFactory struct {
	inProgress map[string]StatefulInProgressApp
	mu         sync.Mutex
}

var once sync.Once
var appFactory statefulAppFactory

func init() {
	once.Do(func() {
		appFactory = statefulAppFactory{
			inProgress: make(map[string]StatefulInProgressApp),
		}
	})
}

func (f *statefulAppFactory) New(
	client client.Client,
	manager *appsv1.ApplicationManager,
	ttl time.Duration,
	create func(client client.Client, manager *appsv1.ApplicationManager, ttl time.Duration) StatefulApp,
) (StatefulApp, StateError) {
	f.mu.Lock()
	defer f.mu.Unlock()

	inProgressApp, ok := f.inProgress[manager.Name]
	if ok {
		if inProgressApp.State() != manager.Status.State.String() {
			a := create(client, manager, ttl)
			if cancelOperation, ok := a.(CancelOperationApp); ok {
				return cancelOperation, nil
			}

			klog.Infof("app %s is doing something in progress, but state is not match, expected: %s, actual: %s",
				manager.Name, manager.Status.State.String(), inProgressApp.State())

			return nil, NewErrorUnknownInProgressApp(func(ctx context.Context) error {
				inProgressApp.Cleanup(ctx)

				// remove the app from the running map
				f.mu.Lock()
				delete(f.inProgress, inProgressApp.GetManager().Name)
				f.mu.Unlock()

				return nil
			})
		}

		klog.Infof("app %s is already doing operation, state: %s", manager.Name, inProgressApp.State())
		return inProgressApp, nil
	}

	return create(client, manager, ttl), nil
}

func (f *statefulAppFactory) waitForPolling(ctx context.Context, app PollableStatefulInProgressApp, finally func(err error)) {
	f.mu.Lock()
	defer f.mu.Unlock()

	_, ok := f.inProgress[app.GetManager().Name]
	if !ok {
		f.inProgress[app.GetManager().Name] = app

		go func() {
			err := app.poll(ctx)
			if err != nil {
				klog.Error("poll ", app.State(), " progress error, ", err, ", ", app.GetManager().Name)
			}

			app.stopPolling()

			klog.Error("stop polling, ", "app name: ", app.GetManager().Name)

			// remove the app from the running map
			f.mu.Lock()
			delete(f.inProgress, app.GetManager().Name)
			f.mu.Unlock()

			finally(err)
		}()

	}
}

func (f *statefulAppFactory) execAndWatch(
	ctx context.Context,
	app StatefulApp,
	exec func(ctx context.Context) (StatefulInProgressApp, error),
) (StatefulInProgressApp, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Check if the app is already running
	if existingApp, ok := f.inProgress[app.GetManager().Name]; ok {
		if existingApp.State() == app.GetManager().Status.State.String() {
			klog.Infof("app %s is already doing operation, state: %s", app.GetManager().Name, existingApp.State())
			return existingApp, nil
		}
	}

	// Execute the app and wait for it to complete
	inProgressApp, err := exec(ctx)
	if err != nil {
		return nil, err
	}

	f.inProgress[inProgressApp.GetManager().Name] = inProgressApp

	go func() {
		if done := inProgressApp.Done(); done != nil {
			<-done
			klog.Infof("app %s has completed, state: %v", inProgressApp.GetManager().Name, inProgressApp.GetManager().Status.State)
		}

		f.mu.Lock()
		delete(f.inProgress, inProgressApp.GetManager().Name)
		f.mu.Unlock()

		// updating state whatever success or failure should be done in the finally block
		// because of the new state will cause the new reconciling request,
		// it will be conflicted of the current operation
		inProgressApp.Finally()

	}()

	return inProgressApp, nil
}

func (f *statefulAppFactory) cancelOperation(name string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	if app, ok := f.inProgress[name]; ok {
		app.Cleanup(context.Background())
		delete(f.inProgress, name)
		return true
	}

	return false
}

func (f *statefulAppFactory) countInProgressApp(state string) int {
	f.mu.Lock()
	defer f.mu.Unlock()

	count := 0
	for _, app := range f.inProgress {
		if app.State() == state {
			count++
		}
	}

	return count
}

func (f *statefulAppFactory) addLimitedStatefulApp(ctx context.Context, limited func() (bool, error), add func() error) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if ok, err := limited(); err != nil {
		klog.Error("check limited stateful app error, ", err)
		return false, err
	} else if !ok {
		return false, nil
	} else {
		if err := add(); err != nil {
			klog.Error("add limited stateful app error, ", err)
			return false, err
		}
	}

	return true, nil
}

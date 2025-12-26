package handlers

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
	sysv1 "olares.com/backup-server/pkg/apis/sys.bytetrade.io/v1"
	"olares.com/backup-server/pkg/client"
	"olares.com/backup-server/pkg/constant"
	"olares.com/backup-server/pkg/util"
	"olares.com/backup-server/pkg/util/log"
	"olares.com/backup-server/pkg/util/pointer"
	"olares.com/backup-server/pkg/util/uuid"
)

type RestoreHandler struct {
	factory  client.Factory
	handlers Interface
}

func NewRestoreHandler(f client.Factory, handlers Interface) *RestoreHandler {
	return &RestoreHandler{
		factory:  f,
		handlers: handlers,
	}
}

func (o *RestoreHandler) UpdateProgress(ctx context.Context, restoreId string, percent int) error {
	restore, err := o.GetById(ctx, restoreId)
	if err != nil {
		return err
	}

	if restore == nil {
		return fmt.Errorf("restore %s not found", restoreId)
	}

	if *restore.Spec.Phase != constant.Running.String() {
		return fmt.Errorf("restore %s is not Running, phase: %s", restoreId, *restore.Spec.Phase)
	}

	restore.Spec.Progress = percent
	return o.update(ctx, restore)
}

func (o *RestoreHandler) UpdatePhase(ctx context.Context, restoreId string, phase string) error {
	restore, err := o.GetById(ctx, restoreId)
	if err != nil {
		return err
	}

	if phase == constant.Running.String() {
		restore.Spec.StartAt = pointer.Time()
	}
	restore.Spec.Phase = pointer.String(phase)

	return o.Update(ctx, restoreId, &restore.Spec)
}

func (o *RestoreHandler) DeleteRestore(restoreId string) error {
	return o.Delete(context.Background(), restoreId)
}

func (o *RestoreHandler) ListRestores(ctx context.Context, owner string, offset int64, limit int64) (*sysv1.RestoreList, error) {
	c, err := o.factory.Sysv1Client()
	if err != nil {
		return nil, err
	}

	restores, err := c.SysV1().Restores(constant.DefaultNamespaceOsFramework).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("owner=%s", owner),
	})

	if err != nil {
		return nil, err
	}

	if restores == nil || restores.Items == nil || len(restores.Items) == 0 {
		return nil, fmt.Errorf("restores not exists")
	}

	sort.Slice(restores.Items, func(i, j int) bool {
		return !restores.Items[i].ObjectMeta.CreationTimestamp.Before(&restores.Items[j].ObjectMeta.CreationTimestamp)
	})

	return restores, nil
}

func (o *RestoreHandler) CreateRestore(ctx context.Context, restoreTypeName string, restoreType *RestoreType) (*sysv1.Restore, error) {
	c, err := o.factory.Sysv1Client()
	if err != nil {
		return nil, err
	}

	var createAt = pointer.Time()
	var phase = constant.Pending.String()

	var restore = &sysv1.Restore{
		TypeMeta: metav1.TypeMeta{
			Kind:       constant.KindRestore,
			APIVersion: sysv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      uuid.NewUUID(),
			Namespace: constant.DefaultNamespaceOsFramework,
			Labels: map[string]string{
				"owner": restoreType.Owner,
				"type":  restoreTypeName,
			},
		},
		Spec: sysv1.RestoreSpec{
			Owner: restoreType.Owner,
			RestoreType: map[string]string{
				restoreTypeName: util.ToJSON(restoreType),
			},
			CreateAt: createAt,
			StartAt:  createAt,
			Progress: 0,
			Phase:    &phase,
		},
	}

	created, err := c.SysV1().Restores(constant.DefaultNamespaceOsFramework).Create(ctx, restore, metav1.CreateOptions{FieldManager: constant.RestoreController})
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (o *RestoreHandler) updateRestoreFailedStatus(backupError error, restore *sysv1.Restore) error {
	restore.Spec.Phase = pointer.String(constant.Failed.String())
	restore.Spec.Message = pointer.String(backupError.Error())
	restore.Spec.EndAt = pointer.Time()

	return o.Update(context.Background(), restore.Name, &restore.Spec) // update failed
}

func (o *RestoreHandler) Update(ctx context.Context, restoreId string, restoreSpec *sysv1.RestoreSpec) error {
	sc, err := o.factory.Sysv1Client()
	if err != nil {
		return err
	}

	r, err := o.GetRestore(ctx, restoreId)
	if err != nil {
		return err
	}
	r.Spec = *restoreSpec

RETRY:
	_, err = sc.SysV1().Restores(constant.DefaultNamespaceOsFramework).Update(ctx, r, metav1.UpdateOptions{
		FieldManager: constant.RestoreController,
	})

	if err != nil && apierrors.IsConflict(err) {
		log.Warnf("update restore %s spec retry", restoreId)
		goto RETRY
	} else if err != nil {
		return errors.WithStack(fmt.Errorf("update restore error: %v", err))
	}

	return nil
}

func (o *RestoreHandler) Delete(ctx context.Context, restoreId string) error {
	sc, err := o.factory.Sysv1Client()
	if err != nil {
		return err
	}

	_, err = o.GetRestore(ctx, restoreId)
	if err != nil {
		return err
	}

RETRY:
	err = sc.SysV1().Restores(constant.DefaultNamespaceOsFramework).Delete(ctx, restoreId, metav1.DeleteOptions{})

	if err != nil && apierrors.IsConflict(err) {
		log.Warnf("delete restore %s spec retry", restoreId)
		goto RETRY
	} else if err != nil {
		return errors.WithStack(fmt.Errorf("delete restore error: %v", err))
	}

	return nil
}

func (o *RestoreHandler) GetById(ctx context.Context, id string) (*sysv1.Restore, error) {
	var ctxTimeout, cancel = context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	c, err := o.factory.Sysv1Client()
	if err != nil {
		return nil, err
	}

	restore, err := c.SysV1().Restores(constant.DefaultNamespaceOsFramework).Get(ctxTimeout, id, metav1.GetOptions{})

	if err != nil {
		return nil, err
	}

	if restore == nil {
		return nil, apierrors.NewNotFound(sysv1.Resource("Restore"), id)
	}

	return restore, nil
}

func (o *RestoreHandler) GetRestore(ctx context.Context, restoreId string) (*sysv1.Restore, error) {
	var ctxTimeout, cancel = context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	c, err := o.factory.Sysv1Client()
	if err != nil {
		return nil, err
	}

	restore, err := c.SysV1().Restores(constant.DefaultNamespaceOsFramework).Get(ctxTimeout, restoreId, metav1.GetOptions{})

	if err != nil {
		return nil, err
	}

	if restore == nil {
		return nil, nil
	}

	return restore, nil
}

func (o *RestoreHandler) SetRestorePhase(restoreId string, phase constant.Phase) error {
	c, err := o.factory.Sysv1Client()
	if err != nil {
		return err
	}

	backoff := wait.Backoff{
		Duration: 1 * time.Second,
		Factor:   2,
		Jitter:   0.1,
		Steps:    5,
	}

	if err = retry.OnError(backoff, func(err error) bool {
		return true
	}, func() error {
		var ctx, cancel = context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		r, err := c.SysV1().Restores(constant.DefaultNamespaceOsFramework).Get(ctx, restoreId, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("retry")
		}

		r.Spec.Phase = pointer.String(phase.String())
		r.Spec.EndAt = pointer.Time()

		switch phase {
		case constant.Canceled:
			r.Spec.Message = pointer.String(constant.MessageTaskCanceled)
		case constant.Failed:
			r.Spec.Message = pointer.String(constant.MessageBackupServerRestart)
		}

		_, err = c.SysV1().Restores(constant.DefaultNamespaceOsFramework).
			Update(ctx, r, metav1.UpdateOptions{})
		if err != nil && apierrors.IsConflict(err) {
			return fmt.Errorf("retry")
		} else if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (o *RestoreHandler) update(ctx context.Context, restore *sysv1.Restore) error {
	sc, err := o.factory.Sysv1Client()
	if err != nil {
		return err
	}

	var getCtx, cancel = context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

RETRY:
	_, err = sc.SysV1().Restores(constant.DefaultNamespaceOsFramework).Update(getCtx, restore, metav1.UpdateOptions{
		FieldManager: constant.RestoreController,
	})

	if err != nil && apierrors.IsConflict(err) {
		log.Warnf("update restore %s spec retry", restore.Name)
		goto RETRY
	} else if err != nil {
		return errors.WithStack(fmt.Errorf("update restore error: %v", err))
	}

	return nil
}

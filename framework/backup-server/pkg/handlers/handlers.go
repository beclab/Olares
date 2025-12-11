package handlers

import (
	"olares.com/backup-server/pkg/client"
	"olares.com/backup-server/pkg/interfaces"
	"olares.com/backup-server/pkg/watchers/notification"
)

var _ Interface = &handlers{}

type Interface interface {
	interfaces.HandlerInterface
	GetBackupHandler() *BackupHandler
	GetSnapshotHandler() *SnapshotHandler
	GetRestoreHandler() *RestoreHandler
	GetNotification() *notification.Notification
}

type handlers struct {
	BackupHandler   *BackupHandler
	SnapshotHandler *SnapshotHandler
	RestoreHandler  *RestoreHandler
	Notification    *notification.Notification
}

func NewHandler(factory client.Factory, n *notification.Notification) Interface {
	var handlers = &handlers{
		Notification: n,
	}

	handlers.BackupHandler = NewBackupHandler(factory, handlers)
	handlers.SnapshotHandler = NewSnapshotHandler(factory, handlers)
	handlers.RestoreHandler = NewRestoreHandler(factory, handlers)

	return handlers
}

func (h *handlers) GetNotification() *notification.Notification {
	return h.Notification
}

func (h *handlers) GetBackupHandler() *BackupHandler {
	return h.BackupHandler
}

func (h *handlers) GetSnapshotHandler() *SnapshotHandler {
	return h.SnapshotHandler
}

func (h *handlers) GetRestoreHandler() *RestoreHandler {
	return h.RestoreHandler
}

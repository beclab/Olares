package handlers

import (
	"olares.com/backup-server/pkg/client"
	"olares.com/backup-server/pkg/interfaces"
)

var _ Interface = &handlers{}

type Interface interface {
	interfaces.HandlerInterface
	GetBackupHandler() *BackupHandler
	GetSnapshotHandler() *SnapshotHandler
	GetRestoreHandler() *RestoreHandler
}

type handlers struct {
	BackupHandler   *BackupHandler
	SnapshotHandler *SnapshotHandler
	RestoreHandler  *RestoreHandler
}

func NewHandler(factory client.Factory) Interface {
	var handlers = &handlers{}

	handlers.BackupHandler = NewBackupHandler(factory, handlers)
	handlers.SnapshotHandler = NewSnapshotHandler(factory, handlers)
	handlers.RestoreHandler = NewRestoreHandler(factory, handlers)

	return handlers
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

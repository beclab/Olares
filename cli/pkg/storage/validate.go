package storage

import (
	"fmt"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

type ValidateModule struct {
	common.KubeModule
	Skip bool
}

func (m *ValidateModule) IsSkip() bool {
	return m.Skip
}

func (m *ValidateModule) Init() {
	m.Name = "ValidateStorageConfig"

	m.Tasks = append(m.Tasks, &task.LocalTask{
		Name:   "ValidateStorageConfig",
		Action: new(ValidateStorageConfig),
	})
}

type ValidateStorageConfig struct {
	common.KubeAction
}

func (a *ValidateStorageConfig) Execute(runtime connector.Runtime) error {
	storageConf := a.KubeConf.Arg.Storage
	if storageConf.StorageBucket == "" {
		return fmt.Errorf("missing storage bucket")
	}
	if storageConf.StorageAccessKey == "" {
		return fmt.Errorf("missing storage access key")
	}
	if storageConf.StorageSecretKey == "" {
		return fmt.Errorf("missing storage secret key")
	}
	return nil
}

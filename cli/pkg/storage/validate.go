package storage

import (
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/task"
	"fmt"
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
		return fmt.Errorf("missing storage bucket, please set it in env %s", common.ENV_S3_BUCKET)
	}
	if storageConf.StorageAccessKey == "" {
		return fmt.Errorf("missing storage access key, please set it in env %s", common.ENV_AWS_ACCESS_KEY_ID_SETUP)
	}
	if storageConf.StorageSecretKey == "" {
		return fmt.Errorf("missing storage secret key, please set it in env %s", common.ENV_AWS_SECRET_ACCESS_KEY_SETUP)
	}
	return nil
}

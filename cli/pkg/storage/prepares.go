package storage

import (
	"fmt"

	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/utils"
)

type CheckEtcdSSL struct {
	common.KubePrepare
}

func (p *CheckEtcdSSL) PreCheck(runtime connector.Runtime) (bool, error) {
	var files = []string{
		"/etc/ssl/etcd/ssl/ca.pem",
		fmt.Sprintf("/etc/ssl/etcd/ssl/node-%s-key.pem", runtime.RemoteHost().GetName()),
		fmt.Sprintf("/etc/ssl/etcd/ssl/node-%s.pem", runtime.RemoteHost().GetName()),
	}
	for _, f := range files {
		if !utils.IsExist(f) {
			return false, nil
		}
	}
	return true, nil
}

type CheckStorageType struct {
	common.KubePrepare
	StorageType string
}

func (p *CheckStorageType) PreCheck(runtime connector.Runtime) (bool, error) {
	storageType := p.KubeConf.Arg.Storage.StorageType
	if storageType == "" || storageType != p.StorageType {
		return false, nil
	}
	return true, nil
}

type CheckStorageVendor struct {
	common.KubePrepare
}

func (p *CheckStorageVendor) PreCheck(runtime connector.Runtime) (bool, error) {
	var storageType = p.KubeConf.Arg.Storage.StorageType
	var storageBucket = p.KubeConf.Arg.Storage.StorageBucket
	if storageType != common.OSS && storageType != common.COS && storageType != common.S3 {
		return false, nil
	}

	if storageBucket == "" {
		return false, nil
	}

	return true, nil
}

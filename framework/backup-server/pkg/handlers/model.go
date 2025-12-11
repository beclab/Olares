package handlers

import (
	"fmt"
	"net/url"
	"strings"

	"olares.com/backup-server/pkg/apiserver/response"
	"olares.com/backup-server/pkg/constant"
	"olares.com/backup-server/pkg/util"
)

type SnapshotNotifyState struct {
	Prepare  bool `json:"prepare"`
	Progress bool `json:"progress"`
	Result   bool `json:"result"`
}

type RestoreType struct {
	Name                string                  `json:"name"` // only for app backup
	Owner               string                  `json:"owner"`
	Type                string                  `json:"type"` // snapshot or url
	Path                string                  `json:"path"` // restore target path
	SubPath             string                  `json:"subPath"`
	SubPathTimestamp    int64                   `json:"subPathTimestamp"`
	BackupId            string                  `json:"backupId"`
	BackupName          string                  `json:"backupName"`
	BackupPath          string                  `json:"backupPath"` // from backupUrl
	Password            string                  `json:"p"`
	SnapshotId          string                  `json:"snapshotId"`
	SnapshotTime        string                  `json:"snapshotTime"` // from backupUrl
	ResticSnapshotId    string                  `json:"resticSnapshotId"`
	ClusterId           string                  `json:"clusterId"`
	Location            string                  `json:"location"`
	Endpoint            string                  `json:"endpoint"`
	TotalBytesProcessed int64                   `json:"totalBytesProcessed"`
	BackupUrl           *RestoreBackupUrlDetail `json:"backupUrl"`
}

type RestoreBackupUrlDetail struct {
	CloudName      string `json:"cloudName"` // awss3 tencentcloud filesystem
	RegionId       string `json:"regionId"`
	Bucket         string `json:"bucket"`
	Prefix         string `json:"prefix"`
	OlaresSuffix   string `json:"suffix"` // only used for space, prev backup did-suffix
	FilesystemPath string `json:"fsPath"` // only used for filesystem
}

type BackupAppResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type BackupAppStatusData struct {
	Status     string   `json:"status"`
	Data       any      `json:"data"`
	EntryFiles []string `json:"entry_files"`
	PgFiles    []string `json:"pg_files"`
}

type RestoreAppStatusData struct {
	BackupAppResponse
	Status string                    `json:"status"`
	Data   *RestoreAppStatusDataBody `json:"data"`
}

type RestoreAppStatusDataBody struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type BackupAppStatus struct {
	BackupAppResponse
	Data *BackupAppStatusData `json:"data"`
}

type proxyRequest struct {
	Op       string      `json:"op"`
	DataType string      `json:"datatype"`
	Version  string      `json:"version"`
	Group    string      `json:"group"`
	Param    interface{} `json:"param,omitempty"`
	Data     string      `json:"data,omitempty"`
	Token    string
}

type passwordResponse struct {
	response.Header
	Data *passwordResponseData `json:"data,omitempty"`
}

type passwordResponseData struct {
	Env   string `json:"env"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

type BackupUrlType struct {
	Schema         string     `json:"schema"`
	Host           string     `json:"host"`
	Path           string     `json:"path"`
	Values         url.Values `json:"values"`
	Location       string     `json:"location"`
	Endpoint       string     `json:"endpoint"`
	BackupId       string     `json:"backup_id"`
	BackupName     string     `json:"backup_name"`
	PvcPath        string     `json:"pvc_path"`
	CloudName      string     `json:"cloud_name"`
	Region         string     `json:"region"`
	Bucket         string     `json:"bucket"`
	Prefix         string     `json:"prefix"`
	OlaresSuffix   string     `json:"suffix"`
	FilesystemPath string     `json:"fs_path"`
}

func (u *BackupUrlType) GetStorage() (*RestoreBackupUrlDetail, error) {
	var detail = &RestoreBackupUrlDetail{
		CloudName:    u.CloudName,
		RegionId:     u.Region,
		Bucket:       u.Bucket,
		Prefix:       u.Prefix,
		OlaresSuffix: u.OlaresSuffix,
	}

	if u.Location == constant.BackupLocationFileSystem.String() {
		detail.FilesystemPath = u.Endpoint
	}

	return detail, nil
}

func (u *BackupUrlType) getBucketAndPrefix() (bucket string, prefix string, suffix string, err error) {
	path := strings.Trim(u.Path, "/")
	paths := strings.Split(path, "/")

	if u.Location == constant.BackupLocationSpace.String() {
		if len(paths) != 4 {
			err = fmt.Errorf("path invalid, path: %s", u.Path)
			return
		}
		suffix, err = util.GetSuffix(paths[1], "-")
		if err != nil {
			return
		}
		bucket = paths[0]
		prefix = paths[1]
		return
	} else {
		if len(paths) < 2 {
			err = fmt.Errorf("path invalid, path: %s", u.Path)
			return
		}
		bucket = paths[0]
		prefix = strings.Join(paths[1:len(paths)-1], "/")
		return
	}
}

func (u *BackupUrlType) getCloudName() string {
	var cloudName string
	if u.Location == constant.BackupLocationSpace.String() {
		if strings.Contains(u.Host, constant.LocationTypeAwsS3Tag) {
			cloudName = "aws"
		} else if strings.Contains(u.Host, constant.LocationTypeTencentCloudTag) {
			cloudName = "tencentcloud"
		}
	}

	return cloudName
}

func (u *BackupUrlType) getRegionId() (string, error) {
	var h = strings.Split(u.Host, ".")
	if len(h) != 4 {
		return "", fmt.Errorf("region invalid, host: %s", u.Host)
	}

	return h[1], nil
}

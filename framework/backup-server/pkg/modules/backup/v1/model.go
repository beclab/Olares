package v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"k8s.io/klog/v2"
	sysv1 "olares.com/backup-server/pkg/apis/sys.bytetrade.io/v1"
	"olares.com/backup-server/pkg/constant"
	"olares.com/backup-server/pkg/handlers"
	"olares.com/backup-server/pkg/util"
	stringx "olares.com/backup-server/pkg/util/string"
)

type LocationConfig struct {
	Name      string `json:"name"`                // olaresId or integrationCloudAccessKey
	CloudName string `json:"cloudName,omitempty"` // olares space cloudName
	RegionId  string `json:"regionId,omitempty"`  // olares space regionId
	Path      string `json:"path"`                // filesystem target path
}

type BackupCreate struct {
	Name             string              `json:"name"`
	Path             string              `json:"path"`
	Location         string              `json:"location"` // space or s3
	BackupCreateType *BackupCreateType   `json:"backupType"`
	LocationConfig   *LocationConfig     `json:"locationConfig,omitempty"`
	BackupPolicies   *sysv1.BackupPolicy `json:"backupPolicy,omitempty"`
	Password         string              `json:"password,omitempty"`
	ConfirmPassword  string              `json:"confirmPassword,omitempty"`
}

type BackupCreateType struct {
	Type string `json:"type"` // app
	Name string `json:"name"` // app name, like wise etc
}

func (b *BackupCreate) verify() error {
	if len(b.Name) == 0 {
		return fmt.Errorf("backup name is empty")
	}

	if b.BackupCreateType != nil && b.BackupCreateType.Type == constant.BackupTypeApp {
		return b.verifyBackupApp()
	}

	return b.verifyBackupFiles()
}

func (b *BackupCreate) verifyBackupApp() error {
	if len(b.BackupCreateType.Name) == 0 {
		return fmt.Errorf("app name is empty")
	}

	return nil
}

func (b *BackupCreate) verifyBackupFiles() error {
	if stringx.IsOnlyWhitespace(b.Name) {
		return fmt.Errorf("backup name cannot consist of only spaces")
	}

	if stringx.IsReservedName(b.Name) {
		return fmt.Errorf("backup name cannot be a reserved name")
	}

	if strings.HasPrefix(b.Name, ".") {
		return fmt.Errorf("backup name must not start with '.'")
	}

	if strings.ContainsRune(b.Name, '\x00') {
		return fmt.Errorf("backup name cannot contain null character")
	}

	if ok, r := stringx.ContainsIllegalChars(b.Name); ok {
		return fmt.Errorf("backup name cannot contain illegal characters: %c", r)
	}

	if b.Path == "" {
		return errors.New("backup path is required")
	}

	return nil
}

type BackupEnabled struct {
	Event string `json:"event"`
}

type Snapshot struct {
	Name string `json:"name,omitempty"`

	CreationTimestamp   int64  `json:"creationTimestamp,omitempty"`
	NextBackupTimestamp *int64 `json:"nextBackupTimestamp,omitempty"`

	Size *int64 `json:"size,omitempty"`

	Phase *string `json:"phase"`

	FailedMessage string `json:"failedMessage,omitempty"`
}

type SnapshotCancel struct {
	Event string `json:"event"`
}

type CreateSnapshot struct {
	Event string `json:"event"`
}

type RestoreCreate struct {
	BackupUrl  string `json:"backupUrl"`
	Password   string `json:"password"`
	SnapshotId string `json:"snapshotId"`
	Path       string `json:"path"`
	SubPath    string `json:"dir"`
}

type RestoreCheckBackupUrl struct {
	BackupUrl string `json:"backupUrl"`
	Password  string `json:"password"`
	Limit     int64  `json:"limit"`
	Offset    int64  `json:"offset"`
}

func (r *RestoreCreate) verify() error {
	var backupUrl = strings.TrimSpace(r.BackupUrl)
	var snapshotId = strings.TrimSpace(r.SnapshotId)
	var password = strings.TrimSpace(r.Password)

	if (backupUrl == "") == (snapshotId == "") {
		return fmt.Errorf("snapshotId is required")
	}

	if backupUrl != "" && password == "" {
		return fmt.Errorf("password is required")
	}

	return nil
}

type RestoreCancel struct {
	Event string `json:"event"`
}

type ResponseBackupListInfo struct {
	TotalPage int64                 `json:"totalPage"`
	Backups   []*ResponseBackupList `json:"backups"`
}

type ResponseBackupList struct {
	Id                  string `json:"id"`
	Name                string `json:"name"`
	BackupType          string `json:"backupType"`
	BackupAppTypeName   string `json:"backupAppTypeName,omitempty"`
	SnapshotFrequency   string `json:"snapshotFrequency"`
	Location            string `json:"location"`           // space, awss3, tencentcloud ...
	LocationConfigName  string `json:"locationConfigName"` // olaresDid / cloudAccessKey
	SnapshotId          string `json:"snapshotId"`
	NextBackupTimestamp *int64 `json:"nextBackupTimestamp,omitempty"`
	CreateAt            int64  `json:"createAt"`
	Status              string `json:"status"`
	Size                string `json:"size"`
	RestoreSize         string `json:"restoreSize"`
	Path                string `json:"path"`
}

type ResponseBackupDetail struct {
	Id                string              `json:"id"`
	Name              string              `json:"name"`
	BackupType        string              `json:"backupType"`
	BackupAppTypeName string              `json:"backupAppTypeName,omitempty"`
	Path              string              `json:"path"`
	BackupPolicies    *sysv1.BackupPolicy `json:"backupPolicies,omitempty"`
	Size              string              `json:"size"`
	RestoreSize       string              `json:"restoreSize"`
}

type ResponseSnapshotList struct {
	Id       string `json:"id"`
	CreateAt int64  `json:"createAt"`
	Size     string `json:"size"`
	Progress int    `json:"progress"`
	Status   string `json:"status"`
}

type ResponseSnapshotDetail struct {
	Id           string `json:"id"`
	Size         string `json:"size"`
	SnapshotType string `json:"snapshotType"`
	Progress     int    `json:"progress"`
	Status       string `json:"status"`
	Message      string `json:"message"`
}

type ResponseRestoreDetail struct {
	BackupName        string `json:"name"`
	BackupPath        string `json:"backupPath,omitempty"`
	BackupType        string `json:"backupType"`
	BackupAppTypeName string `json:"backupAppTypeName,omitempty"`
	SnapshotTime      int64  `json:"snapshotTime"`
	RestorePath       string `json:"restorePath,omitempty"`
	Progress          int    `json:"progress,omitempty"`
	Status            string `json:"status"`
	EndAt             int64  `json:"endAt,omitempty"`
	Message           string `json:"message"`
}

type ResponseRestoreList struct {
	Id                string  `json:"id"`
	BackupName        string  `json:"name"`
	BackupType        string  `json:"backupType"`
	BackupAppTypeName string  `json:"backupAppTypeName,omitempty"`
	Path              string  `json:"path"`
	CreateAt          int64   `json:"createAt"`     // restore createAt
	SnapshotTime      int64   `json:"snapshotTime"` // snapshotTime createAt
	EndAt             int64   `json:"endAt"`
	Status            string  `json:"status"`
	Progress          float64 `json:"progress,omitempty"`
}

type SnapshotDetails struct {
	Name string `json:"name"`

	CreationTimestamp int64 `json:"creationTimestamp"`

	Size *int64 `json:"size"`

	Phase *string `json:"phase"`

	// Config *Config `json:"config,omitempty"`

	FailedMessage string `json:"failedMessage"`

	Owner *string `json:"owner"`

	BackupType string `json:"backupType"`

	SnapshotId string `json:"snapshotId"`

	RepositoryPasswordHash string `json:"repositoryPasswordHash"`

	BackupConfigName string `json:"backupConfigName"`
}

type ListBackupsDetails struct {
	Name string `json:"name"`

	Size *int64 `json:"size,omitempty"`

	SnapshotName string `json:"snapshotName"`

	SnapshotFrequency string `json:"snapshotFrequency"`

	CreationTimestamp int64 `json:"creationTimestamp,omitempty"`

	NextBackupTimestamp *int64 `json:"nextBackupTimestamp,omitempty"`

	Phase string `json:"phase,omitempty"`

	FailedMessage string `json:"failedMessage,omitempty"`
}

type ResponseDescribeBackup struct {
	Name string `json:"name"`

	Path string `json:"path"`

	Size *uint64 `json:"totalSize,omitempty"`

	BackupPolicies *sysv1.BackupPolicy `json:"backupPolicies"`

	Snapshots []Snapshot `json:"snapshots,omitempty"`

	// new list api fields
	SnapshotName string `json:"latestSnapshotName"`

	CreationTimestamp int64 `json:"creationTimestamp,omitempty"`

	NextBackupTimestamp *int64 `json:"nextBackupTimestamp,omitempty"`

	Phase string `json:"phase,omitempty"`

	FailedMessage string `json:"failedMessage,omitempty"`
}

type PostBackupSchedule struct {
	Schedule string `json:"schedule"`

	Paused bool `json:"paused"`
}

type SyncBackup struct {
	UID string `json:"uid"`

	CreationTimestamp int64 `json:"creationTimestamp"`

	Name string `json:"name"`

	Namespace string `json:"namespace,omitempty"`

	BackupConfigName string `json:"backupConfigName"`

	Size *int64 `json:"size"`

	// S3Config *Config `json:"s3Config"`

	Phase *string `json:"phase"`

	FailedMessage string `json:"failedMessage"`

	Owner *string `json:"owner"`

	OlaresVersion *string `json:"olaresVersion"`

	Expiration *int64 `json:"expiration"`

	CompletionTimestamp *int64 `json:"completionTimestamp"`

	BackupType string `json:"backupType"`

	SnapshotId string `json:"snapshotId"`

	S3Repository string `json:"s3Repository"`

	RepositoryPasswordHash string `json:"repositoryPasswordHash"`

	RefFullyBackupUid *string `json:"-"`

	RefFullyBackupName *string `json:"-"`
}

type SyncBackupList []*SyncBackup

func parseResponseSnapshotList(snapshots *sysv1.SnapshotList, currentPage, totalPage, totalCount int64) map[string]interface{} {
	_ = currentPage
	var data = make(map[string]interface{})

	if snapshots == nil || len(snapshots.Items) == 0 {
		data["totalPage"] = 1
		data["totalCount"] = 0
		data["snapshots"] = []struct{}{}
		return data
	}

	var ss []*ResponseSnapshotList
	for _, snapshot := range snapshots.Items {
		var item = &ResponseSnapshotList{
			Id:       snapshot.Name,
			CreateAt: snapshot.Spec.CreateAt.Unix(),
			Size:     handlers.ParseSnapshotSize(snapshot.Spec.Size),
			Progress: snapshot.Spec.Progress,
			Status:   parseMessage(snapshot.Spec.Phase),
		}
		ss = append(ss, item)
	}

	data["snapshots"] = ss
	data["totalPage"] = totalPage
	data["totalCount"] = totalCount

	return data
}

func parseResponseSnapshotOne(snapshot *sysv1.Snapshot) map[string]interface{} {
	var data = make(map[string]interface{})

	data["id"] = snapshot.Name
	data["createAt"] = snapshot.Spec.CreateAt.Unix()
	data["size"] = handlers.ParseSnapshotSize(snapshot.Spec.Size)
	data["progress"] = snapshot.Spec.Progress
	data["status"] = parseMessage(snapshot.Spec.Phase)

	return data
}

func parseResponseSnapshotDetail(snapshot *sysv1.Snapshot) *ResponseSnapshotDetail {
	return &ResponseSnapshotDetail{
		Id:           snapshot.Name,
		Size:         handlers.ParseSnapshotSize(snapshot.Spec.Size),
		SnapshotType: handlers.ParseSnapshotTypeTitle(snapshot.Spec.SnapshotType),
		Progress:     snapshot.Spec.Progress,
		Status:       parseMessage(snapshot.Spec.Phase),
		Message:      parseMessage(snapshot.Spec.Message),
	}
}

func parseResponseBackupDetail(backup *sysv1.Backup) *ResponseBackupDetail {
	var backupType = handlers.GetBackupType(backup)
	var backupAppTypeName string
	if backupType == constant.BackupTypeApp {
		backupAppTypeName = handlers.GetBackupAppName(backup)
	}
	return &ResponseBackupDetail{
		Id:                backup.Name,
		Name:              backup.Spec.Name,
		BackupType:        backupType,
		BackupAppTypeName: backupAppTypeName,
		BackupPolicies:    backup.Spec.BackupPolicy,
		Path:              handlers.ParseBackupTypePath(backup.Spec.BackupType),
		Size:              handlers.ParseSnapshotSize(backup.Spec.Size),
		RestoreSize:       handlers.ParseSnapshotSize(backup.Spec.RestoreSize),
	}
}

func parseResponseBackupCreate(backup *sysv1.Backup) map[string]interface{} {
	var data = make(map[string]interface{})

	locationConfig, err := handlers.GetBackupLocationConfig(backup)
	if err != nil {
		return data
	}

	var location = locationConfig["location"]
	var locationConfigName = locationConfig["name"]
	if location == constant.BackupLocationFileSystem.String() {
		locationConfigName = locationConfig["path"]
	}

	var nextBackupTimestamp = handlers.GetNextBackupTime(*backup.Spec.BackupPolicy)
	var backupType = handlers.GetBackupType(backup)
	if backupType == constant.BackupTypeApp {
		var backupAppName = handlers.GetBackupAppName(backup)
		data["backupAppTypeName"] = backupAppName
	}

	data["id"] = backup.Name
	data["name"] = backup.Spec.Name
	data["backupType"] = backupType
	data["nextBackupTimestamp"] = *nextBackupTimestamp
	data["location"] = location
	data["locationConfigName"] = locationConfigName
	data["createAt"] = backup.Spec.CreateAt.Unix()
	data["size"] = "0"
	data["path"] = handlers.ParseBackupTypePath(backup.Spec.BackupType)
	data["status"] = constant.Pending.String()

	return data
}

func parseResponseBackupOne(backup *sysv1.Backup, snapshot *sysv1.Snapshot) (map[string]interface{}, error) {
	var result = make(map[string]interface{})

	locationConfig, err := handlers.GetBackupLocationConfig(backup)
	if err != nil {
		return nil, err
	}

	backupType := handlers.GetBackupType(backup)
	var backupAppTypeName string
	if backupType == constant.BackupTypeApp {
		backupAppTypeName = handlers.GetBackupAppName(backup)
		result["backupAppTypeName"] = backupAppTypeName
	}
	location := locationConfig["location"]
	locationConfigName := locationConfig["name"]
	if location == constant.BackupLocationFileSystem.String() {
		locationConfigName = locationConfig["path"]
	}

	result["id"] = backup.Name
	result["name"] = backup.Spec.Name
	result["backupType"] = backupType
	result["nextBackupTimestamp"] = handlers.GetNextBackupTime(*backup.Spec.BackupPolicy)
	result["location"] = location
	result["locationConfigName"] = locationConfigName
	result["size"] = fmt.Sprintf("%d", *backup.Spec.Size)
	result["createAt"] = backup.Spec.CreateAt.Unix()
	result["path"] = handlers.ParseBackupTypePath(backup.Spec.BackupType)
	if snapshot != nil {
		result["status"] = *snapshot.Spec.Phase
	} else {
		result["status"] = constant.Pending.String()
	}

	return result, nil
}

func parseResponseBackupList(data *sysv1.BackupList, snapshots *sysv1.SnapshotList, totalPage int64) *ResponseBackupListInfo {
	var result = new(ResponseBackupListInfo)
	if data == nil || data.Items == nil || len(data.Items) == 0 {
		result.TotalPage = 0
		result.Backups = []*ResponseBackupList{}
		return result
	}

	var bs = make(map[string]*sysv1.Snapshot)
	var res []*ResponseBackupList
	if snapshots != nil {
		for _, snapshot := range snapshots.Items {
			if *snapshot.Spec.Phase != constant.Completed.String() {
				continue
			}
			if _, ok := bs[snapshot.Spec.BackupId]; !ok {
				bs[snapshot.Spec.BackupId] = &snapshot
				continue
			}
		}
	}

	for _, backup := range data.Items {
		locationConfig, err := handlers.GetBackupLocationConfig(&backup)
		if err != nil {
			continue
		}

		location := locationConfig["location"]
		locationConfigName := locationConfig["name"]
		backupType := handlers.GetBackupType(&backup)
		if location == constant.BackupLocationFileSystem.String() {
			locationConfigName = locationConfig["path"]
		}
		var backupAppTypeName string
		if backupType == constant.BackupTypeApp {
			backupAppTypeName = handlers.GetBackupAppName(&backup)
		}
		var r = &ResponseBackupList{
			Id:                  backup.Name,
			Name:                backup.Spec.Name,
			BackupType:          backupType,
			BackupAppTypeName:   backupAppTypeName,
			SnapshotFrequency:   handlers.ParseBackupSnapshotFrequency(backup.Spec.BackupPolicy.SnapshotFrequency),
			NextBackupTimestamp: handlers.GetNextBackupTime(*backup.Spec.BackupPolicy),
			CreateAt:            backup.Spec.CreateAt.Unix(),
			Location:            location,
			LocationConfigName:  locationConfigName, // filesystem is target
			Path:                handlers.ParseBackupTypePath(backup.Spec.BackupType),
			RestoreSize:         handlers.ParseSnapshotSize(backup.Spec.RestoreSize),
		}

		if s, ok := bs[backup.Name]; ok {
			r.SnapshotId = s.Name
			r.Size = handlers.ParseSnapshotSize(s.Spec.Size)
			r.Status = *s.Spec.Phase
		} else {
			r.Size = "0"
			r.Status = constant.Pending.String()
		}

		res = append(res, r)
	}

	result.TotalPage = totalPage
	result.Backups = res

	return result
}

func parseResponseRestoreDetailFromBackupUrl(restore *sysv1.Restore) (*ResponseRestoreDetail, error) {
	var result = &ResponseRestoreDetail{}
	var backupType, _ = handlers.GetRestoreType(restore)

	var restoreType = restore.Spec.RestoreType[backupType]

	var typeMap map[string]interface{}
	if err := json.Unmarshal([]byte(restoreType), &typeMap); err != nil {
		return result, err
	}

	var backupName, backupPath, snapshotTime, restorePath, subPath, backupAppTypeName string
	if backupType == constant.BackupTypeApp {
		backupAppTypeName = handlers.GetRestoreAppName(restore)
	}

	_, ok := typeMap["backupName"]
	if ok {
		backupName = typeMap["backupName"].(string)
	}

	_, ok = typeMap["backupPath"]
	if ok {
		backupPath = typeMap["backupPath"].(string)
	}

	_, ok = typeMap["snapshotTime"]
	if ok {
		snapshotTime = typeMap["snapshotTime"].(string)
	}

	_, ok = typeMap["path"]
	if ok {
		restorePath = typeMap["path"].(string)
	}

	_, ok = typeMap["subPath"]
	if ok {
		subPath = typeMap["subPath"].(string)
	}

	result = &ResponseRestoreDetail{
		BackupName:        backupName,
		BackupPath:        backupPath,
		BackupType:        backupType,
		BackupAppTypeName: backupAppTypeName,
		SnapshotTime:      util.ParseToInt64(snapshotTime),
		RestorePath:       filepath.Join(restorePath, subPath),
		Progress:          restore.Spec.Progress,
		Status:            *restore.Spec.Phase,
	}

	if restore.Spec.EndAt != nil {
		result.EndAt = restore.Spec.EndAt.Unix()
	}

	if restore.Spec.Message != nil {
		result.Message = *restore.Spec.Message
	}

	return result, nil
}

func parseResponseRestoreDetail(backup *sysv1.Backup, snapshot *sysv1.Snapshot, restore *sysv1.Restore) *ResponseRestoreDetail {
	var res = &ResponseRestoreDetail{
		BackupName:   backup.Spec.Name,
		BackupPath:   handlers.GetBackupPath(backup),
		SnapshotTime: snapshot.Spec.CreateAt.Unix(),
		RestorePath:  handlers.GetRestorePath(restore),
		Progress:     restore.Spec.Progress,
		Status:       *restore.Spec.Phase,
	}

	if restore.Spec.Message != nil {
		res.Message = *restore.Spec.Message
	}
	return res
}

func parseResponseRestoreCreate(restore *sysv1.Restore, backupName, snapshotTime, restorePath string) map[string]interface{} {
	var data = make(map[string]interface{})
	var backupType, _ = handlers.GetRestoreType(restore)
	var backupAppTypeName string
	if backupType == constant.BackupTypeApp {
		backupAppTypeName = handlers.GetRestoreAppName(restore)
		data["backupAppTypeName"] = backupAppTypeName
	}

	data["id"] = restore.Name
	data["name"] = backupName
	data["backupType"] = backupType
	data["path"] = restorePath
	data["createAt"] = restore.Spec.CreateAt.Unix()
	data["snapshotTime"] = time.Unix(util.ParseToInt64(snapshotTime), 0).Unix()
	data["endAt"] = 0
	data["progress"] = 0
	data["status"] = constant.Pending.String()

	return data
}

func parseResponseRestoreOne(restore *sysv1.Restore, backupAppTypeName string, backupName string, snapshotTime string, restorePath, subPath string) map[string]interface{} {
	var res = make(map[string]interface{})

	res["id"] = restore.Name
	res["name"] = backupName
	res["path"] = filepath.Join(restorePath, subPath)
	res["createAt"] = restore.Spec.CreateAt.Unix()
	res["snapshotTime"] = time.Unix(util.ParseToInt64(snapshotTime), 0)
	res["progress"] = restore.Spec.Progress
	res["status"] = *restore.Spec.Phase

	if restore.Spec.EndAt != nil {
		res["endAt"] = restore.Spec.EndAt.Unix()
	}
	if backupAppTypeName != "" {
		res["backupAppTypeName"] = backupAppTypeName
	}

	return res
}

func parseCheckBackupUrl(snapshots *sysv1.SnapshotList, backupName, backupTypeTag, backupTypeAppName, location, userspacePath string, totalCount int64, totalPage int64) map[string]interface{} {
	var result = make(map[string]interface{})
	result["totalCount"] = totalCount
	result["totalPage"] = totalPage

	if snapshots == nil || len(snapshots.Items) == 0 {
		result["snapshots"] = []struct{}{}
		return result
	}

	var backupPathAbs = snapshots.Items[0].Spec.Location
	var backupPath = util.ReplacePathPrefix(backupPathAbs, userspacePath, constant.ExternalPath)

	var items []map[string]interface{}
	for _, snapshot := range snapshots.Items {
		var item = make(map[string]interface{})
		item["id"] = snapshot.Spec.SnapshotId
		item["createAt"] = snapshot.Spec.CreateAt.Unix()
		item["size"] = fmt.Sprintf("%d", *snapshot.Spec.Size)
		items = append(items, item)
	}

	result["type"] = location
	result["backupName"] = backupName
	result["backupType"] = backupTypeTag
	result["backupPath"] = backupPath
	result["snapshots"] = items

	if backupTypeTag == constant.BackupTypeApp {
		result["backupTypeAppName"] = backupTypeAppName
	}

	return result
}

func parseResponseRestoreList(data *sysv1.RestoreList, totalPage int64) map[string]interface{} {
	var res = make(map[string]interface{})

	if data == nil || data.Items == nil || len(data.Items) == 0 {
		res["totalPage"] = totalPage
		res["restores"] = []struct{}{}
		return res
	}

	var result []*ResponseRestoreList
	for _, restore := range data.Items {
		backupType, err := handlers.GetRestoreType(&restore)
		if err != nil {
			continue
		}
		var backupAppTypeName string
		if backupType == constant.BackupTypeApp {
			backupAppTypeName = handlers.GetRestoreAppName(&restore)
		}
		d, err := handlers.ParseRestoreType(backupType, &restore)
		if err != nil {
			continue
		}

		snapshotTime := time.Unix(util.ParseToInt64(d.SnapshotTime), 0)
		var r = &ResponseRestoreList{
			Id:                restore.Name,
			BackupName:        d.BackupName,
			BackupType:        backupType,
			BackupAppTypeName: backupAppTypeName,
			Path:              filepath.Join(d.Path, d.SubPath),
			CreateAt:          restore.Spec.CreateAt.Unix(),
			SnapshotTime:      snapshotTime.Unix(),
			Status:            *restore.Spec.Phase,
		}

		if restore.Spec.EndAt != nil {
			r.EndAt = restore.Spec.EndAt.Unix()
		}

		result = append(result, r)
	}

	res["totalPage"] = totalPage
	res["restores"] = result

	return res
}

func (s *SyncBackup) FormData() (map[string]string, error) {
	// s3config, err := json.Marshal(s.S3Config)
	// if err != nil {
	// 	klog.Error("parse s3 config error, ", err)
	// 	return nil, err
	// }

	formdata := make(map[string]string)
	formdata["backupConfigName"] = s.BackupConfigName
	formdata["completionTimestamp"] = toString(s.CompletionTimestamp)
	formdata["creationTimestamp"] = toString(s.CreationTimestamp)
	formdata["expiration"] = toString(s.Expiration)
	formdata["name"] = toString(s.Name)
	formdata["phase"] = toString(s.Phase)
	formdata["uid"] = toString(s.UID)
	formdata["size"] = toString(s.Size)
	// formdata["s3Config"] = string(s3config)
	formdata["olaresVersion"] = toString(s.OlaresVersion)
	formdata["owner"] = toString(s.Owner)
	formdata["backupType"] = toString(s.BackupType)
	formdata["s3Repository"] = toString(s.S3Repository)
	formdata["snapshotId"] = toString(s.SnapshotId)
	return formdata, nil
}

func toString(v interface{}) string {
	int64ToStr := func(n int64) string {
		s := strconv.FormatInt(n, 10)
		return s
	}

	switch s := v.(type) {
	case int64:
		return int64ToStr(s)
	case *int64:
		if s == nil {
			return ""
		}
		return int64ToStr(*s)
	case *string:
		if s == nil {
			return ""
		}
		return *s
	case string:
		return s
	}

	klog.Error("unknown field type")
	return ""
}

func parseMessage(msg *string) string {
	if msg == nil {
		return ""
	}
	return *msg
}

func isBackupApp(backupType string) bool {
	return backupType == constant.BackupTypeApp
}

func getBackupType(bc *BackupCreate) string {
	if bc.BackupCreateType != nil && bc.BackupCreateType.Type == constant.BackupTypeApp {
		return constant.BackupTypeApp
	}
	return constant.BackupTypeFile
}

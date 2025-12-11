package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"olares.com/backup-server/pkg/constant"
	"olares.com/backup-server/pkg/util/log"
)

const (
	BackupAppHost = "rss-svc.knowledge-shared:3010"
)

const (
	BackupAppStartPath   = "{app}/backup/start"
	BackupAppStatusPath  = "{app}/backup/status"
	BackupAppResultPath  = "{app}/backup/result"
	RestoreAppStartPath  = "{app}/backup/restore"
	RestoreAppStatusPath = "{app}/backup/restore_status"
)

var appNameMap = map[string]string{
	"wise": "knowledge",
}

type AppHandler struct {
	name  string
	owner string
}

func NewAppHandler(appName, owner string) *AppHandler {
	return &AppHandler{
		name:  appName,
		owner: owner,
	}
}

func (app *AppHandler) StartAppBackup(parentCtx context.Context, backupId, snapshotId string) error {
	var ctx, cancel = context.WithTimeout(parentCtx, 15*time.Second)
	defer cancel()
	var appUrl = strings.ReplaceAll(BackupAppStartPath, "{app}", getAppUrlName(app.name))
	var url = fmt.Sprintf("http://%s/%s", BackupAppHost, appUrl)
	var headers = map[string]string{
		"X-BFL-USER": app.owner,
	}

	var result *BackupAppResponse

	client := resty.New().SetTimeout(15 * time.Second).SetDebug(true)

	data := map[string]string{
		"backup_id":   backupId,
		"snapshot_id": snapshotId,
	}

	resp, err := client.R().
		SetContext(ctx).
		SetHeaders(headers).
		SetFormData(data).
		SetResult(&result).
		Put(url)

	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("start app %s backup error, status code: %d, msg: %s", app.name, resp.StatusCode(), string(resp.Body()))
	}

	if result.Code != 0 {
		return fmt.Errorf("start app %s backup error, code: %d, msg: %s", app.name, result.Code, result.Message)
	}

	log.Infof("start app %s backup, backupId: %s, snapshotId: %s, msg: %s", app.name, backupId, snapshotId, string(resp.Body()))

	return nil
}

func (app *AppHandler) GetAppBackupStatus(parentCtx context.Context, backupId, snapshotId string) (*BackupAppStatus, error) {

	var result *BackupAppStatus
	var appUrl = strings.ReplaceAll(BackupAppStatusPath, "{app}", getAppUrlName(app.name))
	var url = fmt.Sprintf("http://%s/%s?backup_id=%s&snapshot_id=%s", BackupAppHost, appUrl, backupId, snapshotId)
	var headers = map[string]string{
		"X-BFL-USER": app.owner,
	}

	client := resty.New().SetTimeout(15 * time.Second).SetDebug(true)

	resp, err := client.R().
		SetContext(parentCtx).
		SetHeaders(headers).
		SetResult(&result).
		Get(url)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("get app %s backup status error, status code: %d, msg: %s", app.name, resp.StatusCode(), string(resp.Body()))
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("get app %s backup status error, code: %d, msg: %s", app.name, result.Code, result.Message)
	}

	log.Infof("get app %s backup status, backupId: %s, snapshotId: %s, msg: %s", app.name, backupId, snapshotId, string(resp.Body()))

	return result, nil
}

func (app *AppHandler) SendBackupResult(parentCtx context.Context, backupId, snapshotId string, errorMessage error) error {
	var ctx, cancel = context.WithTimeout(parentCtx, 15*time.Second)
	defer cancel()

	var result *BackupAppResponse
	var status = constant.Completed.String()
	var msg string

	if errorMessage != nil {
		status = constant.Failed.String()
		msg = errorMessage.Error()
	}

	var appUrl = strings.ReplaceAll(BackupAppResultPath, "{app}", getAppUrlName(app.name))
	var url = fmt.Sprintf("http://%s/%s", BackupAppHost, appUrl)
	var headers = map[string]string{
		"X-BFL-USER": app.owner,
	}
	var data = map[string]string{
		"backup_id":   backupId,
		"snapshot_id": snapshotId,
		"status":      status,
		"message":     msg,
	}

	client := resty.New().SetTimeout(15 * time.Second).SetDebug(true)
	resp, err := client.R().
		SetContext(ctx).
		SetHeaders(headers).
		SetFormData(data).
		SetResult(&result).
		Put(url)

	if err != nil {
		log.Warnf("send app %s backup result, backupId: %s, snapshotId: %s, error: %v", app.name, backupId, snapshotId, err)
		return nil
	}

	log.Infof("send app %s backup result, backupId: %s, snapshotId: %s, msg: %s", app.name, backupId, snapshotId, string(resp.Body()))

	return nil
}

func (app *AppHandler) StartAppRestore(parentCtx context.Context, restoreId string, resticSnapshotId string, metadata string) error {

	var ctx, cancel = context.WithTimeout(parentCtx, 15*time.Second)
	defer cancel()
	var appUrl = strings.ReplaceAll(RestoreAppStartPath, "{app}", getAppUrlName(app.name))
	var url = fmt.Sprintf("http://%s/%s", BackupAppHost, appUrl)
	var headers = map[string]string{
		"X-BFL-USER": app.owner,
	}

	var result *BackupAppResponse
	var md = make(map[string]interface{})
	if err := json.Unmarshal([]byte(metadata), &md); err != nil {
		return fmt.Errorf("start restore, json convert error: %v, metadata: %s", err, metadata)
	}

	client := resty.New().SetTimeout(15 * time.Second).SetDebug(true)

	data := map[string]interface{}{
		"restore_id":         restoreId,
		"restic_snapshot_id": resticSnapshotId,
		"data":               md,
	}

	resp, err := client.R().
		SetContext(ctx).
		SetHeaders(headers).
		SetBody(data).
		SetResult(&result).
		Put(url)

	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("start restore error, status code: %d, msg: %s", resp.StatusCode(), string(resp.Body()))
	}

	if result.Code != 0 {
		return fmt.Errorf("start restore error, code: %d, msg: %s", result.Code, result.Message)
	}

	log.Infof("start app %s restore, restoreId: %s, snapshotId: %s, msg: %s", app.name, restoreId, resticSnapshotId, string(resp.Body()))

	return nil
}

func (app *AppHandler) GetAppRestoreStatus(parentCtx context.Context, restoreId string) (*RestoreAppStatusData, error) {
	var ctx, cancel = context.WithTimeout(parentCtx, 15*time.Second)
	defer cancel()

	var result *RestoreAppStatusData
	var appUrl = strings.ReplaceAll(RestoreAppStatusPath, "{app}", getAppUrlName(app.name))
	var url = fmt.Sprintf("http://%s/%s?restore_id=%s", BackupAppHost, appUrl, restoreId)
	var headers = map[string]string{
		"X-BFL-USER": app.owner,
	}

	client := resty.New().SetTimeout(15 * time.Second).SetDebug(true)

	resp, err := client.R().
		SetContext(ctx).
		SetHeaders(headers).
		SetResult(&result).
		Get(url)

	if err != nil {
		return result, err
	}

	if resp.StatusCode() != http.StatusOK {
		log.Errorf("get app %s restore status error, status code: %d, msg: %s", app.name, resp.StatusCode(), string(resp.Body()))
		return result, fmt.Errorf("get app restore http status error: %s", string(resp.Body()))
	}

	if result.Code != 0 {
		log.Errorf("get app %s restore status error, code: %d, msg: %s", app.name, result.Code, result.Message)
		return result, fmt.Errorf("get app restore code error, code: %d", result.Code)
	}

	if result.Data == nil {
		log.Errorf("get app %s restore data is empty, data: %s", app.name, string(resp.Body()))
		return result, fmt.Errorf("get app restore data is empty")
	}

	log.Infof("get app %s restore status, restoreId: %s, msg: %s", app.name, restoreId, string(resp.Body()))

	return result, nil
}

func getAppUrlName(appName string) string {
	return appNameMap[appName]
}

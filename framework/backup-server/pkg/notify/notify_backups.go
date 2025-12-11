package notify

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/emicklei/go-restful/v3"
	"github.com/go-resty/resty/v2"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
	"olares.com/backup-server/pkg/constant"
	"olares.com/backup-server/pkg/util"
	httpx "olares.com/backup-server/pkg/util/http"
	"olares.com/backup-server/pkg/util/log"
)

const (
	SendBackupUrl       = "/v1/resource/backup/save"
	SendSnapshotUrl     = "/v1/resource/snapshot/save"
	SendStopBackupUrl   = "/v1/resource/backup/stop"
	CheckBackupUsageUrl = "/v1/resource/backup/usage"
)

type Backup struct {
	UserId         string `json:"user_id"` // did
	Token          string `json:"token"`   // access token
	BackupId       string `json:"backup_id"`
	Name           string `json:"name"`
	BackupType     string `json:"backup_type"`
	BackupTime     int64  `json:"backup_time"`
	BackupPath     string `json:"backup_path"`     // backup path
	BackupLocation string `json:"backup_location"` // location  space / awss3 / tencentcloud / ...
}

type Snapshot struct {
	UserId           string `json:"user_id"`     // did
	BackupId         string `json:"backup_id"`   //
	SnapshotId       string `json:"snapshot_id"` // restic snapshotId
	ResticSnapshotId string `json:"restic_snapshot_id"`
	Size             uint64 `json:"size"` // snapshot size
	BackupSize       uint64 `json:"backup_size"`
	Unit             string `json:"unit"`          // "byte"
	SnapshotTime     int64  `json:"snapshot_time"` // createAt
	Status           string `json:"status"`        // snapshot phase
	Type             string `json:"type"`          // fully / incremental
	Url              string `json:"url"`           // repo URL
	CloudName        string `json:"cloud_name"`    // awss3 / tencentcloud(space）； awss3 / tencentcloud / filesystem
	RegionId         string `json:"region_id"`     // regionId(space); extract from aws/cos URL
	Bucket           string `json:"bucket"`        // bucket(space); extract from aws/cos URL
	Prefix           string `json:"prefix"`        // prefix(space); extract from aws/cos URL
	Message          string `json:"message"`       // message
}

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Usage struct {
	Response
	Data *UsageData `json:"data"`
}

type UsageData struct {
	ToTalSize uint64 `json:"totalSize"`
	UsageSize uint64 `json:"usageSize"`
	CanBackup bool   `json:"canBackup"`
	PlanLevel int    `json:"planLevel"`
}

func NotifyBackup(ctx context.Context, cloudApiUrl string, backup *Backup) error {
	var backoff = wait.Backoff{
		Duration: 2 * time.Second,
		Factor:   2,
		Jitter:   0.5,
		Steps:    5,
	}

	var backupPath = backup.BackupPath

	if err := retry.OnError(backoff, func(err error) bool {
		return true
	}, func() error {
		var url = fmt.Sprintf("%s%s", cloudApiUrl, SendBackupUrl)
		var headers = make(map[string]string)
		headers[restful.HEADER_ContentType] = "application/json"

		var data = make(map[string]interface{})
		data["userid"] = backup.UserId
		data["token"] = backup.Token
		data["backupId"] = backup.BackupId
		data["name"] = backup.Name
		data["backupType"] = parseBackupTypeCode(backup.BackupType)
		data["backupTime"] = backup.BackupTime
		data["backupPath"] = backupPath
		data["backupLocation"] = backup.BackupLocation

		log.Infof("[notify] backup data: %s", util.ToJSON(data))

		var result *Response
		client := resty.New().SetTimeout(15 * time.Second).
			SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true}).
			SetDebug(true)

		resp, err := client.R().
			SetContext(ctx).
			SetHeaders(headers).
			SetBody(data).
			SetResult(&result).
			Post(url)

		if err != nil {
			return fmt.Errorf("[notify] send new backup request error: %v, url: %s", err, url)
		}

		if resp.StatusCode() != http.StatusOK {
			return fmt.Errorf("[notify] send new backup, http status error: %d, url: %s", resp.StatusCode(), url)
		}

		if result.Code != 200 && result.Code != 501 {
			return fmt.Errorf("[notify] send new backup record failed: %d, url: %s", result.Code, url)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func NotifySnapshot(ctx context.Context, cloudApiUrl string, snapshot *Snapshot) error {
	var backoff = wait.Backoff{
		Duration: 5 * time.Second,
		Factor:   2,
		Jitter:   0.1,
		Steps:    10,
	}

	var data = fmt.Sprintf("userid=%s&backupId=%s&snapshotId=%s&resticSnapshotId=%s&size=%d&realSnapshotSize=%d&unit=%s&snapshotTime=%d&status=%s&type=%s&url=%s&cloud=%s&region=%s&bucket=%s&prefix=%s&message=%s", snapshot.UserId, snapshot.BackupId,
		snapshot.SnapshotId, snapshot.ResticSnapshotId, snapshot.Size, snapshot.BackupSize, snapshot.Unit,
		snapshot.SnapshotTime, snapshot.Status, snapshot.Type,
		strings.TrimPrefix(snapshot.Url, "s3:"), snapshot.CloudName, snapshot.RegionId, snapshot.Bucket, snapshot.Prefix,
		snapshot.Message)

	log.Infof("[notify] snapshot data: %s", data)

	if err := retry.OnError(backoff, func(err error) bool {
		return true
	}, func() error {
		var url = fmt.Sprintf("%s%s", cloudApiUrl, SendSnapshotUrl)
		var headers = make(map[string]string)
		headers[restful.HEADER_ContentType] = "application/x-www-form-urlencoded"

		result, err := httpx.Post[Response](ctx, url, headers, data, true)
		if err != nil {
			return err
		}

		if result.Code != 200 && result.Code != 501 { // todo
			return fmt.Errorf("[notify] snapshot record failed, code: %d, message: %s", result.Code, result.Message)
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func NotifyStopBackup(ctx context.Context, cloudApiUrl string, userId, token, backupId string) error {
	var backoff = wait.Backoff{
		Duration: 2 * time.Second,
		Factor:   2,
		Jitter:   0.1,
		Steps:    3,
	}

	var data = fmt.Sprintf("userid=%s&token=%s&backupId=%s", userId, token, backupId)

	log.Infof("[notify] delete backup data: %s", data)

	if err := retry.OnError(backoff, func(err error) bool {
		return true
	}, func() error {
		var url = fmt.Sprintf("%s%s", cloudApiUrl, SendStopBackupUrl)
		var headers = make(map[string]string)
		headers[restful.HEADER_ContentType] = "application/x-www-form-urlencoded"

		result, err := httpx.Post[Response](ctx, url, headers, data, true)
		if err != nil {
			log.Errorf("[notify] delete backup record failed: %v", err)
			return err
		}

		if result.Code != 200 {
			return fmt.Errorf("[notify] delete backup record failed, code: %d, msg: %s", result.Code, result.Message)
		}
		return nil
	}); err != nil {
		log.Errorf(err.Error())
		return err
	}

	return nil
}

func CheckCloudStorageQuotaAndPermission(ctx context.Context, cloudApiUrl string, userId, token string) (*Usage, error) {
	var backoff = wait.Backoff{
		Duration: 2 * time.Second,
		Factor:   2,
		Jitter:   0.1,
		Steps:    3,
	}

	var data = fmt.Sprintf("userid=%s&token=%s", userId, token)
	var result *Usage

	log.Infof("[notify] check backup usage data: %s", data)

	if err := retry.OnError(backoff, func(err error) bool {
		return true
	}, func() error {
		var err error
		var url = fmt.Sprintf("%s%s", cloudApiUrl, CheckBackupUsageUrl)
		var headers = make(map[string]string)
		headers[restful.HEADER_ContentType] = "application/x-www-form-urlencoded"

		result, err = httpx.Post[Usage](ctx, url, headers, data, true)
		if err != nil {
			log.Errorf("[notify] check backup usage failed: %v", err)
			return err
		}

		if result.Code != 200 {
			log.Errorf("[notify] check backup usage failed, code: %d, msg: %s", result.Code, result.Message)
			return fmt.Errorf("check usage error, code: %d, message: %s", result.Code, result.Message)
		}

		if result.Data == nil {
			log.Errorf("[notify] check backup usage failed, data is nil")
			return fmt.Errorf("check backup usage failed, data is nil")
		}

		return nil
	}); err != nil {
		log.Errorf(err.Error())
		return nil, err
	}

	return result, nil
}

func parseBackupTypeCode(backupType string) int {
	switch backupType {
	case constant.BackupTypeApp:
		return 2
	default:
		return 1
	}
}

package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/v4/disk"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	sysv1 "olares.com/backup-server/pkg/apis/sys.bytetrade.io/v1"
	"olares.com/backup-server/pkg/constant"
	"olares.com/backup-server/pkg/handlers"
	integration "olares.com/backup-server/pkg/integration"
	"olares.com/backup-server/pkg/notify"
	"olares.com/backup-server/pkg/util"
	"olares.com/backup-server/pkg/util/log"
	"olares.com/backup-server/pkg/util/pointer"
	"olares.com/backup-server/pkg/watchers/notification"
	backupssdk "olares.com/backups-sdk"
	backupssdkoptions "olares.com/backups-sdk/pkg/options"
	backupssdkrestic "olares.com/backups-sdk/pkg/restic"
	backupssdkstorage "olares.com/backups-sdk/pkg/storage"
	backupssdkmodel "olares.com/backups-sdk/pkg/storage/model"
	"olares.com/backups-sdk/pkg/utils"
)

type StorageBackup struct {
	Handlers   handlers.Interface
	SnapshotId string
	Ctx        context.Context
	Cancel     context.CancelFunc

	Backup              *sysv1.Backup
	Snapshot            *sysv1.Snapshot
	Params              *BackupParameters
	SnapshotType        *int
	IntegrationChanged  bool
	IntegrationName     string
	IntegrationEndpoint string

	UserspacePvcName string
	UserspacePvcPath string
	AppcachePvcName  string
	AppcachePvcPath  string

	LastProgressPercent int
	LastProgressTime    time.Time

	BackupType           string
	BackupAppTypeName    string
	BackupAppStatus      *handlers.BackupAppStatus
	BackupAppFiles       []string
	BackupAppFilesPrefix string
	BackupAppMetadata    string

	BackupSize uint64
}

type BackupParameters struct {
	Password             string
	Path                 string
	SourcePath           string // file backup
	Location             map[string]string
	LocationInFileSystem string
	SnapshotType         string
}

type FilesPrefixData struct {
	Files []string `json:"files"`
}

func (s *StorageBackup) RunBackup() error {
	if err := s.checkBackupExists(); err != nil {
		return errors.WithStack(err)
	}

	var backupName = s.Backup.Spec.Name
	var snapshotId = s.Snapshot.Name

	var f = func() error {
		var e error
		if e = s.getUserspacePvc(); e != nil {
			return errors.WithStack(e)
		}

		if e = s.validateSnapshotPreconditions(); e != nil {
			return errors.WithStack(e)
		}

		if e = s.checkSnapshotType(); e != nil {
			return errors.WithStack(e)
		}

		if e = s.prepareBackupParams(); e != nil {
			return errors.WithStack(e)
		}

		if e = s.checkDiskSize(); e != nil {
			return errors.WithStack(e)
		}

		if e = s.prepareForRun(); e != nil {
			return errors.WithStack(e)
		}

		if e = s.readyForBackupApp(); e != nil {
			return errors.WithStack(e)
		}

		if e = s.formatAppBackupFiles(); e != nil {
			return errors.WithStack(e)
		}

		return nil
	}
	if err := f(); err != nil {
		log.Errorf("Backup %s,%s, prepare for run error: %v", backupName, snapshotId, err)
		if e := s.updateBackupResult(nil, nil, 0, err); e != nil {
			log.Errorf("Backup %s,%s, update backup failed result error: %v", backupName, snapshotId, e)
		}
		return s.sendAppBackupResult(err)
	}

	log.Infof("Backup %s,%s, locationConfig: %s, integrationChanged: %v ,ifAppStatus: %s", backupName, snapshotId, util.ToJSON(s.Params.Location), s.IntegrationChanged, util.ToJSON(s.BackupAppStatus))

	backupResult, backupStorageObj, backupTotalSize, backupErr := s.execute()
	if backupErr != nil {
		log.Errorf("Backup %s,%s, error: %v", backupName, snapshotId, backupErr)
	} else {
		log.Infof("Backup %s,%s, success, result: %s, storageObj: %s, totalSize: %d", backupName, snapshotId,
			util.ToJSON(backupResult), util.ToJSON(backupStorageObj), backupTotalSize)
	}

	if err := s.updateBackupResult(backupResult, backupStorageObj, backupTotalSize, backupErr); err != nil {
		log.Errorf("Backup %s,%s, update backup running result error: %v", backupName, snapshotId, err)
	} else {
		log.Infof("Backup %s,%s, backup completed", backupName, snapshotId)
	}

	return s.sendAppBackupResult(backupErr)
}

func (s *StorageBackup) checkBackupExists() error {
	snapshot, err := s.Handlers.GetSnapshotHandler().GetById(s.Ctx, s.SnapshotId)
	if err != nil {
		log.Errorf("snapshot not found: %v", err)
		return fmt.Errorf("snapshot not found")
	}
	backup, err := s.Handlers.GetBackupHandler().GetById(s.Ctx, snapshot.Spec.BackupId)
	if err != nil {
		log.Errorf("backup not found: %v", err)
		return fmt.Errorf("backup not found")
	}

	s.Backup = backup
	s.Snapshot = snapshot
	s.BackupType = handlers.GetBackupType(s.Backup)
	if s.BackupType == constant.BackupTypeApp {
		s.BackupAppTypeName = handlers.GetBackupAppName(s.Backup)
	}

	return nil
}

func (s *StorageBackup) getUserspacePvc() error {
	var backupName = s.Backup.Spec.Name
	var snapshotId = s.Snapshot.Name

	userspacePvcPath, userspacePvcName, appcachePvcPath, appcachePvcName, err := handlers.GetUserspacePvc(s.Backup.Spec.Owner)
	if err != nil {
		log.Errorf("Backup %s,%s, get userspace pvc error: %v", backupName, snapshotId, err)
		return fmt.Errorf("get userspace pvc error: %v", err)
	}

	s.UserspacePvcName = userspacePvcName
	s.UserspacePvcPath = userspacePvcPath
	s.AppcachePvcName = appcachePvcName
	s.AppcachePvcPath = appcachePvcPath

	return nil
}

func (s *StorageBackup) validateSnapshotPreconditions() error {
	var backupName = s.Backup.Spec.Name
	var snapshotId = s.Snapshot.Name
	var phase = *s.Snapshot.Spec.Phase
	if phase != constant.Pending.String() { // other phase ?
		log.Errorf("Backup %s,%s, snapshot phase %s invalid", backupName, snapshotId, phase)
		return fmt.Errorf("snapshot phase invalid")
	}
	return nil
}

func (s *StorageBackup) checkSnapshotType() error {
	snapshotType, err := s.Handlers.GetSnapshotHandler().GetSnapshotType(s.Ctx, s.Backup.Name)
	if err != nil {
		log.Errorf("Backup %s,%s, get snapshot type error: %v", s.Backup.Spec.Name, s.Snapshot.Name, err)
		return fmt.Errorf("list snapshots error: %v", err)
	}

	s.SnapshotType = handlers.ParseSnapshotType(snapshotType)
	return nil
}

func (s *StorageBackup) prepareBackupParams() error {
	var external, cache bool
	var backupSourcePath string
	var backupSourceRealPath string
	var locationInFileSystem string
	var backupName = s.Backup.Spec.Name
	var snapshotId = s.Snapshot.Name

	var token, err = integration.IntegrationService.GetAuthToken(s.Backup.Spec.Owner, fmt.Sprintf("user-system-%s", s.Backup.Spec.Owner), constant.DefaultServiceAccountSettings)
	if err != nil {
		log.Errorf("Backup %s,%s, get auth token error: %v", backupName, snapshotId, err)
		return err
	}

	password, err := handlers.GetBackupPassword(s.Ctx, s.Backup.Spec.Owner, s.Backup.Spec.Name, token)
	if err != nil {
		log.Errorf("Backup %s,%s, get password error: %v", backupName, snapshotId, err)
		return err
	}

	if s.BackupType == constant.BackupTypeFile {
		backupSourcePath = handlers.ParseBackupTypePath(s.Backup.Spec.BackupType)
		external, cache, backupSourceRealPath = handlers.TrimPathPrefix(backupSourcePath)
		var filesPrefix []string
		if external {
			filesPrefix = append(filesPrefix, filepath.Join(constant.ExternalPath, backupSourceRealPath))
		} else if cache {
			filesPrefix = append(filesPrefix, filepath.Join(s.AppcachePvcPath, backupSourceRealPath))
		} else {
			filesPrefix = append(filesPrefix, filepath.Join(s.UserspacePvcPath, backupSourceRealPath))
		}

		filesPrefixBytes, _ := json.Marshal(filesPrefix)
		s.BackupAppFilesPrefix = string(filesPrefixBytes)
	}

	location, err := handlers.GetBackupLocationConfig(s.Backup)
	if err != nil {
		log.Errorf("Backup %s,%s, get location config error: %v", backupName, snapshotId, err)
		return fmt.Errorf("get location config error: %v", err)
	}

	if location == nil {
		log.Errorf("Backup %s,%s, location config not exists", backupName, snapshotId)
		return fmt.Errorf("location config not exists")
	}

	loc := location["location"]
	if loc == constant.BackupLocationFileSystem.String() {
		locPath := location["path"]
		locationInFileSystem = locPath

		external, cache, locPath = handlers.TrimPathPrefix(locPath)
		if external {
			location["path"] = path.Join(constant.ExternalPath, locPath)
		} else if cache {
			location["path"] = path.Join(s.AppcachePvcPath, locPath)
		} else {
			location["path"] = path.Join(s.UserspacePvcPath, locPath)
		}
	}

	var backupPath string
	if s.BackupType == constant.BackupTypeFile {
		var tmpBackupExternal, tmpCache, tmpBackupPath = handlers.TrimPathPrefix(handlers.GetBackupPath(s.Backup))
		if tmpBackupExternal {
			backupPath = path.Join(constant.ExternalPath, tmpBackupPath)
		} else if tmpCache {
			backupPath = path.Join(s.AppcachePvcPath, tmpBackupPath)
		} else {
			backupPath = path.Join(s.UserspacePvcPath, tmpBackupPath)
		}
	}

	s.Params = &BackupParameters{
		Path:                 backupPath,
		SourcePath:           backupSourcePath,
		Password:             password,
		Location:             location,
		LocationInFileSystem: locationInFileSystem,
	}

	return nil
}

func (s *StorageBackup) checkDiskSize() error {
	if s.BackupType == constant.BackupTypeApp {
		return nil
	}
	var backupName = s.Backup.Spec.Name
	var snapshotId = s.Snapshot.Name

	var location = s.Params.Location["location"]

	if location == constant.BackupLocationSpace.String() {
		backupSize, err := util.DirSize(s.Params.Path)
		if err != nil {
			log.Errorf("Backup %s,%s, get backup disk size error: %v, path: %s", backupName, snapshotId, err, s.Params.Path)
			return fmt.Errorf("get backup disk size error: %v, path: %s", err, s.Params.Path)
		}
		s.BackupSize = backupSize
	}

	if location == constant.BackupLocationFileSystem.String() {
		var target = s.Params.Location["path"]

		usage, err := disk.Usage(target)
		if err != nil {
			log.Errorf("Backup %s,%s, check disk free space error: %v, path: %s", backupName, snapshotId, err, target)
			return err
		}

		log.Infof("Backup %s,%s, check disk free space: %s, path: %s", backupName, snapshotId, usage.String(), target)

		if usage.UsedPercent > FreeLimit {
			log.Errorf("Backup %s,%s, disk usage has reached %.2f%%", backupName, snapshotId, FreeLimit)
			return fmt.Errorf("Disk usage has reached %.2f%%. Please clean up disk space first.", FreeLimit)
		}

		backupSize, err := util.DirSize(s.Params.Path)
		if err != nil {
			log.Errorf("Backup %s,%s, get backup disk size error: %v, path: %s", backupName, snapshotId, err, s.Params.Path)
			return fmt.Errorf("get backup disk size error: %v, path: %s", err, s.Params.Path)
		}

		requiredSpace := uint64(float64(backupSize) * 1.05)
		if usage.Free < requiredSpace {
			log.Errorf("not enough free space on target disk, required: %s, available: %s, location: %s", util.FormatBytes(requiredSpace), util.FormatBytes(usage.Free), s.Params.LocationInFileSystem)
			return fmt.Errorf("Insufficient space on the target disk.")
		}
	}

	return nil
}

func (s *StorageBackup) prepareForRun() error {
	var data = map[string]interface{}{
		"id":       s.Snapshot.Name,
		"type":     constant.MessageTypeBackup,
		"backupId": s.Backup.Name,
		"progress": 0,
		"size":     "0",
		"status":   constant.Running.String(),
		"message":  "",
	}
	notification.DataSender.Send(s.Backup.Spec.Owner, data)

	return s.Handlers.GetSnapshotHandler().UpdatePhase(s.Ctx, s.Snapshot.Name, constant.Running.String(), "Backup start running")
}

func (s *StorageBackup) progressCallback(percentDone float64) {

	select {
	case <-s.Ctx.Done():
		return
	default:
	}

	var percent = int(percentDone * progressDone)

	if percent == progressDone {
		percent = progressDone - 1
		s.Handlers.GetSnapshotHandler().UpdateProgress(s.Ctx, s.SnapshotId, percent)

		var data = map[string]interface{}{
			"id":       s.Snapshot.Name,
			"type":     constant.MessageTypeBackup,
			"backupId": s.Backup.Name,
			"progress": percent,
			"status":   constant.Running.String(),
			"message":  "",
		}
		notification.DataSender.Send(s.Backup.Spec.Owner, data)

		return
	}

	if time.Since(s.LastProgressTime) >= progressInterval*time.Second && s.LastProgressPercent != percent {
		s.LastProgressPercent = percent
		s.LastProgressTime = time.Now()

		s.Handlers.GetSnapshotHandler().UpdateProgress(s.Ctx, s.SnapshotId, percent)

		var data = map[string]interface{}{
			"id":       s.Snapshot.Name,
			"type":     constant.MessageTypeBackup,
			"backupId": s.Backup.Name,
			"progress": percent,
			"status":   constant.Running.String(),
			"message":  "",
		}
		notification.DataSender.Send(s.Backup.Spec.Owner, data)
	}
}

func (s *StorageBackup) execute() (backupOutput *backupssdkrestic.SummaryOutput,
	backupStorageObj *backupssdkmodel.StorageInfo, backupTotalSize uint64, backupError error) {
	var isSpaceBackup bool
	var logger = log.GetLogger()
	var backupId = s.Backup.Name
	var backupName = s.Backup.Spec.Name
	var snapshotId = s.Snapshot.Name
	var location = s.Params.Location["location"]
	var backupType = s.BackupType
	var backupAppTypeName = s.BackupAppTypeName

	log.Infof("Backup %s,%s, location: %s, backupType: %s, prepare", backupName, snapshotId, location, backupType)

	var backupService *backupssdkstorage.BackupService
	var options backupssdkoptions.Option

	switch location {
	case constant.BackupLocationSpace.String():
		isSpaceBackup = true
		backupOutput, backupStorageObj, backupTotalSize, backupError = s.backupToSpace()
	case constant.BackupLocationAwsS3.String():
		token, err := s.getIntegrationCloud() // aws backup
		if err != nil {
			backupError = fmt.Errorf("get %s token error: %v", token.Type, err)
			return
		}
		options = &backupssdkoptions.AwsBackupOption{
			RepoId:          backupId,
			RepoName:        backupName,
			Path:            s.Params.Path,
			Files:           s.BackupAppFiles,
			FilesPrefixPath: s.BackupAppFilesPrefix,
			Metadata:        s.BackupAppMetadata,
			Endpoint:        token.Endpoint,
			AccessKey:       token.AccessKey,
			SecretAccessKey: token.SecretKey,
			LimitUploadRate: util.EnvOrDefault(constant.EnvLimitUploadRate, ""),
		}
		backupService = backupssdk.NewBackupService(&backupssdkstorage.BackupOption{
			Password:                 s.Params.Password,
			Operator:                 constant.StorageOperatorApp,
			BackupType:               backupType,
			BackupAppTypeName:        backupAppTypeName,
			BackupFileTypeSourcePath: s.Params.SourcePath,
			Ctx:                      s.Ctx,
			Logger:                   logger,
			Aws:                      options.(*backupssdkoptions.AwsBackupOption),
		})
	case constant.BackupLocationTencentCloud.String():
		token, err := s.getIntegrationCloud() // cos backup
		if err != nil {
			backupError = fmt.Errorf("get tencentcloud token error: %v", err)
			return
		}
		options = &backupssdkoptions.TencentCloudBackupOption{
			RepoId:          backupId,
			RepoName:        backupName,
			Path:            s.Params.Path,
			Files:           s.BackupAppFiles,
			FilesPrefixPath: s.BackupAppFilesPrefix,
			Metadata:        s.BackupAppMetadata,
			Endpoint:        token.Endpoint,
			AccessKey:       token.AccessKey,
			SecretAccessKey: token.SecretKey,
			LimitUploadRate: util.EnvOrDefault(constant.EnvLimitUploadRate, ""),
		}
		backupService = backupssdk.NewBackupService(&backupssdkstorage.BackupOption{
			Password:                 s.Params.Password,
			Operator:                 constant.StorageOperatorApp,
			BackupType:               backupType,
			BackupAppTypeName:        backupAppTypeName,
			BackupFileTypeSourcePath: s.Params.SourcePath,
			Ctx:                      s.Ctx,
			Logger:                   logger,
			TencentCloud:             options.(*backupssdkoptions.TencentCloudBackupOption),
		})
	case constant.BackupLocationFileSystem.String():
		options = &backupssdkoptions.FilesystemBackupOption{
			RepoId:          backupId,
			RepoName:        backupName,
			Endpoint:        s.Params.Location["path"],
			Path:            s.Params.Path,
			Files:           s.BackupAppFiles,
			FilesPrefixPath: s.BackupAppFilesPrefix,
			Metadata:        s.BackupAppMetadata,
		}
		backupService = backupssdk.NewBackupService(&backupssdkstorage.BackupOption{
			Password:                 s.Params.Password,
			Operator:                 constant.StorageOperatorApp,
			BackupType:               backupType,
			BackupAppTypeName:        backupAppTypeName,
			BackupFileTypeSourcePath: s.Params.SourcePath,
			Ctx:                      s.Ctx,
			Logger:                   logger,
			Filesystem:               options.(*backupssdkoptions.FilesystemBackupOption),
		})
	}

	if !isSpaceBackup {
		backupOutput, backupStorageObj, backupError = backupService.Backup(false, s.progressCallback)
		if backupError == nil {
			stats, err := s.getStats(options)
			if err != nil {
				log.Errorf("Backup %s,%s, get stats error: %v", backupName, snapshotId, err)
			} else {
				log.Infof("Backup %s,%s, get stats: %s", backupName, snapshotId, util.ToJSON(stats))
				backupTotalSize = stats.TotalSize
			}
		}
	}

	return
}

func (s *StorageBackup) backupToSpace() (backupOutput *backupssdkrestic.SummaryOutput, backupStorageObj *backupssdkmodel.StorageInfo, totalSize uint64, err error) {
	var backupId = s.Backup.Name
	var backupName = s.Backup.Spec.Name
	var location = s.Params.Location
	var olaresId = location["name"]

	var spaceToken *integration.SpaceToken
	var spaceBackupOption backupssdkoptions.Option
	var usage *notify.Usage
	var compareUsage bool

	for {
		spaceToken, err = integration.IntegrationManager().GetIntegrationSpaceToken(s.Ctx, s.Backup.Spec.Owner, olaresId) // backupToSpace
		if err != nil {
			log.Errorf("Backup %s,%s, get space token error: %v", backupName, s.Snapshot.Name, err)
			break
		}

		spaceBackupOption = &backupssdkoptions.SpaceBackupOption{
			RepoId:          backupId,
			RepoName:        backupName,
			Path:            s.Params.Path,
			Files:           s.BackupAppFiles,
			FilesPrefixPath: s.BackupAppFilesPrefix,
			Metadata:        s.BackupAppMetadata,
			OlaresDid:       spaceToken.OlaresDid,
			AccessToken:     spaceToken.AccessToken,
			ClusterId:       location["clusterId"],
			CloudName:       location["cloudName"],
			RegionId:        location["regionId"],
			CloudApiMirror:  constant.SyncServerURL,
			LimitUploadRate: util.EnvOrDefault(constant.EnvLimitUploadRate, ""),
		}

		var backupService = backupssdk.NewBackupService(&backupssdkstorage.BackupOption{
			Password:                 s.Params.Password,
			Operator:                 constant.StorageOperatorApp,
			BackupType:               s.BackupType,
			BackupAppTypeName:        s.BackupAppTypeName,
			BackupFileTypeSourcePath: s.Params.SourcePath,
			Ctx:                      s.Ctx,
			Logger:                   log.GetLogger(),
			Space:                    spaceBackupOption.(*backupssdkoptions.SpaceBackupOption),
		})

		// todo need to enhance
		if !compareUsage {
			usage, err = notify.CheckCloudStorageQuotaAndPermission(s.Ctx, constant.SyncServerURL, spaceToken.OlaresDid, spaceToken.AccessToken)
			if err != nil {
				log.Errorf("Backup %s,%s, check cloud storage quota and permission error: %v", backupName, s.SnapshotId, err)
				break
			}

			log.Infof("Backup %s,%s, usage: %s", backupName, s.SnapshotId, util.ToJSON(usage))
			if usage.Data.PlanLevel == constant.FreeUser {
				err = errors.New("You are not currently subscribed to Olares Space.")
				break
			}

			if !usage.Data.CanBackup || (s.BackupSize > (usage.Data.ToTalSize - usage.Data.UsageSize)) {
				err = errors.New("Insufficient storage in Olares Space.")
				break
			}

			compareUsage = true
		}

		backupOutput, backupStorageObj, err = backupService.Backup(false, s.progressCallback)

		if err != nil {
			if strings.Contains(err.Error(), "refresh-token error") {
				continue
			} else {
				// err = fmt.Errorf("space backup error: %v", err)
				break
			}
		}
		break
	}
	if err == nil {
		stats, err := s.getStats(spaceBackupOption)
		if err != nil {

			log.Errorf("Backup %s,%s, get stats error: %v", s.Backup.Spec.Name, s.SnapshotId, err)
		} else {
			totalSize = stats.TotalSize
		}
	}

	return
}

func (s *StorageBackup) getStats(opt backupssdkoptions.Option) (*backupssdkrestic.StatsContainer, error) {
	var options = &backupssdkstorage.SnapshotsOption{
		Password: s.Params.Password,
		Logger:   log.GetLogger(),
	}

	switch opt.(type) {
	case *backupssdkoptions.SpaceBackupOption:
		var location = s.Params.Location
		var olaresId = location["name"]
		spaceToken, err := integration.IntegrationManager().GetIntegrationSpaceToken(s.Ctx, s.Backup.Spec.Owner, olaresId) // space  getStats
		if err != nil {
			err = fmt.Errorf("get space token error: %v", err)
			break
		}
		o := opt.(*backupssdkoptions.SpaceBackupOption)
		options.Space = &backupssdkoptions.SpaceSnapshotsOption{
			RepoId:         o.RepoId,
			RepoName:       o.RepoName,
			OlaresDid:      spaceToken.OlaresDid,
			AccessToken:    spaceToken.AccessToken,
			ClusterId:      location["clusterId"],
			CloudName:      location["cloudName"],
			RegionId:       location["regionId"],
			CloudApiMirror: constant.SyncServerURL,
		}
	case *backupssdkoptions.AwsBackupOption:
		o := opt.(*backupssdkoptions.AwsBackupOption)
		options.Aws = &backupssdkoptions.AwsSnapshotsOption{
			RepoId:          o.RepoId,
			RepoName:        o.RepoName,
			Endpoint:        o.Endpoint,
			AccessKey:       o.AccessKey,
			SecretAccessKey: o.SecretAccessKey,
		}
	case *backupssdkoptions.TencentCloudBackupOption:
		o := opt.(*backupssdkoptions.TencentCloudBackupOption)
		options.TencentCloud = &backupssdkoptions.TencentCloudSnapshotsOption{
			RepoId:          o.RepoId,
			RepoName:        o.RepoName,
			Endpoint:        o.Endpoint,
			AccessKey:       o.AccessKey,
			SecretAccessKey: o.SecretAccessKey,
		}
	case *backupssdkoptions.FilesystemBackupOption:
		o := opt.(*backupssdkoptions.FilesystemBackupOption)
		options.Filesystem = &backupssdkoptions.FilesystemSnapshotsOption{
			RepoId:   o.RepoId,
			RepoName: o.RepoName,
			Endpoint: o.Endpoint,
		}
	}

	statsService := backupssdk.NewStatsService(options)
	return statsService.Stats()
}

func (s *StorageBackup) updateBackupResult(backupOutput *backupssdkrestic.SummaryOutput,
	backupStorageObj *backupssdkmodel.StorageInfo, backupTotalSize uint64, backupError error) error {

	return wait.PollImmediate(time.Second*10, 5*time.Hour, func() (bool, error) {
		log.Infof("Backup %s,%s, update backup result, if err: %v", s.Backup.Spec.Name, s.SnapshotId, backupError)
		select {
		case <-s.Ctx.Done():
			return true, nil
		default:
			var msg string
			var endAt = pointer.Time()

			backup, err := s.Handlers.GetBackupHandler().GetById(s.Ctx, s.Backup.Name)
			if err != nil {
				log.Errorf("Backup %s,%s, get backup error: %v", s.Backup.Spec.Name, s.SnapshotId, err)
				if apierrors.IsNotFound(err) {
					return true, nil
				}
				if util.ListMatchContains(ConnectErrors, err.Error()) {
					return false, nil
				}
			}

			// todo snapshot.Spec.Extra
			snapshot, err := s.Handlers.GetSnapshotHandler().GetById(s.Ctx, s.Snapshot.Name)
			if err != nil {
				log.Errorf("Backup %s,%s, get snapshot error: %v", s.Backup.Spec.Name, s.SnapshotId, err)
				if apierrors.IsNotFound(err) {
					return true, nil
				}
				if util.ListMatchContains(ConnectErrors, err.Error()) {
					log.Errorf("Backup %s,%s, get snapshot error: %v, id: %s", s.Backup.Spec.Name, s.SnapshotId, err)
					return false, nil
				}
			}

			var eventData = make(map[string]interface{})
			eventData["id"] = s.Snapshot.Name
			eventData["type"] = constant.MessageTypeBackup
			eventData["backupId"] = s.Backup.Name
			eventData["endat"] = endAt.Unix()

			var phase constant.Phase = constant.Completed

			if backupError != nil {
				msg = backupError.Error()
				if strings.Contains(backupError.Error(), strings.ToLower(constant.Canceled.String())) {
					phase = constant.Canceled
				} else {
					phase = constant.Failed
				}
				snapshot.Spec.Phase = pointer.String(phase.String())
				snapshot.Spec.Message = pointer.String(msg)
				snapshot.Spec.ResticPhase = pointer.String(phase.String())
				if s.SnapshotType != nil {
					snapshot.Spec.SnapshotType = s.SnapshotType
				}
				eventData["status"] = phase.String()
				eventData["message"] = msg
			} else {
				msg = phase.String()
				eventData["size"] = fmt.Sprintf("%d", backupOutput.TotalBytesProcessed)
				eventData["totalSize"] = fmt.Sprintf("%d", backupTotalSize)
				eventData["restoreSize"] = fmt.Sprintf("%d", backupOutput.RestoreSize)
				eventData["progress"] = progressDone
				eventData["status"] = phase.String()
				eventData["message"] = msg
				snapshot.Spec.SnapshotType = s.SnapshotType
				snapshot.Spec.SnapshotId = pointer.String(backupOutput.SnapshotID)
				snapshot.Spec.Size = pointer.UInt64Ptr(backupOutput.TotalBytesProcessed)
				snapshot.Spec.Progress = progressDone
				snapshot.Spec.Phase = pointer.String(phase.String())
				snapshot.Spec.Message = pointer.String(phase.String())
				snapshot.Spec.ResticPhase = pointer.String(phase.String())
				snapshot.Spec.ResticMessage = pointer.String(util.ToJSON(backupOutput))
			}

			snapshot.Spec.EndAt = endAt

			var extra = snapshot.Spec.Extra
			if extra == nil {
				extra = make(map[string]string)
			}
			extra["backup_total_size"] = fmt.Sprintf("%d", backupTotalSize)
			if backupStorageObj != nil {
				extra["storage"] = util.ToJSON(backupStorageObj)
			}
			if s.BackupType == constant.BackupTypeApp {
				extra["app_metadata"] = util.ToJSON(s.BackupAppStatus)
			}

			snapshot.Spec.Extra = extra

			if backupOutput != nil {
				var newLocation, newLocationData = s.buildLocation()
				if err := s.Handlers.GetBackupHandler().UpdateTotalSize(s.Ctx, backup, backupTotalSize, backupOutput.RestoreSize, newLocation, newLocationData); err != nil {
					log.Errorf("Backup %s,%s, update backup total size error: %v", backup.Spec.Name, s.Snapshot.Name, err)
				}
			}

			notification.DataSender.Send(s.Backup.Spec.Owner, eventData)

			err = s.Handlers.GetSnapshotHandler().UpdateBackupResult(s.Ctx, snapshot)
			if err != nil {
				return false, err
			}
			return true, nil
		}
	})

}

func (s *StorageBackup) sendAppBackupResult(backupError error) error {
	if s.BackupType != constant.BackupTypeApp {
		return nil
	}

	var appName = handlers.GetBackupAppName(s.Backup)
	var appHandler = handlers.NewAppHandler(appName, s.Backup.Spec.Owner)

	return appHandler.SendBackupResult(s.Ctx, s.Backup.Name, s.Snapshot.Name, backupError)
}

func (s *StorageBackup) getIntegrationCloud() (*integration.IntegrationToken, error) {
	var l = s.Params.Location
	var location = l["location"]
	var locationIntegrationName = l["name"]
	var endpoint = l["endpoint"]

	accounts, err := integration.IntegrationManager().GetIntegrationAccountsByLocation(s.Ctx, s.Backup.Spec.Owner, location)
	if err != nil {
		return nil, err
	}

	var tokens []*integration.IntegrationToken

	for _, account := range accounts {
		token, _ := integration.IntegrationManager().GetIntegrationCloudToken(s.Ctx, s.Backup.Spec.Owner, location, account)
		tokens = append(tokens, token)
	}

	if tokens == nil || len(tokens) == 0 {
		return nil, fmt.Errorf("no integration cloud tokens found")
	}

	var result *integration.IntegrationToken

	for _, t := range tokens {
		if t.AccessKey == locationIntegrationName {
			result = t
			break
		}
	}

	if result != nil {
		s.IntegrationChanged = false
		return result, nil
	}

	for _, t := range tokens {
		if t.Endpoint == endpoint {
			result = t
			break
		}
	}

	if result == nil {
		return nil, fmt.Errorf("the integration token was not found, or the endpoint does not match. please check the integration configuration, backup endpoint: %s", endpoint)
	}

	s.IntegrationChanged = true
	s.IntegrationName = result.AccessKey
	s.IntegrationEndpoint = result.Endpoint

	return result, nil
}

func (s *StorageBackup) buildLocation() (string, string) {
	var l = s.Params.Location
	var location = l["location"]

	if !s.IntegrationChanged {
		return "", ""
	}

	if location == constant.BackupLocationSpace.String() || location == constant.BackupLocationFileSystem.String() {
		return "", ""
	}

	var locationData = make(map[string]string)
	locationData["name"] = s.IntegrationName
	locationData["endpoint"] = s.IntegrationEndpoint

	return location, util.ToJSON(locationData)
}

func (s *StorageBackup) readyForBackupApp() error {
	if s.BackupType != constant.BackupTypeApp {
		return nil
	}

	var err error

	var appName = handlers.GetBackupAppName(s.Backup)
	var appHandler = handlers.NewAppHandler(appName, s.Backup.Spec.Owner)

	err = appHandler.StartAppBackup(s.Ctx, s.Backup.Name, s.Snapshot.Name)
	if err != nil {
		log.Errorf("Backup %s,%s,%s, start app backup error: %v", s.Backup.Spec.Name, s.Snapshot.Name, appName, err)
		return err
	}

	log.Infof("Backup %s,%s,%s, waiting for check app backup status", s.Backup.Spec.Name, s.Snapshot.Name, appName)

	var ticker = time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			result, e := appHandler.GetAppBackupStatus(s.Ctx, s.Backup.Name, s.Snapshot.Name)
			if e != nil {
				log.Errorf("Backup %s,%s,%s, get app backup status error: %v", s.Backup.Spec.Name, s.Snapshot.Name, appName, e)
				err = fmt.Errorf("get app backup status error: %v", e)
				return err
			}

			log.Infof("Backup %s,%s,%s, get app backup status, data: %s", s.Backup.Spec.Name, s.Snapshot.Name, appName, util.ToJSON(result))

			if result.Data == nil {
				log.Errorf("Backup %s,%s,%s, get app backup status error, data is nil", s.Backup.Spec.Name, s.Snapshot.Name, appName)
				err = fmt.Errorf("get app backup status error, data is nil")
				return err
			}

			if result.Data.Status == constant.BackupAppStatusPreparing {
				continue
			}

			if result.Data.Status == constant.BackupAppStatusFinish {
				result.Data.EntryFiles = util.TrimArrayPrefix(result.Data.EntryFiles, "/olares")
				result.Data.PgFiles = util.TrimArrayPrefix(result.Data.PgFiles, "/olares")
				s.BackupAppStatus = result
				return nil
			}
		case <-s.Ctx.Done():
			if e := s.Ctx.Err(); e != nil {
				if errors.Is(err, context.DeadlineExceeded) {
					log.Errorf("Backup %s,%s,%s, start app backup timeout", s.Backup.Spec.Name, s.Snapshot.Name, appName)
					return fmt.Errorf("app backup status timeout")
				}
				if errors.Is(err, context.Canceled) {
					log.Errorf("Backup %s,%s,%s, app backup canceled", s.Backup.Spec.Name, s.Snapshot.Name, appName)
					return fmt.Errorf("backup canceled")
				}
			}
			return nil
		}
	}
}

func (s *StorageBackup) formatAppBackupFiles() error {
	if s.BackupType != constant.BackupTypeApp {
		return nil
	}

	var entryFiles = s.BackupAppStatus.Data.EntryFiles
	var pgFiles = s.BackupAppStatus.Data.PgFiles
	var files []string

	if len(entryFiles) > 0 {
		for _, f := range entryFiles {
			if err := s.replaceFilePaths(f, filepath.Join(s.UserspacePvcPath, "Home"), true); err != nil {
				return err
			} else {
				files = append(files, f)
			}
		}
	}

	if len(pgFiles) > 0 {
		for _, f := range pgFiles {
			if err := s.replaceFilePaths(f, "/olares", false); err != nil {
				return err
			} else {
				files = append(files, f)
			}
		}
	}

	var filesPrefix []string = []string{
		filepath.Join(s.UserspacePvcPath, "Home"),
		filepath.Join("/rootfs/middleware-backup/pg_backup", s.Backup.Spec.Owner),
	}
	filesPrefixBytes, _ := json.Marshal(filesPrefix)
	s.BackupAppFilesPrefix = string(filesPrefixBytes)
	s.BackupAppFiles = files
	s.BackupAppMetadata = utils.ToJSON(s.BackupAppStatus.Data.Data)

	return nil
}

func (s *StorageBackup) replaceFilePaths(fp string, replacedFilePathPrefix string, appendPrefix bool) error {
	tempFilePath := fp + ".tmp"

	srcFile, err := os.Open(fp)
	if err != nil {
		return fmt.Errorf("failed to open source file: %v", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(tempFilePath)
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %v", err)
	}

	var maxScanTokenSize = 512 * 1024

	writer := bufio.NewWriter(dstFile)

	scanner := bufio.NewScanner(srcFile)

	buf := make([]byte, maxScanTokenSize)
	scanner.Buffer(buf, maxScanTokenSize)

	for scanner.Scan() {
		line := scanner.Text()
		var newLine string

		if appendPrefix {
			newLine = filepath.Join(replacedFilePathPrefix, line)
		} else {
			newLine = strings.TrimPrefix(line, replacedFilePathPrefix)
		}

		if _, err := writer.WriteString(newLine + "\n"); err != nil {
			writer.Flush()
			dstFile.Close()
			os.Remove(tempFilePath)
			return fmt.Errorf("failed to write to file: %v", err)
		}
	}

	if err := scanner.Err(); err != nil {
		writer.Flush()
		dstFile.Close()
		os.Remove(tempFilePath)
		return fmt.Errorf("failed to scan file: %v", err)
	}

	if err := writer.Flush(); err != nil {
		dstFile.Close()
		os.Remove(tempFilePath)
		return fmt.Errorf("failed to flush writer: %v", err)
	}

	if err := dstFile.Close(); err != nil {
		os.Remove(tempFilePath)
		return fmt.Errorf("failed to close destination file: %v", err)
	}

	if err := os.Rename(tempFilePath, fp); err != nil {
		os.Remove(tempFilePath)
		return fmt.Errorf("failed to replace original file: %v", err)
	}

	return nil
}

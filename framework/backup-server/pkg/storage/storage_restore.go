package storage

import (
	"context"
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
	"olares.com/backup-server/pkg/util"
	"olares.com/backup-server/pkg/util/log"
	"olares.com/backup-server/pkg/util/pointer"
	"olares.com/backup-server/pkg/watchers/notification"
	backupssdk "olares.com/backups-sdk"

	backupssdkoptions "olares.com/backups-sdk/pkg/options"
	backupssdkrestic "olares.com/backups-sdk/pkg/restic"
	backupssdkstorage "olares.com/backups-sdk/pkg/storage"
)

type StorageRestore struct {
	Handlers  handlers.Interface
	RestoreId string
	Ctx       context.Context
	Cancel    context.CancelFunc

	Backup           *sysv1.Backup
	Snapshot         *sysv1.Snapshot
	Restore          *sysv1.Restore
	RestoreType      *handlers.RestoreType
	BackupType       string
	RestoreAppStatus *handlers.RestoreAppStatusData

	UserspacePvcName string
	UserspacePvcPath string
	AppcachePvcName  string
	AppcachePvcPath  string
	Params           *RestoreParameters

	LastProgressPercent int
	LastProgressTime    time.Time
}

type RestoreParameters struct {
	BackupId   string
	BackupName string
	Password   string
	RootPath   string
	Path       string
	Location   map[string]string
}

func (s *StorageRestore) RunRestore() error {
	if err := s.checkRestoreExists(); err != nil {
		return errors.WithStack(err)
	}

	var f = func() error {
		var e error
		if e = s.getUserspacePvc(); e != nil {
			return errors.WithStack(e)
		}

		if e = s.prepareRestoreParams(); e != nil {
			return errors.WithStack(e)
		}

		if e = s.checkDiskSize(); e != nil {
			return errors.WithStack(e)
		}

		if e = s.prepareForRun(); e != nil {
			return errors.WithStack(e)
		}
		return nil
	}

	if err := f(); err != nil {
		if err := s.updateRestoreResult(nil, "", 0, err); err != nil {
			log.Errorf("Restore %s, update restore failed result error: %v", s.RestoreId, err)
		}
		return nil
	}

	restoreResult, metadata, totalBytes, restoreErr := s.execute()

	if restoreErr != nil {
		log.Errorf("Restore %s, error: %v", s.RestoreId, restoreErr)
		s.removeRestoreFiles()
	} else {
		s.renameRestoreDirectory()
		if err := s.readyForRestoreApp(s.RestoreType.ResticSnapshotId, metadata); err != nil {
			restoreErr = err
		}
	}

	log.Infof("Restore %s success, result: %s", s.RestoreId, util.ToJSON(restoreResult))

	if err := s.updateRestoreResult(restoreResult, metadata, totalBytes, restoreErr); err != nil {
		log.Errorf("Restore %s, update restore result error: %v", s.RestoreId, err)
	} else {
		log.Infof("Restore %s, restore completed", s.RestoreId)
	}

	return nil
}

func (s *StorageRestore) getUserspacePvc() error {
	userspacePvcPath, userspacePvcName, appcachePvcPath, appcachePvcName, err := handlers.GetUserspacePvc(s.Restore.Spec.Owner)
	if err != nil {
		log.Errorf("Restore %s, get userspace pvc error: %v", s.RestoreId, err)
		return fmt.Errorf("get userspace pvc error: %v", err)
	}

	log.Infof("Restore %s, upvcName: %s, upvcPath: %s, cpvcName: %s, cpvcPath: %s", s.RestoreId, userspacePvcName, userspacePvcPath, appcachePvcName, appcachePvcPath)

	s.UserspacePvcName = userspacePvcName
	s.UserspacePvcPath = userspacePvcPath
	s.AppcachePvcName = appcachePvcName
	s.AppcachePvcPath = appcachePvcPath

	return nil
}

func (s *StorageRestore) checkRestoreExists() error {
	restore, err := s.Handlers.GetRestoreHandler().GetById(s.Ctx, s.RestoreId)
	if err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("get restore %s error: %v", s.RestoreId, err)
	}

	if restore == nil {
		return fmt.Errorf("restore %s not exists", s.RestoreId)
	}

	backupType, err := handlers.GetRestoreType(restore)
	if err != nil {
		return fmt.Errorf("restore %s backup type invalid", s.RestoreId)
	}

	restoreType, err := handlers.ParseRestoreType(backupType, restore)
	if err != nil {
		return fmt.Errorf("restore %s type %v invalid", s.RestoreId, restore.Spec.RestoreType)
	}

	log.Infof("restore %s, backupType: %s, data: %s", s.RestoreId, backupType, util.ToJSON(restoreType))
	if restoreType.Type == constant.RestoreTypeSnapshot {
		snapshot, err := s.Handlers.GetSnapshotHandler().GetById(s.Ctx, restoreType.SnapshotId)
		if err != nil {
			return fmt.Errorf("snapshot not found: %v", err)
		}
		backup, err := s.Handlers.GetBackupHandler().GetById(s.Ctx, snapshot.Spec.BackupId)
		if err != nil {
			return fmt.Errorf("backup not found: %v", err)
		}

		s.Backup = backup
		s.Snapshot = snapshot
	}

	s.Restore = restore
	s.RestoreType = restoreType
	s.BackupType = backupType

	return nil
}

func (s *StorageRestore) prepareRestoreParams() error {
	var backupId string
	var backupName string
	var password string
	var locationConfig = make(map[string]string)
	var err error
	var token string
	var external, cache bool

	if s.RestoreType.Type == constant.RestoreTypeSnapshot {
		token, err = integration.IntegrationService.GetAuthToken(s.Backup.Spec.Owner, fmt.Sprintf("user-system-%s", s.Backup.Spec.Owner), constant.DefaultServiceAccountSettings)
		if err != nil {
			log.Errorf("Restore %s get auth token error: %v", s.RestoreId, err)
			return err
		}
		backupId = s.Backup.Name
		backupName = s.Backup.Spec.Name
		password, err = handlers.GetBackupPassword(s.Ctx, s.Backup.Spec.Owner, s.Backup.Spec.Name, token)
		if err != nil {
			return fmt.Errorf("Restore %s get password error: %v", s.RestoreId, err)
		}

		locationConfig, err = handlers.GetBackupLocationConfig(s.Backup)
		if err != nil {
			return fmt.Errorf("Restore %s get location config error: %v", s.RestoreId, err)
		}
		if locationConfig == nil {
			return fmt.Errorf("Restore %s location config not found", s.RestoreId)
		}

		location := locationConfig["location"]
		if location == constant.BackupLocationFileSystem.String() {
			locPath := locationConfig["path"]
			external, cache, locPath = handlers.TrimPathPrefix(locPath)
			if external {
				locationConfig["path"] = path.Join(constant.ExternalPath, locPath)
			} else if cache {
				locationConfig["path"] = path.Join(s.AppcachePvcPath, locPath)
			} else {
				locationConfig["path"] = path.Join(s.UserspacePvcPath, locPath)
			}
		}
	} else {
		// ~ backupUrl
		log.Infof("restore from backupUrl, ready to get integration token, owner: %s, location: %s, prefix: %s", s.RestoreType.Owner, s.RestoreType.Location, s.RestoreType.BackupUrl.Prefix)
		integrationName, err := integration.IntegrationManager().GetIntegrationNameByLocation(s.Ctx,
			s.RestoreType.Owner,
			s.RestoreType.Location,
			s.RestoreType.BackupUrl.Bucket,
			s.RestoreType.BackupUrl.RegionId,
			s.RestoreType.BackupUrl.Prefix,
		)
		if err != nil {
			log.Errorf("get restore integration name error: %v", err)
			return err
		}

		locationConfig["location"] = s.RestoreType.Location
		locationConfig["name"] = integrationName
		locationConfig["cloudName"] = s.RestoreType.BackupUrl.CloudName
		locationConfig["regionId"] = s.RestoreType.BackupUrl.RegionId
		locationConfig["clusterId"] = s.RestoreType.ClusterId
		locationConfig["suffix"] = s.RestoreType.BackupUrl.OlaresSuffix
		locationConfig["path"] = s.RestoreType.BackupUrl.FilesystemPath

		p, _ := util.Base64decode(s.RestoreType.Password)
		password = string(p)

		backupId = s.RestoreType.BackupId
		backupName = s.RestoreType.BackupName
	}

	var restorePath string
	var rootRestorePath string
	if s.BackupType == constant.BackupTypeFile {
		var dotRestorePath = path.Join(s.RestoreType.Path, fmt.Sprintf("%s.restore-%d", s.RestoreType.SubPath, s.RestoreType.SubPathTimestamp)) + "/"
		var dotRestoreRootPath = path.Join(s.RestoreType.Path)
		var tmpRestoreExternal, tmpRestoreCache, tmpRestorePath = handlers.TrimPathPrefix(dotRestorePath)
		log.Infof("Restore %s, dotRPath: %s, dotRRootPath: %s, tmpRestorePath: %s", s.RestoreId, dotRestorePath, dotRestoreRootPath, tmpRestorePath)
		log.Infof("Restore %s, cachePvcPath: %s, userPvcPath: %s,", s.RestoreId, s.AppcachePvcPath, s.UserspacePvcPath)
		var _, _, tmpRootRestorePath = handlers.TrimPathPrefix(dotRestoreRootPath)
		if tmpRestoreExternal {
			restorePath = path.Join(constant.ExternalPath, tmpRestorePath)
			rootRestorePath = path.Join(constant.ExternalPath, tmpRootRestorePath)
		} else if tmpRestoreCache {
			restorePath = path.Join(s.AppcachePvcPath, tmpRestorePath)
			rootRestorePath = path.Join(s.AppcachePvcPath, tmpRootRestorePath)
		} else {
			restorePath = path.Join(s.UserspacePvcPath, tmpRestorePath)
			rootRestorePath = path.Join(s.UserspacePvcPath, tmpRootRestorePath)
		}
	} else {
		restorePath = filepath.Join(s.UserspacePvcPath, "Home") // for app restore
		rootRestorePath = filepath.Join(s.UserspacePvcPath, "Home")
	}

	log.Infof("restore: %s, backupType: %s, rootRestoreTargetPath: %s, restoreTargetPath: %s, locationConfig: %v", s.RestoreId, s.BackupType, rootRestorePath, restorePath, util.ToJSON(locationConfig))
	s.Params = &RestoreParameters{
		BackupId:   backupId,
		BackupName: backupName,
		RootPath:   rootRestorePath,
		Path:       restorePath,
		Password:   password,
		Location:   locationConfig,
	}

	return nil
}

func (s *StorageRestore) checkDiskSize() error {
	log.Infof("Restore %s, check disk size, path: %s", s.RestoreId, s.Params.RootPath)

	usage, err := disk.Usage(s.Params.RootPath)
	if err != nil {
		log.Errorf("Restore %s check disk free space error: %v, path: %s", s.RestoreId, err, s.Params.RootPath)
		return err
	}

	log.Infof("Restore %s, check disk free space: %s, path: %s", s.RestoreId, usage.String(), s.Params.RootPath)

	if usage.UsedPercent > FreeLimit {
		log.Errorf("Restore %s, disk usage has reached %.2f%%", s.RestoreId, FreeLimit)
		return fmt.Errorf("Disk usage has reached %.2f%%. Please clean up disk space first.", FreeLimit)
	}

	requiredSpace := uint64(float64(s.RestoreType.TotalBytesProcessed) * 1.05)
	if usage.Free < requiredSpace {
		log.Errorf("not enough free space on target disk, required: %s, available: %s, location: %s", util.FormatBytes(requiredSpace), util.FormatBytes(usage.Free), s.Params.RootPath)
		return fmt.Errorf("Insufficient space on the target disk.")
	}

	return nil
}

func (s *StorageRestore) prepareForRun() error {
	var data = map[string]interface{}{
		"id":       s.Restore.Name,
		"type":     constant.MessageTypeRestore,
		"progress": 0,
		"status":   constant.Running.String(),
		"message":  "",
	}
	notification.DataSender.Send(s.RestoreType.Owner, data)

	return s.Handlers.GetRestoreHandler().UpdatePhase(s.Ctx, s.Restore.Name, constant.Running.String())
}

func (s *StorageRestore) progressCallback(percentDone float64) {

	select {
	case <-s.Ctx.Done():
		return
	default:
	}

	var percent = int(percentDone * progressDone)

	if time.Since(s.LastProgressTime) >= progressInterval*time.Second && s.LastProgressPercent != percent {
		s.LastProgressPercent = percent
		s.LastProgressTime = time.Now()

		s.Handlers.GetRestoreHandler().UpdateProgress(s.Ctx, s.RestoreId, percent)

		var data = map[string]interface{}{
			"id":       s.RestoreId,
			"type":     constant.MessageTypeRestore,
			"progress": percent,
			"status":   constant.Running.String(),
			"message":  "",
		}
		notification.DataSender.Send(s.RestoreType.Owner, data)
	}
}

func (s *StorageRestore) execute() (restoreOutput map[string]*backupssdkrestic.RestoreSummaryOutput, metadata string, totalBytes uint64, restoreError error) {
	var isSpaceRestore bool
	var logger = log.GetLogger()
	var resticSnapshotId = s.RestoreType.ResticSnapshotId
	var location = s.Params.Location["location"]
	var backupId = s.Params.BackupId
	var backupName = s.Params.BackupName

	log.Infof("Restore %s prepare: %s, backupType: %s, resticSnapshotId: %s, location: %s", s.RestoreId, s.RestoreType.Type, s.BackupType, resticSnapshotId, location)

	var restoreService *backupssdkstorage.RestoreService

	switch location {
	case constant.BackupLocationSpace.String():
		isSpaceRestore = true
		restoreOutput, metadata, totalBytes, restoreError = s.restoreFromSpace()
	case constant.BackupLocationAwsS3.String():
		token, err := s.getIntegrationCloud() // s3
		if err != nil {
			restoreError = fmt.Errorf("get %s token error: %v", token.Type, err)
			return
		}
		restoreService = backupssdk.NewRestoreService(&backupssdkstorage.RestoreOption{
			Password: s.Params.Password,
			Operator: constant.StorageOperatorApp,
			Ctx:      s.Ctx,
			Logger:   logger,
			Aws: &backupssdkoptions.AwsRestoreOption{
				RepoId:            backupId,
				RepoName:          backupName,
				SnapshotId:        resticSnapshotId,
				Path:              s.Params.Path,
				Endpoint:          s.RestoreType.Endpoint,
				AccessKey:         token.AccessKey,
				SecretAccessKey:   token.SecretKey,
				LimitDownloadRate: util.EnvOrDefault(constant.EnvLimitDownloadRate, ""),
			},
		})
	case constant.BackupLocationTencentCloud.String():
		token, err := s.getIntegrationCloud() // cos
		if err != nil {
			restoreError = fmt.Errorf("get %s token error: %v", token.Type, err)
			return
		}
		restoreService = backupssdk.NewRestoreService(&backupssdkstorage.RestoreOption{
			Password: s.Params.Password,
			Operator: constant.StorageOperatorApp,
			Ctx:      s.Ctx,
			Logger:   logger,
			TencentCloud: &backupssdkoptions.TencentCloudRestoreOption{
				RepoId:            backupId,
				RepoName:          backupName,
				SnapshotId:        resticSnapshotId,
				Path:              s.Params.Path,
				Endpoint:          s.RestoreType.Endpoint,
				AccessKey:         token.AccessKey,
				SecretAccessKey:   token.SecretKey,
				LimitDownloadRate: util.EnvOrDefault(constant.EnvLimitDownloadRate, ""),
			},
		})
	case constant.BackupLocationFileSystem.String():
		restoreService = backupssdk.NewRestoreService(&backupssdkstorage.RestoreOption{
			Password: s.Params.Password,
			Operator: constant.StorageOperatorApp,
			Ctx:      s.Ctx,
			Logger:   logger,
			Filesystem: &backupssdkoptions.FilesystemRestoreOption{
				RepoId:     backupId,
				RepoName:   backupName,
				SnapshotId: resticSnapshotId,
				Endpoint:   s.Params.Location["path"],
				Path:       s.Params.Path,
			},
		})
	}

	if !isSpaceRestore {
		restoreOutput, metadata, totalBytes, restoreError = restoreService.Restore(s.progressCallback)
	}

	return
}

func (s *StorageRestore) restoreFromSpace() (restoreOutput map[string]*backupssdkrestic.RestoreSummaryOutput, metadata string, totalBytes uint64, err error) {
	var backupId, backupName string
	var olaresSuffix string
	var owner = s.RestoreType.Owner
	if s.RestoreType.Type == constant.RestoreTypeSnapshot {
		backupId = s.Backup.Name
		backupName = s.Backup.Spec.Name
	} else {
		backupId = s.RestoreType.BackupId
		backupName = s.RestoreType.BackupName
		olaresSuffix = s.RestoreType.BackupUrl.OlaresSuffix
	}

	var spaceToken *integration.SpaceToken
	var resticSnapshotId = s.RestoreType.ResticSnapshotId
	var location = s.Params.Location

	for {
		spaceToken, err = integration.IntegrationManager().GetIntegrationSpaceToken(s.Ctx, owner, location["name"]) // restoreFromSpace
		if err != nil {
			err = fmt.Errorf("get space token error: %v", err)
			break
		}

		var spaceRestoreOption = &backupssdkoptions.SpaceRestoreOption{
			RepoId:            backupId,
			RepoName:          backupName,
			RepoSuffix:        olaresSuffix, // only used for backupUrl
			SnapshotId:        resticSnapshotId,
			Path:              s.Params.Path,
			OlaresDid:         spaceToken.OlaresDid,
			AccessToken:       spaceToken.AccessToken,
			ClusterId:         location["clusterId"],
			CloudName:         location["cloudName"],
			RegionId:          location["regionId"],
			CloudApiMirror:    constant.SyncServerURL,
			LimitDownloadRate: util.EnvOrDefault(constant.EnvLimitDownloadRate, ""),
		}

		var restoreService = backupssdk.NewRestoreService(&backupssdkstorage.RestoreOption{
			Password: s.Params.Password,
			Operator: constant.StorageOperatorApp,
			Ctx:      s.Ctx,
			Logger:   log.GetLogger(),
			Space:    spaceRestoreOption,
		})

		restoreOutput, metadata, totalBytes, err = restoreService.Restore(s.progressCallback)

		if err != nil {
			if strings.Contains(err.Error(), "refresh-token error") {
				spaceToken, err = integration.IntegrationManager().GetIntegrationSpaceToken(s.Ctx, owner, location["name"]) // restoreFromSpace
				if err != nil {
					err = fmt.Errorf("get space token error: %v", err)
					break
				}
				continue
			} else {
				break
			}
		}
		break
	}

	return
}

func (s *StorageRestore) updateRestoreResult(restoreOutput map[string]*backupssdkrestic.RestoreSummaryOutput, metadata string, totalBytes uint64, restoreError error) error {

	return wait.PollImmediate(time.Second*10, 5*time.Hour, func() (bool, error) {
		log.Infof("Restore %s, update restore result, if err: %v", s.RestoreId, restoreError)
		select {
		case <-s.Ctx.Done():
			return true, nil
		default:
			var msg, phase string
			var notifyProgress int

			restore, err := s.Handlers.GetRestoreHandler().GetRestore(s.Ctx, s.RestoreId)
			if err != nil {
				log.Errorf("Restore %s, get restore error: %v", s.RestoreId, err)
				if apierrors.IsNotFound(err) {
					return true, err
				}
				if util.ListMatchContains(ConnectErrors, err.Error()) {
					return false, nil
				}
			}

			if restoreError != nil {
				msg = restoreError.Error()
				phase = constant.Failed.String()
				notifyProgress = restore.Spec.Progress
				restore.Spec.Phase = pointer.String(constant.Failed.String())
				restore.Spec.Message = pointer.String(restoreError.Error())
				restore.Spec.ResticPhase = pointer.String(phase)
			} else {
				msg = constant.Completed.String()
				phase = constant.Completed.String()
				notifyProgress = progressDone
				restore.Spec.Size = pointer.UInt64Ptr(totalBytes)
				restore.Spec.Progress = progressDone
				restore.Spec.Phase = pointer.String(phase)
				restore.Spec.Message = pointer.String(phase)
				restore.Spec.ResticPhase = pointer.String(phase)
				restore.Spec.ResticMessage = pointer.String(util.ToJSON(restoreOutput))
			}

			extra := restore.Spec.Extra
			if extra == nil {
				extra = make(map[string]string)
			}

			if metadata != "" {
				extra["metadata"] = metadata
			}
			if s.RestoreAppStatus != nil {
				extra["restore_app_status"] = util.ToJSON(s.RestoreAppStatus)
			}

			restore.Spec.Extra = extra
			restore.Spec.EndAt = pointer.Time()

			var data = map[string]interface{}{
				"id":       s.Restore.Name,
				"type":     constant.MessageTypeRestore,
				"progress": notifyProgress,
				"endat":    restore.Spec.EndAt.Unix(),
				"status":   phase,
				"message":  msg,
			}

			notification.DataSender.Send(s.RestoreType.Owner, data)

			if err = s.Handlers.GetRestoreHandler().Update(s.Ctx, s.RestoreId, &restore.Spec); err != nil {
				return false, err
			}
			return true, nil
		}
	})
}

func (s *StorageRestore) getIntegrationCloud() (*integration.IntegrationToken, error) {
	var l = s.Params.Location
	var location = l["location"]
	var locationIntegrationName = l["name"]

	return integration.IntegrationManager().GetIntegrationCloudToken(s.Ctx, s.Restore.Spec.Owner, location, locationIntegrationName)
}

func (s *StorageRestore) removeRestoreFiles() {
	if s.BackupType == constant.BackupTypeApp {
		return
	}

	if err := os.RemoveAll(s.Params.Path); err != nil {
		log.Errorf("Restore %s, remove error: %v", s.RestoreId, err)
	}
}

func (s *StorageRestore) renameRestoreDirectory() {
	if s.BackupType == constant.BackupTypeApp {
		return
	}

	if err := os.Rename(s.Params.Path, strings.ReplaceAll(s.Params.Path, ".restore", "")); err != nil {
		log.Errorf("Restore %s, rename error: %v", s.RestoreId, err)
	}
}

func (s *StorageRestore) readyForRestoreApp(resticSnapshotId, metadata string) error {
	if s.BackupType != constant.BackupTypeApp {
		return nil
	}

	var ctx, cancel = context.WithTimeout(s.Ctx, 15*time.Second)
	defer cancel()

	var err error

	var appName = handlers.GetRestoreAppName(s.Restore)
	var appHandler = handlers.NewAppHandler(appName, s.Restore.Spec.Owner)

	err = appHandler.StartAppRestore(ctx, s.RestoreId, resticSnapshotId, metadata)
	if err != nil {
		log.Errorf("Restore %s, start app %s restore error: %v", s.RestoreId, appName, err)
		return err
	}

	log.Infof("Restore %s, waiting for check app %s restore status", s.RestoreId, appName)

	var ticker = time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			result, err := appHandler.GetAppRestoreStatus(s.Ctx, s.RestoreId)
			if err != nil {
				log.Errorf("Restore %s,%s, %v", s.RestoreId, appName, err)
				return err
			}

			log.Infof("Restore %s,%s, get app restore status, data: %s", s.RestoreId, appName, util.ToJSON(result))

			if result.Data.Status == constant.BackupAppStatusError {
				err = fmt.Errorf("app restore status is error, msg: %s", result.Data.Message)
				s.RestoreAppStatus = result
				return err
			}

			if result.Data.Status == constant.BackupAppStatusRunning {
				continue
			}

			if result.Data.Status == constant.BackupAppStatusFinish {
				s.RestoreAppStatus = result
				return nil
			}

		case <-ctx.Done():
			if e := ctx.Err(); e != nil {
				if errors.Is(err, context.DeadlineExceeded) {
					log.Errorf("Restore %s,%s, start app restore timeout", s.RestoreId, appName)
					return fmt.Errorf("app restore status timeout")
				}
				if errors.Is(err, context.Canceled) {
					log.Errorf("Restore %s,%s, app restore canceled", s.RestoreId, appName)
					return fmt.Errorf("restore canceled")
				}
			}
			return nil
		}
	}
}

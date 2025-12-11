package storage

import (
	"context"

	"olares.com/backup-server/pkg/constant"
	"olares.com/backup-server/pkg/handlers"
	"olares.com/backup-server/pkg/integration"
	"olares.com/backup-server/pkg/util/log"

	backupssdk "olares.com/backups-sdk"
	backupssdkoptions "olares.com/backups-sdk/pkg/options"
	backupssdkrestic "olares.com/backups-sdk/pkg/restic"
	backupssdkstorage "olares.com/backups-sdk/pkg/storage"
)

type StorageSnapshots struct {
	Handlers handlers.Interface
}

func (s *StorageSnapshots) GetSnapshot(ctx context.Context, password, owner, location, endpoint, backupName, backupId, snapshotId, cloudName, regionId, clusterId string) (*backupssdkrestic.Snapshot, error) {
	var snapshotService *backupssdkstorage.SnapshotsService
	var logger = log.GetLogger()

	switch location {
	case constant.BackupLocationSpace.String():
		return s.getSpaceSnapshot(ctx, owner, password, backupId, backupName, cloudName, regionId, snapshotId, clusterId)
	case constant.BackupLocationAwsS3.String():
		token, err := s.getIntegrationCloud(ctx, owner, location, endpoint)
		if err != nil {
			return nil, err
		}

		snapshotService = backupssdk.NewSnapshotsService(&backupssdkstorage.SnapshotsOption{
			Password:   password,
			Operator:   constant.StorageOperatorApp,
			SnapshotId: snapshotId,
			Logger:     logger,
			Aws: &backupssdkoptions.AwsSnapshotsOption{
				RepoId:          backupId,
				RepoName:        backupName,
				Endpoint:        endpoint,
				AccessKey:       token.AccessKey,
				SecretAccessKey: token.SecretKey,
			},
		})
	case constant.BackupLocationTencentCloud.String():
		token, err := s.getIntegrationCloud(ctx, owner, location, endpoint)
		if err != nil {
			return nil, err
		}
		snapshotService = backupssdk.NewSnapshotsService(&backupssdkstorage.SnapshotsOption{
			Password:   password,
			Operator:   constant.StorageOperatorApp,
			SnapshotId: snapshotId,
			Logger:     log.GetLogger(),
			TencentCloud: &backupssdkoptions.TencentCloudSnapshotsOption{
				RepoId:          backupId,
				RepoName:        backupName,
				Endpoint:        endpoint,
				AccessKey:       token.AccessKey,
				SecretAccessKey: token.SecretKey,
			},
		})
	case constant.BackupLocationFileSystem.String():
		snapshotService = backupssdk.NewSnapshotsService(&backupssdkstorage.SnapshotsOption{
			Password:   password,
			Operator:   constant.StorageOperatorApp,
			SnapshotId: snapshotId,
			Logger:     logger,
			Filesystem: &backupssdkoptions.FilesystemSnapshotsOption{
				RepoId:   backupId,
				RepoName: backupName,
				Endpoint: endpoint,
			},
		})
	}

	res, err := snapshotService.Snapshots()

	if err != nil {
		return nil, err
	}

	return res.First(), nil
}

func (s *StorageSnapshots) getSpaceSnapshot(ctx context.Context, owner, password, backupId, backupName, cloudName, regionId, snapshotId, clusterId string) (*backupssdkrestic.Snapshot, error) {
	account, err := s.Handlers.GetSnapshotHandler().GetOlaresId(owner)
	if err != nil {
		return nil, err
	}

	spaceToken, err := integration.IntegrationManager().GetIntegrationSpaceToken(ctx, owner, account)
	if err != nil {
		return nil, err
	}

	var snapshotService = backupssdk.NewSnapshotsService(&backupssdkstorage.SnapshotsOption{
		Password:   password,
		Operator:   constant.StorageOperatorApp,
		SnapshotId: snapshotId,
		Logger:     log.GetLogger(),
		Space: &backupssdkoptions.SpaceSnapshotsOption{
			RepoId:         backupId,
			RepoName:       backupName,
			OlaresDid:      spaceToken.OlaresDid,
			AccessToken:    spaceToken.AccessToken,
			ClusterId:      clusterId,
			CloudName:      cloudName,
			RegionId:       regionId,
			CloudApiMirror: constant.SyncServerURL,
		},
	})

	res, err := snapshotService.Snapshots()
	if err != nil {
		return nil, err
	}

	return res.First(), nil
}

func (s *StorageSnapshots) GetSnapshots(ctx context.Context, password, owner, location, endpoint, backupName, backupId string) (*backupssdkrestic.SnapshotList, error) {
	var snapshotService *backupssdkstorage.SnapshotsService
	var logger = log.GetLogger()

	switch location {
	case constant.BackupLocationAwsS3.String():
		token, err := s.getIntegrationCloud(ctx, owner, location, endpoint)
		if err != nil {
			return nil, err
		}

		snapshotService = backupssdk.NewSnapshotsService(&backupssdkstorage.SnapshotsOption{
			Password: password,
			Logger:   logger,
			Operator: constant.StorageOperatorApp,
			Aws: &backupssdkoptions.AwsSnapshotsOption{
				RepoId:          backupId,
				RepoName:        backupName,
				Endpoint:        endpoint,
				AccessKey:       token.AccessKey,
				SecretAccessKey: token.SecretKey,
			},
		})
	case constant.BackupLocationTencentCloud.String():
		token, err := s.getIntegrationCloud(ctx, owner, location, endpoint)
		if err != nil {
			return nil, err
		}
		snapshotService = backupssdk.NewSnapshotsService(&backupssdkstorage.SnapshotsOption{
			Password: password,
			Logger:   log.GetLogger(),
			Operator: constant.StorageOperatorApp,
			TencentCloud: &backupssdkoptions.TencentCloudSnapshotsOption{
				RepoId:          backupId,
				RepoName:        backupName,
				Endpoint:        endpoint,
				AccessKey:       token.AccessKey,
				SecretAccessKey: token.SecretKey,
			},
		})
	case constant.BackupLocationFileSystem.String():
		snapshotService = backupssdk.NewSnapshotsService(&backupssdkstorage.SnapshotsOption{
			Password: password,
			Logger:   logger,
			Operator: constant.StorageOperatorApp,
			Filesystem: &backupssdkoptions.FilesystemSnapshotsOption{
				RepoId:   backupId,
				RepoName: backupName,
				Endpoint: endpoint,
			},
		})
	}

	return snapshotService.Snapshots()
}

func (s *StorageSnapshots) getIntegrationCloud(ctx context.Context, owner, location, endpoint string) (*integration.IntegrationToken, error) {
	return integration.IntegrationManager().GetIntegrationCloudAccount(ctx, owner, location, endpoint)
}

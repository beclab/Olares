package storage

import (
	"context"
	"fmt"

	"olares.com/backup-server/pkg/constant"
	"olares.com/backup-server/pkg/handlers"
	integration "olares.com/backup-server/pkg/integration"
	"olares.com/backup-server/pkg/util/log"

	backupssdk "olares.com/backups-sdk"
	backupssdkoptions "olares.com/backups-sdk/pkg/options"
	backupssdkstorage "olares.com/backups-sdk/pkg/storage"
)

type StorageRegion struct {
	Handlers handlers.Interface
}

func (s *StorageRegion) GetRegions(ctx context.Context, owner, olaresId string) ([]map[string]string, error) {
	var spaceToken, err = integration.IntegrationManager().GetIntegrationSpaceToken(ctx, owner, olaresId) // only for Space
	if err != nil {
		err = fmt.Errorf("get space token error: %v", err)
		return nil, err
	}

	var spaceRegionOption = &backupssdkoptions.SpaceRegionOptions{
		OlaresDid:      spaceToken.OlaresDid,
		AccessToken:    spaceToken.AccessToken,
		CloudApiMirror: constant.SyncServerURL,
	}

	var regionService = backupssdk.NewRegionService(&backupssdkstorage.RegionOption{
		Ctx:    context.Background(),
		Logger: log.GetLogger(),
		Space:  spaceRegionOption,
	})

	return regionService.Regions()
}

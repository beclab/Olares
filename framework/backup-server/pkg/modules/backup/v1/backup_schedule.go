package v1

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
	sysv1 "olares.com/backup-server/pkg/apis/sys.bytetrade.io/v1"
	"olares.com/backup-server/pkg/client"
	"olares.com/backup-server/pkg/constant"
	"olares.com/backup-server/pkg/handlers"
	"olares.com/backup-server/pkg/integration"
	"olares.com/backup-server/pkg/notify"
	"olares.com/backup-server/pkg/util"
	"olares.com/backup-server/pkg/util/log"
	"olares.com/backup-server/pkg/util/pointer"
)

type BackupPlan struct {
	owner    string
	endpoint string
	c        *BackupCreate
	factory  client.Factory
	handler  handlers.Interface
}

func NewBackupPlan(owner string, factory client.Factory, handler handlers.Interface) *BackupPlan {
	return &BackupPlan{
		owner:   owner,
		factory: factory,
		handler: handler,
	}
}

func (o *BackupPlan) Apply(ctx context.Context, c *BackupCreate) (*sysv1.Backup, error) {
	var err error
	o.c = c

	if err = o.validate(ctx); err != nil {
		return nil, errors.WithStack(err)
	}

	if err = o.verifyUsage(); err != nil {
		return nil, err
	}

	return o.apply(ctx)
}

func (o *BackupPlan) Update(ctx context.Context, c *BackupCreate, backup *sysv1.Backup) error {
	var err error
	o.c = c

	if err = o.validBackupPolicy(); err != nil { // update
		return errors.WithStack(err)
	}

	if err = o.update(ctx, backup); err != nil { // update backup plan

	}
	return nil
}

func (o *BackupPlan) validate(ctx context.Context) error {
	if o.c.Name == "" {
		return errors.New("name is required")
	}
	if o.owner == "" {
		return errors.New("owner is required")
	}

	if err := o.validLocation(); err != nil {
		return err
	}

	if err := o.validIntegration(ctx); err != nil {
		return err
	}

	if err := o.validBackupPolicy(); err != nil {
		return err
	}

	return nil
}

func (o *BackupPlan) mergeConfig(clusterId string) *sysv1.BackupSpec {
	var createAt = pointer.Time()
	var name = strings.TrimSpace(o.c.Name)
	var backupType = getBackupType(o.c)
	var backupTypeData = make(map[string]string)
	if isBackupApp(backupType) {
		backupTypeData[constant.BackupTypeApp] = o.buildBackupAppType()
	} else {
		backupTypeData[constant.BackupTypeFile] = o.buildBackupType() // ! trim path prefix, like '/Files'
	}

	bc := &sysv1.BackupSpec{
		Name:       name,
		Owner:      o.owner,
		BackupType: backupTypeData,
		Notified:   false,
		Size:       pointer.UInt64Ptr(0),
		CreateAt:   createAt,
		Extra: map[string]string{
			"size_updated": fmt.Sprintf("%d", createAt.Unix()),
		},
	}

	if o.c.Location != "" && o.c.LocationConfig != nil {
		var locationName = o.c.Location
		var location = make(map[string]string)
		// {"space": "..."}
		location[locationName] = o.buildLocationConfig(o.c.Location, clusterId, o.c.LocationConfig)
		bc.Location = location
	}

	if o.c.BackupPolicies != nil {
		bc.BackupPolicy = o.c.BackupPolicies
		bc.BackupPolicy.Enabled = true
	}
	return bc
}

func (o *BackupPlan) update(ctx context.Context, backup *sysv1.Backup) error {
	backup.Spec.BackupPolicy = o.c.BackupPolicies

	return o.handler.GetBackupHandler().UpdateBackupPolicy(ctx, backup)
}

func (o *BackupPlan) apply(ctx context.Context) (*sysv1.Backup, error) {
	var (
		backupSpec *sysv1.BackupSpec
	)

	clusterId, err := o.getClusterId(ctx)
	if err != nil {
		return nil, errors.WithStack(fmt.Errorf("get cluster id error: %v", err))
	}

	backupSpec = o.mergeConfig(clusterId)
	if o.c != nil {
		// todo update
		// if o.c.Location != "" && configSpec.Location != o.c.Location {
		// 	return errors.New("change location is not allowed")
		// }
	}

	backupType := getBackupType(o.c)
	backupAppTypeName := ""
	if o.c.BackupCreateType != nil {
		backupAppTypeName = o.c.BackupCreateType.Name
	}

	log.Infof("merged backup spec: %s", util.ToJSON(backupSpec))

	backup, err := o.handler.GetBackupHandler().Create(ctx, o.owner, o.c.Name, o.c.Path, backupType, backupAppTypeName, backupSpec)
	if err != nil {
		return nil, err
	}

	log.Infof("create backup %s, id %s", backup.Spec.Name, backup.Name)

	return backup, nil
}

func (o *BackupPlan) validLocation() error {
	log.Infof("new backup %s location %s", o.c.Name, util.ToJSON(o.c.LocationConfig))

	location := o.c.Location
	locationConfig := o.c.LocationConfig

	if ok := util.ListContains([]string{
		constant.BackupLocationSpace.String(),
		constant.BackupLocationAwsS3.String(),
		constant.BackupLocationTencentCloud.String(),
		constant.BackupLocationFileSystem.String(),
	}, location); !ok {
		return errors.Errorf("backup %s location %s not support", o.c.Name, location)
	}

	if location == constant.BackupLocationSpace.String() {
		if locationConfig.CloudName == "" || locationConfig.RegionId == "" {
			return errors.Errorf("backup %s location space invalid, cloudName: %s, regionId: %s", o.c.Name, locationConfig.CloudName, locationConfig.RegionId)
		}
	} else if location == constant.BackupLocationAwsS3.String() || location == constant.BackupLocationTencentCloud.String() {
		if locationConfig.Name == "" {
			return errors.Errorf("backup %s location %s invalid, please check name", o.c.Name, location)
		}
	} else if location == constant.BackupLocationFileSystem.String() {
		if locationConfig.Path == "" {
			return errors.Errorf("backup %s location %s path %s invalid, please check target path", o.c.Name, location, locationConfig.Path)
		}

		if o.c.BackupCreateType == nil || o.c.BackupCreateType.Type == constant.BackupTypeFile {
			if strings.HasPrefix(locationConfig.Path, o.c.Path) {
				return errors.Errorf("the path %s to be backed up contains the backup storage path %s", o.c.Path, locationConfig.Path)
			}
		}
	}

	return nil
}

func (o *BackupPlan) validIntegration(ctx context.Context) error {
	var owner = o.owner
	var location = o.c.Location
	var locationConfig = o.c.LocationConfig

	if location == constant.BackupLocationFileSystem.String() {
		return nil
	}
	integrationName, err := integration.IntegrationManager().ValidIntegrationNameByLocationName(ctx, owner, location, locationConfig.Name)
	if err != nil {
		return errors.WithStack(err)
	}

	log.Infof("backup %s location %s integration %s", o.c.Name, location, integrationName)

	if location == constant.BackupLocationSpace.String() {
		o.c.LocationConfig.Name = integrationName
		return nil
	}

	if location == constant.BackupLocationAwsS3.String() || location == constant.BackupLocationTencentCloud.String() {
		integrationToken, err := integration.IntegrationManager().GetIntegrationCloudToken(ctx, owner, location, integrationName)
		if err != nil {
			log.Errorf("integration: %s, location: %s, token get error: %v", integrationName, location, err)
			return errors.WithStack(err)
		}
		log.Infof("backup %s location: %s, integration: %s, endpoint: %s", o.c.Name, location, integrationName, integrationToken.Endpoint)
		o.endpoint = integrationToken.Endpoint
	}

	return nil
}

func (o *BackupPlan) verifyUsage() error {
	var location = o.c.Location
	var accountName = o.c.LocationConfig.Name

	if location != constant.BackupLocationSpace.String() {
		return nil
	}

	spaceToken, err := integration.IntegrationManager().GetIntegrationSpaceToken(context.Background(), o.owner, accountName)
	if err != nil {
		return errors.New(constant.MessageTokenExpired)
	}

	spaceUsage, err := notify.CheckCloudStorageQuotaAndPermission(context.Background(), constant.SyncServerURL, spaceToken.OlaresDid, spaceToken.AccessToken)
	if err != nil {
		return err
	}

	if spaceUsage.Data.PlanLevel == constant.FreeUser {
		return errors.New(constant.MessagePlanLevelFreeUser)
	}

	if !spaceUsage.Data.CanBackup {
		return errors.New(constant.MessagePlanLevelBackupSpaceForbidden)
	}

	return nil

}

func (o *BackupPlan) validBackupPolicy() error {
	log.Infof("backup %s location %s", o.c.Name, util.ToJSON(o.c.BackupPolicies))

	var timespanOfDay = o.c.BackupPolicies.TimesOfDay
	o.c.BackupPolicies.TimespanOfDay = timespanOfDay
	o.c.BackupPolicies.Enabled = true
	policy := o.c.BackupPolicies

	if ok := util.ListContains([]string{
		constant.BackupSnapshotFrequencyHourly.String(),
		constant.BackupSnapshotFrequencyDaily.String(),
		constant.BackupSnapshotFrequencyWeekly.String(),
		constant.BackupSnapshotFrequencyMonthly.String(),
	}, policy.SnapshotFrequency); !ok {
		return errors.Errorf("backup %s snapshot frequency %s not support", o.c.Name, policy.SnapshotFrequency)
	}

	if !strings.Contains(o.c.BackupPolicies.TimesOfDay, ":") {
		timeInUTC, err := util.ParseTimestampToLocal(o.c.BackupPolicies.TimesOfDay)
		if err != nil {
			return errors.Errorf("backup %s snapshot times of day invalid, eg: '48600000'", o.c.Name)
		}
		o.c.BackupPolicies.TimesOfDay = timeInUTC
	} else {
		timeSplit := strings.Split(o.c.BackupPolicies.TimesOfDay, ":")
		if !strings.Contains(o.c.BackupPolicies.TimesOfDay, ":") || len(timeSplit) != 2 {
			return errors.New("invalid times of day format, eg: '07:30'")
		}
	}

	if policy.SnapshotFrequency == constant.BackupSnapshotFrequencyWeekly.String() {
		if policy.DayOfWeek < 1 || policy.DayOfWeek > 7 {
			return errors.Errorf("backup %s day of week invalid, eg: '1', '2'...'7'", o.c.Name)
		}
	}

	if policy.SnapshotFrequency == constant.BackupSnapshotFrequencyMonthly.String() {
		if policy.DateOfMonth < 1 || policy.DateOfMonth > 31 {
			return errors.Errorf("backup %s day of month invalid, eg: '1', '2'...'31'", o.c.Name)
		}
	}

	return nil

}

func (o *BackupPlan) getClusterId(ctx context.Context) (string, error) {
	var clusterId string
	factory, err := client.NewFactory()
	if err != nil {
		return clusterId, errors.WithStack(err)
	}

	dynamicClient, err := factory.DynamicClient()
	if err != nil {
		return clusterId, errors.WithStack(err)
	}

	var backoff = wait.Backoff{
		Duration: 2 * time.Second,
		Factor:   2,
		Jitter:   0.1,
		Steps:    5,
	}

	if err := retry.OnError(backoff, func(err error) bool {
		return true
	}, func() error {
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		unstructuredUser, err := dynamicClient.Resource(constant.OlaresGVR).Get(ctx, constant.OlaresName, metav1.GetOptions{})
		if err != nil {
			return errors.WithStack(err)
		}
		obj := unstructuredUser.UnstructuredContent()
		clusterId, _, err = unstructured.NestedString(obj, "metadata", "labels", "bytetrade.io/cluster-id")
		if err != nil {
			return errors.WithStack(err)
		}
		if clusterId == "" {
			return errors.WithStack(fmt.Errorf("cluster id not found"))
		}
		return nil
	}); err != nil {
		return clusterId, errors.WithStack(err)
	}

	return clusterId, nil
}

func (o *BackupPlan) validPassword() error {
	if !(o.c.Password == o.c.ConfirmPassword && o.c.Password != "") {
		return errors.Errorf("password not match")
	}
	return nil
}

func (o *BackupPlan) buildBackupAppType() string {
	var backupType = make(map[string]string)
	backupType["name"] = o.c.BackupCreateType.Name
	return util.ToJSON(backupType)
}

func (o *BackupPlan) buildBackupType() string {
	var backupType = make(map[string]string)
	backupType["path"] = o.c.Path
	return util.ToJSON(backupType)
}

func (o *BackupPlan) buildLocationConfig(location string, clusterId string, config *LocationConfig) string {
	var data = make(map[string]string)

	switch location {
	case constant.BackupLocationFileSystem.String():
		data["path"] = config.Path
	case constant.BackupLocationSpace.String():
		data["cloudName"] = config.CloudName
		data["regionId"] = config.RegionId
		data["clusterId"] = clusterId
	case constant.BackupLocationAwsS3.String(),
		constant.BackupLocationTencentCloud.String():
		data["endpoint"] = o.endpoint
	}

	data["name"] = config.Name

	return util.ToJSON(data)
}

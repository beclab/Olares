package constant

import (
	"net/url"
	"os"

	"k8s.io/apimachinery/pkg/runtime/schema"
	scheme "olares.com/backup-server/pkg/apis/sys.bytetrade.io/v1"
)

const (
	EnvOlaresSystemRemoteService = "OLARES_SYSTEM_REMOTE_SERVICE"
	DefaultSyncServerURL         = "https://api.olares.com"
)

var (
	SyncServerURL       string
	OlaresRemoteService string = DefaultSyncServerURL
	NodeName            string = os.Getenv("NODE_NAME")
)

const (
	DefaultSnapshotSizeUnit = "byte"
	BflUserKey              = "X-BFL-USER"

	DefaultIntegrationProviderUrl = "http://integration-provider-svc.os-protected:28080"
	DefaultNamespaceOsFramework   = "os-framework"
	DefaultServiceAccount         = "backup-cluster-view"
	DefaultServiceAccountSettings = "user-backend"

	FullyBackup         string = "fully"
	IncrementalBackup   string = "incremental"
	UnKnownBackup       string = "unknown"
	FullyBackupId       int    = 0
	IncrementalBackupId int    = 1
	UnKnownBackupId     int    = 2

	KindSnapshot string = "Snapshot"
	KindBackup   string = "Backup"
	KindRestore  string = "Restore"

	SnapshotController string = "snapshot-controller"
	BackupController   string = "backup-controller"
	RestoreController  string = "restore-controller"

	BackupTypeFile string = "file"
	BackupTypeApp  string = "app"

	BackupPause  string = "pause"
	BackupResume string = "resume"
	BackupCancel string = "cancel"
	BackupDelete string = "delete"

	BackupAppStatusPreparing string = "preparing"
	BackupAppStatusRunning   string = "running"
	BackupAppStatusFinish    string = "finish"
	BackupAppStatusError     string = "err"

	RestoreTypeSnapshot string = "snapshot"
	RestoreTypeUrl      string = "url"

	StorageOperatorCli string = "cli"
	StorageOperatorApp string = "app"

	EventBackup  string = "backup-state-event"
	EventRestore string = "restore-state-event"

	TraceId string = "traceId"

	CachePath    string = "/appcache"
	ExternalPath string = "/shares"

	DefaultStoragePrefix = "olares-backups/"

	EnvLimitUploadRate   = "LIMIT_UPLOAD_RATE"
	EnvLimitDownloadRate = "LIMIT_DOWNLOAD_RATE"
	EnvIntegrationDebug  = "BACKUP_INTEGRATION_DEBUG"

	FreeSpaceStorage uint64 = 2 * 1024 * 1024 * 1024

	MessageTypeBackup  = "backup"
	MessageTypeRestore = "restore"
)

const (
	MessageBackupServerRestart           = "Backup task terminated due to service restart."
	MessageTaskCanceled                  = "Task canceled."
	MessageTaskExists                    = "A snapshot task is currently running or queued."
	MessageTokenExpired                  = "Access token expired. Please re-connect to your Olares Space in LarePass."
	MessagePlanLevelFreeUser             = "Your current plan on Olares Space doesn't include the backup service. Upgrade now to enable it. Alternatively, you can connect and back up to your own AWS S3, Tencent COS, or a local disk."
	MessagePlanLevelBackupSpaceForbidden = "Insufficient storage in Olares Space."
	MessageBackupUrlIncorrect            = "Unable to locate the configuration file. Please provide the correct storage path."
)

type BackupLocation string

func (b BackupLocation) String() string {
	return string(b)
}

const (
	BackupLocationSpace        BackupLocation = "space"
	BackupLocationAwsS3        BackupLocation = "awss3"
	BackupLocationTencentCloud BackupLocation = "tencentcloud"
	BackupLocationFileSystem   BackupLocation = "filesystem"

	BackupLocationSpaceAlias      BackupLocation = "Olares Space"
	BackupLocationAwsS3Alias      BackupLocation = "AWS S3"
	BackupLocationCosAlias        BackupLocation = "TencentCloud COS"
	BackupLocationFileSystemAlias BackupLocation = "Local"
	BackupLocationUnKnownAlias    BackupLocation = "Unknown"

	LocationTypeSpaceTag        = "did:key:"
	LocationTypeAwsS3Tag        = "amazonaws.com"
	LocationTypeTencentCloudTag = "myqcloud.com"
)

type BackupSnapshotFrequency string

func (b BackupSnapshotFrequency) String() string {
	return string(b)
}

const (
	BackupSnapshotFrequencyHourly  BackupSnapshotFrequency = "@hourly"
	BackupSnapshotFrequencyDaily   BackupSnapshotFrequency = "@daily"
	BackupSnapshotFrequencyWeekly  BackupSnapshotFrequency = "@weekly"
	BackupSnapshotFrequencyMonthly BackupSnapshotFrequency = "@monthly"
)

var (
	BackendTokenHeader         = "Terminus-Nonce"
	BackendAuthorizationHeader = "Authorization"
	OlaresName                 = "terminus"

	BackupGVR = schema.GroupVersionResource{
		Group:    scheme.SchemeGroupVersion.Group,
		Version:  scheme.SchemeGroupVersion.Version,
		Resource: "backups",
	}

	SnapshotGVR = schema.GroupVersionResource{
		Group:    scheme.SchemeGroupVersion.Group,
		Version:  scheme.SchemeGroupVersion.Version,
		Resource: "snapshots",
	}

	RestoreGVR = schema.GroupVersionResource{
		Group:    scheme.SchemeGroupVersion.Group,
		Version:  scheme.SchemeGroupVersion.Version,
		Resource: "restores",
	}

	OlaresGVR = schema.GroupVersionResource{
		Group:    "sys.bytetrade.io",
		Version:  "v1alpha1",
		Resource: OlaresName,
	}

	UsersGVR = schema.GroupVersionResource{
		Group:    "iam.kubesphere.io",
		Version:  "v1alpha2",
		Resource: "users",
	}
)

type Phase string

func (s Phase) String() string {
	return string(s)
}

const (
	New       Phase = "New"
	Pending   Phase = "Pending"
	Running   Phase = "Running"
	Failed    Phase = "Failed"
	Completed Phase = "Completed"
	Canceled  Phase = "Canceled"
	Rejected  Phase = "Rejected"
	Deleted   Phase = "Deleted"
)

const (
	MaxConcurrency = 1
	NonBlocking    = true
)

const (
	FreeUser    = 1
	MemeberUser = 2
)

func ReloadEnvDependantVars() error {
	var err error
	SyncServerURL, err = url.JoinPath(OlaresRemoteService, "space")
	if err != nil {
		return err
	}
	return nil
}

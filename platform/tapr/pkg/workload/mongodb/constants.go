package mongodb

import (
	"bytetrade.io/web3os/tapr/pkg/constants"
	psmdbv1 "github.com/percona/percona-server-mongodb-operator/pkg/apis/psmdb/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	PerconaMongoCluster                  = "mongo-cluster"
	PerconaMongoProxy                    = "mongos"
	PerconaMongoService                  = PerconaMongoCluster + "-" + PerconaMongoProxy
	PerconaMongoClusterBackup            = "mongocluster-backup"
	PerconaMongoClusterLastBackupPBMName = "percona/psmdb-last-backup-pbmname"
)

var (
	PSMDBClassGVR = schema.GroupVersionResource{
		Group:    psmdbv1.SchemeGroupVersion.Group,
		Version:  psmdbv1.SchemeGroupVersion.Version,
		Resource: "perconaservermongodbs",
	}

	PSMDBBackupClassGVR = schema.GroupVersionResource{
		Group:    psmdbv1.SchemeGroupVersion.Group,
		Version:  psmdbv1.SchemeGroupVersion.Version,
		Resource: "perconaservermongodbbackups",
	}

	PSMDBRestoreClassGVR = schema.GroupVersionResource{
		Group:    psmdbv1.SchemeGroupVersion.Group,
		Version:  psmdbv1.SchemeGroupVersion.Version,
		Resource: "perconaservermongodbrestores",
	}
)

const (
	PSMDB_NAME               = "mongo-cluster"
	PSMDB_NAMESPACE          = constants.PlatformNamespace
	PSMDB_SECRET             = "mdb-cluster-name-secrets"
	PSMDB_ADMIN_KEY          = "MONGODB_DATABASE_ADMIN_USER"
	PSMDB_ADMIN_PASSWORD_KEY = "MONGODB_DATABASE_ADMIN_PASSWORD"
	PSMDB_SVC                = "mongo-cluster-mongos"
	PSMDB_RESTORE_NAME       = "mongocluster-restore"
)

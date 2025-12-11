package mongodb

import (
	"github.com/percona/percona-backup-mongodb/pbm"
	psmdbv1 "github.com/percona/percona-server-mongodb-operator/pkg/apis/psmdb/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	ClusterBackup = psmdbv1.PerconaServerMongoDBBackup{
		TypeMeta: metav1.TypeMeta{
			APIVersion: psmdbv1.SchemeGroupVersion.String(),
			Kind:       "PerconaServerMongoDBBackup",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:       PerconaMongoClusterBackup,
			Finalizers: []string{"delete-backup"},
		},

		Spec: psmdbv1.PerconaServerMongoDBBackupSpec{
			ClusterName: PerconaMongoCluster,
			StorageName: "s3-local",
			Type:        pbm.PhysicalBackup,
		},
	}

	ClusterRestore = psmdbv1.PerconaServerMongoDBRestore{
		TypeMeta: metav1.TypeMeta{
			APIVersion: psmdbv1.SchemeGroupVersion.String(),
			Kind:       "PerconaServerMongoDBRestore",
		},

		ObjectMeta: metav1.ObjectMeta{
			Name: "mongocluster-restore",
		},

		Spec: psmdbv1.PerconaServerMongoDBRestoreSpec{
			ClusterName: PerconaMongoCluster,
			BackupName:  PerconaMongoClusterBackup,
		},
	}
)

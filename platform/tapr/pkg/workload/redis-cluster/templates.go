package rediscluster

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	store "kmodules.xyz/objectstore-api/api/v1"
)

const (
	AuthSectionBegin = `
	Authority {
`
	AuthUserBegin = `
		Auth %s {
`
	AuthUserModeWrite = `
			Mode write
`
	AuthUserNamespace = `
			Namespace %s
`
	AuthUserEnd = `
		}
`
	AuthSectionEnd = `
	}	
`
	AuthSectionExprPrev = "################################### AUTHORITY ##################################"
	AuthSectionExprNext = "################################### SERVERS ####################################"
	AuthSectionExpr     = AuthSectionExprPrev + "([^#]+)" + AuthSectionExprNext
)

var (
	hostPathType corev1.HostPathType = corev1.HostPathDirectoryOrCreate

	ClusterBackup = RedisClusterBackup{
		TypeMeta: metav1.TypeMeta{
			APIVersion: SchemeGroupVersion.String(),
			Kind:       "RedisClusterBackup",
		},

		ObjectMeta: metav1.ObjectMeta{
			Name: "rediscluster-backup",
			Annotations: map[string]string{
				"redis.kun/scope": "cluster-scoped",
			},
		},

		Spec: RedisClusterBackupSpec{
			Image:            "beclab/redis-backup-tools:redis-6.2.13",
			RedisClusterName: "redis-cluster",
			Backend: store.Backend{
				Local: &store.LocalSpec{
					MountPath: "/back",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Type: &hostPathType,
							Path: "",
						},
					},
				},
			},
		},
	}
)

package citus

import (
	"time"

	"bytetrade.io/web3os/tapr/pkg/apis/apr/v1alpha1"
	"bytetrade.io/web3os/tapr/pkg/constants"

	appv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	CitusAdminSecretName     = "pgcluster-admin"
	CitusVolumeName          = "pgdata"
	CitusBackupVolumeName    = "pgbackup"
	CitusHeadlessServiceName = "citus-headless"
	CitusMasterServiceName   = "citus-master-svc"
	PGClusterName            = "citus"
	PGClusterNamespace       = constants.PlatformNamespace
	DefaultPGAdminUser       = "olares"
	PGClusterBackup          = "citus-backup"
	PGClusterRestore         = "citus-restore"
)

var (
	PGNodeHBATrust = `
local   all             all                                     trust
host    all             all             127.0.0.1/32            trust
host    all             all             ::1/128                 trust
local   replication     all                                     trust
host    replication     all             127.0.0.1/32            trust
host    replication     all             ::1/128                 trust
`
	PGNodeHBAScram = `
host all all all scram-sha-256
`
)

var (
	DefaultReplicas    int32 = 1
	PGDataHostPathType       = corev1.HostPathDirectoryOrCreate
	CitusImage               = "beclab/citus:11.3"
)

var (

	// sts template
	CitusStatefulset = appv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: PGClusterName,
			Labels: map[string]string{
				"managed-by": "citus-operator",
			},
		},
		Spec: appv1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":                    "citus",
					"app.kubernetes.io/name": "citus",
				},
			},
			ServiceName: CitusHeadlessServiceName,
			Replicas:    &DefaultReplicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":                         "citus",
						"app.kubernetes.io/name":      "citus",
						"app.bytetrade.io/middleware": "true",
					},
				},
				Spec: corev1.PodSpec{
					Affinity: &corev1.Affinity{
						NodeAffinity: &corev1.NodeAffinity{
							PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{
								{
									Preference: corev1.NodeSelectorTerm{
										MatchExpressions: []corev1.NodeSelectorRequirement{
											{
												Key:      "kubernetes.io/os",
												Operator: "In",
												Values:   []string{"linux"},
											},
											{
												Key:      "node-role.kubernetes.io/master",
												Operator: "Exists",
											},
										},
									},
									Weight: 10,
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:            "postgres",
							Image:           CitusImage,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Args:            []string{"-N", "1000"},
							Env: []corev1.EnvVar{
								{
									Name: "POD_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
								{
									Name: "POSTGRES_PASSWORD",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{Name: CitusAdminSecretName},
											Key:                  "password",
										},
									},
								},
								{
									Name: "POSTGRES_USER",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{Name: CitusAdminSecretName},
											Key:                  "user",
										},
									},
								},
								{
									Name:  "PGDATA",
									Value: "/var/lib/postgresql/data/$(POD_NAME)",
								},
							}, // env
							Ports: []corev1.ContainerPort{
								{
									Name:          "postgres",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: 5432,
								},
							},
							LivenessProbe: &corev1.Probe{
								InitialDelaySeconds: 10,
								ProbeHandler: corev1.ProbeHandler{
									TCPSocket: &corev1.TCPSocketAction{Port: intstr.FromInt(5432)},
								},
							},
							// StartupProbe: &corev1.Probe{
							// 	InitialDelaySeconds: 10,
							// 	ProbeHandler: corev1.ProbeHandler{
							// 		Exec: &corev1.ExecAction{
							// 			Command: []string{
							// 				"bash", "-c", "[ -S /var/run/postgresql/.s.PGSQL.5432 ]",
							// 			},
							// 		},
							// 	},
							// },
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      CitusVolumeName,
									MountPath: "/var/lib/postgresql/data",
								},
								{
									Name:      CitusBackupVolumeName,
									MountPath: "/backup",
								},
							},
						}, // container postgres
					}, // containers
					Volumes: []corev1.Volume{
						{
							Name: CitusVolumeName,
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Type: &PGDataHostPathType,
								},
							},
						},
						{
							Name: CitusBackupVolumeName,
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Type: &PGDataHostPathType,
								},
							},
						},
					},
				}, // pod spec
			}, // pod template
		}, // sts spec
	}

	// admin user secret template
	CitusAdminSecret = corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: CitusAdminSecretName,
		},
		StringData: map[string]string{
			"user":     "",
			"password": "",
		},
	}

	// headless service template
	CitusHeadlessService = corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: CitusHeadlessServiceName,
		},

		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app":                    "citus",
				"app.kubernetes.io/name": "citus",
			},

			ClusterIP: "None",
			Ports: []corev1.ServicePort{
				{
					Name: "citus",
					Port: 5432,
				},
			},
		},
	}

	hostPathType = corev1.HostPathDirectoryOrCreate

	ClusterBackup = v1alpha1.PGClusterBackup{
		ObjectMeta: metav1.ObjectMeta{
			Name: PGClusterBackup,
		},
		Spec: v1alpha1.PGClusterBackupSpec{
			ClusterName: PGClusterName,
			VolumeSpec: &corev1.Volume{
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{
						Type: &hostPathType,
					},
				},
			},
		},
	}

	ClusterRestore = v1alpha1.PGClusterRestore{
		ObjectMeta: metav1.ObjectMeta{
			Name: PGClusterRestore,
		},
		Spec: v1alpha1.PGClusterRestoreSpec{
			ClusterName: PGClusterName,
			BackupName:  PGClusterBackup,
		},
	}

	jobTTL = int32(time.Duration(5 * time.Minute).Seconds())

	// backup job template
	BackupJob = batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pgc-backup-job",
			Namespace: "",
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: &jobTTL,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:            "backup",
							Image:           CitusImage,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Env: []corev1.EnvVar{
								{
									Name:  "PG_HOST",
									Value: "citus-0.citus-headless",
								},
								{
									Name:  "PG_PORT",
									Value: "5432",
								},
							},
							Command: []string{
								"sh",
								"-c",
								"pg_dumpall -U ${PGUSER} -h ${PG_HOST} -p ${PG_PORT} -f ${BACKUP_FILENAME}",
							},
						}, // container 1
					}, // end containers
				},
			}, // end template
		},
	} // end backup job define

	// restore job template
	RestoreJob = batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pgc-restore-job",
			Namespace: "",
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: &jobTTL,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:            "restore",
							Image:           CitusImage,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Env: []corev1.EnvVar{
								{
									Name:  "PG_HOST",
									Value: "citus-0.citus-headless",
								},
								{
									Name:  "PG_PORT",
									Value: "5432",
								},
							},
							Command: []string{
								"sh",
								"-c",
								"psql -U ${PGUSER} -h ${PG_HOST} -p ${PG_PORT} -f ${BACKUP_FILENAME}",
							},
						}, // container 1
					}, // end containers
				},
			}, // end template
		},
	} // end restore job define
)

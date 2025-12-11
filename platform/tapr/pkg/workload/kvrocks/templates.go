package kvrocks

import (
	"bytetrade.io/web3os/tapr/pkg/apis/apr/v1alpha1"
	"bytetrade.io/web3os/tapr/pkg/workload/utils"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"
)

const (
	DefaultKVRocksName      = "kvrocks"
	DefaultKVRocksImage     = "beclab/kvrocks:0.1.1"
	KVRocksVolumeName       = "kvrdata"
	KVRocksBackupVolumeName = "kvrbackup"
	KVRocksBackupDir        = "/backup"
	KVRocksDataDir          = "/var/lib/kvrocks-data"
	KVRocksConfDir          = "/var/lib/kvrocks"
	KVRocksBackupName       = "kvrocks-backup"
	KVRocksRestoreName      = "kvrocks-restore"
)

var (
	KVRocksStatefulSet = appv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: DefaultKVRocksName,
			Labels: map[string]string{
				"managed-by": "kvrocks-operator",
			},
		},

		Spec: appv1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":                    "kvrocks",
					"app.kubernetes.io/name": "kvrocks",
				},
			},

			Replicas: pointer.Int32(1),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":                         "kvrocks",
						"app.kubernetes.io/name":      "kvrocks",
						"app.bytetrade.io/middleware": "true",
						"pod-template-version":        "v1.1",
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
					InitContainers: []corev1.Container{
						{
							Name:            "init-kvrocks-cfg",
							Image:           DefaultKVRocksImage,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Command: []string{
								"sh",
								"-c",
								"test -f " + KVRocksDataDir + "/kvrocks.conf || cp -f " + KVRocksConfDir + "/kvrocks.conf " + KVRocksDataDir + "/kvrocks.conf",
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      KVRocksVolumeName,
									MountPath: KVRocksDataDir,
								},
							},
							SecurityContext: &corev1.SecurityContext{
								RunAsUser: pointer.Int64(0),
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:            "kvrocks",
							Image:           DefaultKVRocksImage,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Command: []string{
								"kvrocks",
								"-c", KVRocksDataDir + "/kvrocks.conf",
								"--dir", KVRocksDataDir,
								"--backup-dir", KVRocksBackupDir,
								"--pidfile", "/var/run/kvrocks/kvrocks.pid",
								"--bind", "0.0.0.0"},
							Env: []corev1.EnvVar{
								{
									Name: "POD_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
							}, // env
							Ports: []corev1.ContainerPort{
								{
									Name:          "kvrocks",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: 6666,
								},
							},
							SecurityContext: &corev1.SecurityContext{
								RunAsUser: pointer.Int64(0),
							},
							LivenessProbe: &corev1.Probe{
								InitialDelaySeconds: 10,
								ProbeHandler: corev1.ProbeHandler{
									TCPSocket: &corev1.TCPSocketAction{Port: intstr.FromInt(6666)},
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      KVRocksVolumeName,
									MountPath: KVRocksDataDir,
								},
								{
									Name:      KVRocksBackupVolumeName,
									MountPath: KVRocksBackupDir,
								},
							},
						}, // container kvrocks

					}, // containers

					Volumes: []corev1.Volume{
						{
							Name: KVRocksVolumeName,
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Type: utils.AnyPtr(corev1.HostPathDirectoryOrCreate),
								},
							},
						},
						{
							Name: KVRocksBackupVolumeName,
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Type: utils.AnyPtr(corev1.HostPathDirectoryOrCreate),
								},
							},
						},
					}, // volumes
				}, // pod spec
			}, // template
		}, // sts spec
	}

	KVRocksService = corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "",
		},

		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app":                    "kvrocks",
				"app.kubernetes.io/name": "kvrocks",
			},

			Type: corev1.ServiceTypeClusterIP,

			Ports: []corev1.ServicePort{
				{
					Name:       "kvrocks",
					Port:       6379,
					Protocol:   corev1.ProtocolTCP,
					TargetPort: intstr.FromInt(6666),
				},
			},
		},
	}

	KVRocksBackup = v1alpha1.KVRocksBackup{
		ObjectMeta: metav1.ObjectMeta{
			Name: KVRocksBackupName,
		},
		Spec: v1alpha1.KVRocksBackupSpec{},
	}

	KVRocksRestore = v1alpha1.KVRocksRestore{
		ObjectMeta: metav1.ObjectMeta{
			Name: KVRocksRestoreName,
		},
		Spec: v1alpha1.KVRocksRestoreSpec{},
	}
)

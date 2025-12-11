package rediscluster

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	store "kmodules.xyz/objectstore-api/api/v1"
)

type StorageType string

const (
	PersistentClaim StorageType = "persistent-claim"
	Ephemeral       StorageType = "ephemeral"
	HostPath        StorageType = "hostPath"
)

// RedisRole RedisCluster Node Role type
type RedisRole string

const (
	// RedisClusterNodeRoleMaster RedisCluster Master node role
	RedisClusterNodeRoleMaster RedisRole = "Master"
	// RedisClusterNodeRoleSlave RedisCluster Master node role
	RedisClusterNodeRoleSlave RedisRole = "Slave"
	// RedisClusterNodeRoleNone None node role
	RedisClusterNodeRoleNone RedisRole = "None"
)

// ClusterStatus Redis Cluster status
type ClusterStatus string

const (
	// ClusterStatusOK ClusterStatus OK
	ClusterStatusOK ClusterStatus = "Healthy"
	// ClusterStatusKO ClusterStatus KO
	ClusterStatusKO ClusterStatus = "Failed"
	// ClusterStatusCreating ClusterStatus Creating
	ClusterStatusCreating = "Creating"
	// ClusterStatusScaling ClusterStatus Scaling
	ClusterStatusScaling ClusterStatus = "Scaling"
	// ClusterStatusCalculatingRebalancing ClusterStatus Rebalancing
	ClusterStatusCalculatingRebalancing ClusterStatus = "Calculating Rebalancing"
	// ClusterStatusRebalancing ClusterStatus Rebalancing
	ClusterStatusRebalancing ClusterStatus = "Rebalancing"
	// ClusterStatusRollingUpdate ClusterStatus RollingUpdate
	ClusterStatusRollingUpdate ClusterStatus = "RollingUpdate"
	// ClusterStatusResetPassword ClusterStatus ResetPassword
	ClusterStatusResetPassword ClusterStatus = "ResetPassword"
)

// NodesPlacementInfo Redis Nodes placement mode information
type NodesPlacementInfo string

const (
	// NodesPlacementInfoBestEffort the cluster nodes placement is in best effort,
	// it means you can have 2 masters (or more) on the same VM.
	NodesPlacementInfoBestEffort NodesPlacementInfo = "BestEffort"
	// NodesPlacementInfoOptimal the cluster nodes placement is optimal,
	// it means on master by VM
	NodesPlacementInfoOptimal NodesPlacementInfo = "Optimal"
)

type RestorePhase string

const (
	// RestorePhaseRunning used for Restore that are currently running.
	RestorePhaseRunning RestorePhase = "Running"
	// RestorePhaseRestart used for Restore that are restart master nodes.
	RestorePhaseRestart RestorePhase = "Restart"
	// RestorePhaseSucceeded used for Restore that are Succeeded.
	RestorePhaseSucceeded RestorePhase = "Succeeded"
)

// DistributedRedisClusterSpec defines the desired state of DistributedRedisCluster
type DistributedRedisClusterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	Image            string                        `json:"image,omitempty"`
	ImagePullPolicy  corev1.PullPolicy             `json:"imagePullPolicy,omitempty"`
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	Command          []string                      `json:"command,omitempty"`
	Env              []corev1.EnvVar               `json:"env,omitempty"`
	MasterSize       int32                         `json:"masterSize,omitempty"`
	ClusterReplicas  int32                         `json:"clusterReplicas,omitempty"`
	ServiceName      string                        `json:"serviceName,omitempty"`
	Config           map[string]string             `json:"config,omitempty"`
	// Set RequiredAntiAffinity to force the master-slave node anti-affinity.
	RequiredAntiAffinity     bool                         `json:"requiredAntiAffinity,omitempty"`
	Affinity                 *corev1.Affinity             `json:"affinity,omitempty"`
	NodeSelector             map[string]string            `json:"nodeSelector,omitempty"`
	ToleRations              []corev1.Toleration          `json:"toleRations,omitempty"`
	SecurityContext          *corev1.PodSecurityContext   `json:"securityContext,omitempty"`
	ContainerSecurityContext *corev1.SecurityContext      `json:"containerSecurityContext,omitempty"`
	Annotations              map[string]string            `json:"annotations,omitempty"`
	Storage                  *RedisStorage                `json:"storage,omitempty"`
	Resources                *corev1.ResourceRequirements `json:"resources,omitempty"`
	PasswordSecret           *corev1.LocalObjectReference `json:"passwordSecret,omitempty"`
	Monitor                  *AgentSpec                   `json:"monitor,omitempty"`
	Init                     *InitSpec                    `json:"init,omitempty"`
}

type AgentSpec struct {
	Image      string          `json:"image,omitempty"`
	Prometheus *PrometheusSpec `json:"prometheus,omitempty"`
	// Arguments to the entrypoint.
	// The docker image's CMD is used if this is not provided.
	// Variable references $(VAR_NAME) are expanded using the container's environment. If a variable
	// cannot be resolved, the reference in the input string will be unchanged. The $(VAR_NAME) syntax
	// can be escaped with a double $$, ie: $$(VAR_NAME). Escaped references will never be expanded,
	// regardless of whether the variable exists or not.
	// Cannot be updated.
	// More info: https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/#running-a-command-in-a-shell
	// +optional
	Args []string `json:"args,omitempty"`
	// List of environment variables to set in the container.
	// Cannot be updated.
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	Env []corev1.EnvVar `json:"env,omitempty" patchStrategy:"merge" patchMergeKey:"name"`
	// Compute Resources required by exporter container.
	// Cannot be updated.
	// More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	// Security options the pod should run with.
	// More info: https://kubernetes.io/docs/concepts/policy/security-context/
	// More info: https://kubernetes.io/docs/tasks/configure-pod-container/security-context/
	// +optional
	SecurityContext *corev1.SecurityContext `json:"securityContext,omitempty"`
}

type PrometheusSpec struct {
	// Port number for the exporter side car.
	Port int32 `json:"port,omitempty"`

	// Namespace of Prometheus. Service monitors will be created in this namespace.
	Namespace string `json:"namespace,omitempty"`
	// Labels are key value pairs that is used to select Prometheus instance via ServiceMonitor labels.
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// Interval at which metrics should be scraped
	Interval string `json:"interval,omitempty"`
	//Annotations map[string]string `json:"annotations,omitempty"`
}

type InitSpec struct {
	BackupSource *BackupSourceSpec `json:"backupSource,omitempty"`
}

type BackupSourceSpec struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	// Arguments to the restore job
	Args []string `json:"args,omitempty"`
}

// RedisStorage defines the structure used to store the Redis Data
type RedisStorage struct {
	Size        resource.Quantity            `json:"size,omitempty"`
	Type        StorageType                  `json:"type"`
	Class       string                       `json:"class,omitempty"`
	DeleteClaim bool                         `json:"deleteClaim,omitempty"`
	HostPath    *corev1.HostPathVolumeSource `json:"hostPath,omitempty"`
}

// DistributedRedisClusterStatus defines the observed state of DistributedRedisCluster
// +k8s:openapi-gen=true
type DistributedRedisClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	Status               ClusterStatus      `json:"status"`
	Reason               string             `json:"reason,omitempty"`
	NumberOfMaster       int32              `json:"numberOfMaster,omitempty"`
	MinReplicationFactor int32              `json:"minReplicationFactor,omitempty"`
	MaxReplicationFactor int32              `json:"maxReplicationFactor,omitempty"`
	NodesPlacement       NodesPlacementInfo `json:"nodesPlacementInfo,omitempty"`
	Nodes                []RedisClusterNode `json:"nodes"`
	// +optional
	Restore Restore `json:"restore"`
}

type Restore struct {
	Phase  RestorePhase        `json:"phase,omitempty"`
	Backup *RedisClusterBackup `json:"backup,omitempty"`
}

// RedisClusterNode represent a RedisCluster Node
type RedisClusterNode struct {
	ID          string    `json:"id"`
	Role        RedisRole `json:"role"`
	IP          string    `json:"ip"`
	Port        string    `json:"port"`
	Slots       []string  `json:"slots,omitempty"`
	MasterRef   string    `json:"masterRef,omitempty"`
	PodName     string    `json:"podName"`
	NodeName    string    `json:"nodeName"`
	StatefulSet string    `json:"statefulSet"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:printcolumn:name="replicas",type=number,JSONPath=`.spec.clusterReplicas`
// +kubebuilder:printcolumn:name="service",type=string,JSONPath=`.spec.serviceName`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced, shortName={drc}, categories={all}
// DistributedRedisCluster is the Schema for the distributedredisclusters API
type DistributedRedisCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DistributedRedisClusterSpec   `json:"spec,omitempty"`
	Status DistributedRedisClusterStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DistributedRedisClusterList contains a list of DistributedRedisCluster
type DistributedRedisClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DistributedRedisCluster `json:"items"`
}

const (
	ResourceSingularBackup = "backup"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RedisClusterBackupSpec defines the desired state of RedisClusterBackup
// +k8s:openapi-gen=true
type RedisClusterBackupSpec struct {
	Image                 string        `json:"image,omitempty"`
	RedisClusterName      string        `json:"redisClusterName"`
	Storage               *RedisStorage `json:"storage,omitempty"`
	store.Backend         `json:",inline"`
	PodSpec               *PodSpec `json:"podSpec,omitempty"`
	ActiveDeadlineSeconds *int64   `json:"activeDeadlineSeconds,omitempty"`
}

type PodSpec struct {
	// ServiceAccountName is the name of the ServiceAccount to use to run this pod.
	// More info: https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`

	// Arguments to the entrypoint.
	// The docker image's CMD is used if this is not provided.
	// Variable references $(VAR_NAME) are expanded using the container's environment. If a variable
	// cannot be resolved, the reference in the input string will be unchanged. The $(VAR_NAME) syntax
	// can be escaped with a double $$, ie: $$(VAR_NAME). Escaped references will never be expanded,
	// regardless of whether the variable exists or not.
	// Cannot be updated.
	// More info: https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/#running-a-command-in-a-shell
	// +optional
	Args []string `json:"args,omitempty"`

	// NodeSelector is a selector which must be true for the pod to fit on a node.
	// Selector which must match a node's labels for the pod to be scheduled on that node.
	// More info: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Compute Resources required by the sidecar container.
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// If specified, the pod's scheduling constraints
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// If specified, the pod will be dispatched by specified scheduler.
	// If not specified, the pod will be dispatched by default scheduler.
	// +optional
	SchedulerName string `json:"schedulerName,omitempty"`

	// If specified, the pod's tolerations.
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	// ImagePullSecrets is an optional list of references to secrets in the same namespace to use for pulling any of the images used by this PodSpec.
	// If specified, these secrets will be passed to individual puller implementations for them to use.
	// +optional
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`

	// List of environment variables to set in the container.
	// Cannot be updated.
	// +optional
	Env []corev1.EnvVar `json:"env,omitempty"`

	// List of initialization containers belonging to the pod.
	// Init containers are executed in order prior to containers being started. If any
	// init container fails, the pod is considered to have failed and is handled according
	// to its restartPolicy. The name for an init container or normal container must be
	// unique among all containers.
	// Init containers may not have Lifecycle actions, Readiness probes, or Liveness probes.
	// The resourceRequirements of an init container are taken into account during scheduling
	// by finding the highest request/limit for each resource type, and then using the max of
	// of that value or the sum of the normal containers. Limits are applied to init containers
	// in a similar fashion.
	// Init containers cannot currently be added or removed.
	// Cannot be updated.
	// More info: https://kubernetes.io/docs/concepts/workloads/pods/init-containers/
	// +patchMergeKey=name
	// +patchStrategy=merge
	InitContainers []corev1.Container `json:"initContainers,omitempty" patchStrategy:"merge" patchMergeKey:"name"`

	// If specified, indicates the pod's priority. "system-node-critical" and
	// "system-cluster-critical" are two special keywords which indicate the
	// highest priorities with the former being the highest priority. Any other
	// name must be defined by creating a PriorityClass object with that name.
	// If not specified, the pod priority will be default or zero if there is no
	// default.
	// +optional
	PriorityClassName string `json:"priorityClassName,omitempty"`
	// The priority value. Various system components use this field to find the
	// priority of the pod. When Priority Admission Controller is enabled, it
	// prevents users from setting this field. The admission controller populates
	// this field from PriorityClassName.
	// The higher the value, the higher the priority.
	// +optional
	Priority *int32 `json:"priority,omitempty"`

	// SecurityContext holds pod-level security attributes and common container settings.
	// Optional: Defaults to empty.  See type description for default values of each field.
	// +optional
	SecurityContext *corev1.PodSecurityContext `json:"securityContext,omitempty"`

	// Periodic probe of container liveness.
	// Container will be restarted if the probe fails.
	// Controllers may set default LivenessProbe if no liveness probe is provided.
	// To ignore defaulting, set the value to empty LivenessProbe "{}".
	// Cannot be updated.
	// More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes
	// +optional
	LivenessProbe *corev1.Probe `json:"livenessProbe,omitempty"`

	// Periodic probe of container service readiness.
	// Container will be removed from service endpoints if the probe fails.
	// Cannot be updated.
	// Controllers may set default ReadinessProbe if no readyness probe is provided.
	// To ignore defaulting, set the value to empty ReadynessProbe "{}".
	// More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes
	// +optional
	ReadinessProbe *corev1.Probe `json:"readinessProbe,omitempty"`

	// Actions that the management system should take in response to container lifecycle events.
	// Cannot be updated.
	// +optional
	Lifecycle *corev1.Lifecycle `json:"lifecycle,omitempty"`
}

type BackupPhase string

const (
	// used for Backup that are currently running
	BackupPhaseRunning BackupPhase = "Running"
	// used for Backup that are Succeeded
	BackupPhaseSucceeded BackupPhase = "Succeeded"
	// used for Backup that are Failed
	BackupPhaseFailed BackupPhase = "Failed"
	// used for Backup that are Ignored
	BackupPhaseIgnored BackupPhase = "Ignored"
)

// RedisClusterBackupStatus defines the observed state of RedisClusterBackup
// +k8s:openapi-gen=true
type RedisClusterBackupStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	StartTime       *metav1.Time `json:"startTime,omitempty"`
	CompletionTime  *metav1.Time `json:"completionTime,omitempty"`
	Phase           BackupPhase  `json:"phase,omitempty"`
	Reason          string       `json:"reason,omitempty"`
	MasterSize      int32        `json:"masterSize,omitempty"`
	ClusterReplicas int32        `json:"clusterReplicas,omitempty"`
	ClusterImage    string       `json:"clusterImage,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RedisClusterBackup is the Schema for the redisclusterbackups API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=redisclusterbackups,scope=Namespaced
type RedisClusterBackup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RedisClusterBackupSpec   `json:"spec,omitempty"`
	Status RedisClusterBackupStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RedisClusterBackupList contains a list of RedisClusterBackup
type RedisClusterBackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RedisClusterBackup `json:"items"`
}

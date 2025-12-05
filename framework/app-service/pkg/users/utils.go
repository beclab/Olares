package users

import (
	"context"
	"errors"
	"fmt"
	iamv1alpha2 "github.com/beclab/api/iam/v1alpha2"
	"github.com/thoas/go-funk"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

const (
	minPodNumPerUser              = 25
	reservedPodNumForUser         = 10
	userAnnotationLimitsCpuKey    = "bytetrade.io/user-cpu-limit"
	userAnnotationLimitsMemoryKey = "bytetrade.io/user-memory-limit"
	userAnnotationRoleKey         = "bytetrade.io/owner-role"
)

var SystemReservedKeyWords = []string{
	"user",
	"system",
	"space",
	"default",
	"os",
	"kubesphere",
	"kube",
	"kubekey",
	"kubernetes",
	"gpu",
	"tapr",
	"bfl",
	"bytetrade",
	"project",
	"pod",
}

func CheckUsername(user *iamv1alpha2.User) error {
	if funk.Contains(SystemReservedKeyWords, user.Name) {
		return fmt.Errorf("user create: %q is a system reserved keyword and cannot be set as a username", user.Name)
	}
	return nil
}

func CheckUserRole(user *iamv1alpha2.User) error {
	//TODO:hys
	return nil
}

func CheckClusterPodCapacity(ctx context.Context, ctrlClient client.Client) (bool, error) {
	var nodes corev1.NodeList
	err := ctrlClient.List(ctx, &nodes)
	if err != nil {
		return false, err
	}

	var currentPodNum, maxPodNum int64
	nodeMap := sets.String{}
	for _, node := range nodes.Items {
		if !IsNodeReady(&node) || node.Spec.Unschedulable {
			continue
		}
		pods, _ := node.Status.Capacity.Pods().AsInt64()
		maxPodNum += pods
		nodeMap.Insert(node.Name)
	}
	var pods corev1.PodList
	err = ctrlClient.List(ctx, &pods)
	if err != nil {
		return false, err
	}

	for _, pod := range pods.Items {
		if IsPodActive(&pod) && (nodeMap.Has(pod.Spec.NodeName) || pod.Status.Phase == corev1.PodPending) {
			currentPodNum++
		}
	}

	klog.Infof("currentPodNum: %v, maxPodNum: %v", currentPodNum, maxPodNum)
	if currentPodNum+minPodNumPerUser > maxPodNum-reservedPodNumForUser {
		return false, nil
	}
	return true, nil
}

func IsPodActive(p *corev1.Pod) bool {
	return corev1.PodSucceeded != p.Status.Phase &&
		corev1.PodFailed != p.Status.Phase &&
		p.DeletionTimestamp == nil
}

func IsNodeReady(node *corev1.Node) bool {
	for _, c := range node.Status.Conditions {
		if c.Type == corev1.NodeReady {
			return c.Status == corev1.ConditionTrue
		}
	}
	return false
}

func ValidateResourceLimits(user *iamv1alpha2.User) error {
	if user.Annotations == nil {
		return errors.New("nil user annotations")
	}
	memoryLimit := user.Annotations[userAnnotationLimitsMemoryKey]
	cpuLimit := user.Annotations[userAnnotationLimitsCpuKey]

	memory, err := resource.ParseQuantity(memoryLimit)
	if err != nil {
		return fmt.Errorf("invalid format of memory limit %w", err)
	}

	cpu, err := resource.ParseQuantity(cpuLimit)
	if err != nil {
		return fmt.Errorf("invalid format of cpu limit %w", err)
	}

	// Check against default limits
	defaultMemoryLimit, _ := resource.ParseQuantity(os.Getenv("USER_DEFAULT_MEMORY_LIMIT"))
	defaultCpuLimit, _ := resource.ParseQuantity(os.Getenv("USER_DEFAULT_CPU_LIMIT"))

	if defaultMemoryLimit.Cmp(memory) > 0 {
		return fmt.Errorf("memory limit cannot be less than %s", defaultMemoryLimit.String())
	}

	if defaultCpuLimit.Cmp(cpu) > 0 {
		return fmt.Errorf("cpu limit cannot be less than %s", defaultCpuLimit.String())
	}

	return nil
}

type OlaresName string

func (s OlaresName) UserName() string {
	return s.UserAndDomain()[0]
}

func NewOlaresName(username, domainName string) OlaresName {
	return OlaresName(fmt.Sprintf("%s@%s", username, domainName))
}

func (s OlaresName) UserAndDomain() []string {
	return strings.Split(string(s), "@")
}

func (s OlaresName) UserZone() string {
	return strings.Join(s.UserAndDomain(), ".")
}

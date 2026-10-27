package users

import (
	"errors"
	"fmt"
	iamv1alpha2 "github.com/beclab/api/iam/v1alpha2"
	"github.com/thoas/go-funk"
	"k8s.io/apimachinery/pkg/api/resource"
	"os"
)

const (
	userAnnotationLimitsCpuKey    = "bytetrade.io/user-cpu-limit"
	userAnnotationLimitsMemoryKey = "bytetrade.io/user-memory-limit"
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

package watchers

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ValidWatchDuration(meta *metav1.ObjectMeta) bool {
	return (time.Since(meta.CreationTimestamp.Time) < time.Minute ||
		(meta.DeletionTimestamp != nil && time.Since(meta.DeletionTimestamp.Time) < time.Minute))
}

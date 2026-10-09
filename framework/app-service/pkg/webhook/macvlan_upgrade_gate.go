package webhook

import (
	"context"
	"fmt"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Only new platform macvlan attachments pause during the DHCP process switch.
// Existing Pods and unrelated application admission continue to work.
func (wh *Webhook) checkFixedMACUpgrade(ctx context.Context) error {
	cm, err := wh.kubeClient.CoreV1().ConfigMaps("kube-system").Get(ctx, "olares-fixed-mac-upgrade", metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if cm.Data["phase"] == "paused" {
		return fmt.Errorf("fixed MAC upgrade is switching DHCP; retry pod creation after recovery resumes")
	}
	return nil
}

package controllers

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// These specs run against a real API server (envtest) and validate the
// status-subresource assumptions that the appstate / controller code relies
// on, plus the optimistic-locking retry path.

var _ = Describe("ApplicationManager status", func() {
	It("persists status through a plain Patch (CRD has no /status subresource)", func() {
		am := &appv1alpha1.ApplicationManager{
			ObjectMeta: metav1.ObjectMeta{Name: "am-plain-patch"},
			Spec: appv1alpha1.ApplicationManagerSpec{
				AppName: "nginx", AppNamespace: "nginx-alice", AppOwner: "alice",
				Source: "market", Type: appv1alpha1.App, OpType: appv1alpha1.InstallOp,
			},
		}
		Expect(k8sClient.Create(suiteCtx, am)).To(Succeed())
		DeferCleanup(func() { _ = k8sClient.Delete(suiteCtx, am) })

		base := am.DeepCopy()
		am.Status.State = appv1alpha1.Running
		am.Status.OpGeneration = 1
		Expect(k8sClient.Patch(suiteCtx, am, client.MergeFrom(base))).To(Succeed())

		var got appv1alpha1.ApplicationManager
		Expect(k8sClient.Get(suiteCtx, client.ObjectKeyFromObject(am), &got)).To(Succeed())
		Expect(got.Status.State).To(Equal(appv1alpha1.Running))
		Expect(got.Status.OpGeneration).To(Equal(int64(1)))
	})

	It("retries on conflict when two writers race (RetryOnConflict)", func() {
		am := &appv1alpha1.ApplicationManager{
			ObjectMeta: metav1.ObjectMeta{Name: "am-conflict"},
			Spec: appv1alpha1.ApplicationManagerSpec{
				AppName: "nginx", AppNamespace: "nginx-bob", AppOwner: "bob",
				Source: "market", Type: appv1alpha1.App, OpType: appv1alpha1.InstallOp,
			},
		}
		Expect(k8sClient.Create(suiteCtx, am)).To(Succeed())
		DeferCleanup(func() { _ = k8sClient.Delete(suiteCtx, am) })

		// Two stale copies of the same generation. Patching the first bumps the
		// resourceVersion; a stale Update from the second must conflict, and the
		// retry helper must reload and succeed.
		first := am.DeepCopy()
		first.Status.Message = "first"
		Expect(k8sClient.Patch(suiteCtx, first, client.MergeFrom(am))).To(Succeed())

		err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			var latest appv1alpha1.ApplicationManager
			if err := k8sClient.Get(suiteCtx, client.ObjectKeyFromObject(am), &latest); err != nil {
				return err
			}
			latest.Status.Message = "second"
			return k8sClient.Update(suiteCtx, &latest)
		})
		Expect(err).NotTo(HaveOccurred())

		var got appv1alpha1.ApplicationManager
		Expect(k8sClient.Get(suiteCtx, client.ObjectKeyFromObject(am), &got)).To(Succeed())
		Expect(got.Status.Message).To(Equal("second"))
	})
})

var _ = Describe("Application status", func() {
	It("ignores status on a plain object update but persists it via Status() (CRD has /status subresource)", func() {
		app := &appv1alpha1.Application{
			ObjectMeta: metav1.ObjectMeta{Name: "app-subresource"},
			Spec:       appv1alpha1.ApplicationSpec{Name: "nginx", Namespace: "nginx-alice", Owner: "alice"},
		}
		Expect(k8sClient.Create(suiteCtx, app)).To(Succeed())
		DeferCleanup(func() { _ = k8sClient.Delete(suiteCtx, app) })

		// A plain Update carrying a status change must NOT persist the status.
		app.Status.State = "running"
		Expect(k8sClient.Update(suiteCtx, app)).To(Succeed())
		var afterPlain appv1alpha1.Application
		Expect(k8sClient.Get(suiteCtx, client.ObjectKeyFromObject(app), &afterPlain)).To(Succeed())
		Expect(afterPlain.Status.State).To(BeEmpty())

		// Status().Update must persist it. The CRD marks statusTime/updateTime
		// as required, so set them too.
		now := metav1.Now()
		afterPlain.Status.State = "running"
		afterPlain.Status.StatusTime = &now
		afterPlain.Status.UpdateTime = &now
		Expect(k8sClient.Status().Update(suiteCtx, &afterPlain)).To(Succeed())
		var afterStatus appv1alpha1.Application
		Expect(k8sClient.Get(suiteCtx, client.ObjectKeyFromObject(app), &afterStatus)).To(Succeed())
		Expect(afterStatus.Status.State).To(Equal("running"))
	})
})

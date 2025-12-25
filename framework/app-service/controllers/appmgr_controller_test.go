package controllers

//
//import (
//	"github.com/beclab/Olares/framework/app-service/pkg/appinstaller"
//	"github.com/beclab/Olares/framework/app-service/pkg/kubesphere"
//	"context"
//	"encoding/json"
//	"github.com/agiledragon/gomonkey/v2"
//	. "github.com/onsi/ginkgo/v2"
//	. "github.com/onsi/gomega"
//	"k8s.io/apimachinery/pkg/types"
//	"k8s.io/client-go/kubernetes/scheme"
//	"k8s.io/client-go/rest"
//	ctrl "sigs.k8s.io/controller-runtime"
//	"sigs.k8s.io/controller-runtime/pkg/client"
//	"sigs.k8s.io/controller-runtime/pkg/client/fake"
//	"sigs.k8s.io/controller-runtime/pkg/reconcile"
//
//	appv1alpha1 "github.com/beclab/Olares/framework/app-service/api/app.bytetrade.io/v1alpha1"
//	appsv1 "k8s.io/api/apps/v1"
//	corev1 "k8s.io/api/core/v1"
//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
//	"k8s.io/apimachinery/pkg/runtime"
//)
//
//type mockImageManager struct{}
//
//func (m *mockImageManager) UpdateStatus(ctx context.Context, name, state, message string) error {
//	return nil
//}
//
//func (m *mockImageManager) Create(ctx context.Context, am *appv1alpha1.ApplicationManager, refs []appv1alpha1.Ref) error {
//	return nil
//}
//
//func (m *mockImageManager) PollDownloadProgress(ctx context.Context, am *appv1alpha1.ApplicationManager) error {
//	return nil
//}
//
//var _ = Describe("ApplicationManagerController", func() {
//	var (
//		ctx           context.Context
//		cancel        context.CancelFunc
//		k8sClient     client.Client
//		controller    *ApplicationManagerController
//		testScheme    *runtime.Scheme
//		mockImgClient *mockImageManager
//	)
//	var patchForgetUsername *gomonkey.Patches
//	var patchForHandleDownloading *gomonkey.Patches
//	var patchForhandleDownloading *gomonkey.Patches
//
//	var patchForHandleInstalling *gomonkey.Patches
//	var patchForHandleInitializing *gomonkey.Patches
//
//	BeforeEach(func() {
//		ctx, cancel = context.WithCancel(context.Background())
//
//		testScheme = runtime.NewScheme()
//		Expect(scheme.AddToScheme(testScheme)).To(Succeed())
//		Expect(appv1alpha1.AddToScheme(testScheme)).To(Succeed())
//		Expect(corev1.AddToScheme(testScheme)).To(Succeed())
//		Expect(appsv1.AddToScheme(testScheme)).To(Succeed())
//
//		// 创建一个模拟的 K8s 客户端
//		k8sClient = fake.NewClientBuilder().WithScheme(testScheme).Build()
//
//		patchForgetUsername = gomonkey.ApplyFunc(kubesphere.GetAdminUsername, func(_ context.Context, _ *rest.Config) (string, error) {
//			return "admin", nil
//		})
//
//		// 创建模拟的 ImageManager
//		mockImgClient = &mockImageManager{}
//
//		// 创建控制器
//		controller = &ApplicationManagerController{
//			Client:      k8sClient,
//			KubeConfig:  &rest.Config{},
//			ImageClient: mockImgClient,
//		}
//		patchForHandleDownloading = gomonkey.ApplyFunc(controller.HandleDownloading, func(_ context.Context, _ *appv1alpha1.ApplicationManager) error {
//			return nil
//		})
//		patchForhandleDownloading = gomonkey.ApplyFunc(controller.handleDownloading, func(_ context.Context, _ *appv1alpha1.ApplicationManager) error {
//			return nil
//		})
//
//		patchForHandleInstalling = gomonkey.ApplyFunc(controller.HandleInstalling, func(_ context.Context, _ *appv1alpha1.ApplicationManager) error {
//			return nil
//		})
//		patchForHandleInitializing = gomonkey.ApplyFunc(controller.HandleInitializing, func(_ context.Context, _ *appv1alpha1.ApplicationManager) error {
//			return nil
//		})
//	})
//
//	AfterEach(func() {
//		cancel()
//		patchForgetUsername.Reset()
//		patchForHandleDownloading.Reset()
//		patchForhandleDownloading.Reset()
//		patchForHandleInstalling.Reset()
//		patchForHandleInitializing.Reset()
//	})
//
//	Context("Reconcile", func() {
//		It("should handle non-existent ApplicationManager", func() {
//			req := reconcile.Request{
//				NamespacedName: types.NamespacedName{
//					Name: "non-existent",
//				},
//			}
//
//			result, err := controller.Reconcile(ctx, req)
//			Expect(err).NotTo(HaveOccurred())
//			Expect(result).To(Equal(ctrl.Result{}))
//		})
//
//		It("should handle ApplicationManager in Pending state", func() {
//			am := &appv1alpha1.ApplicationManager{
//				ObjectMeta: metav1.ObjectMeta{
//					Name: "test-app",
//				},
//				Spec: appv1alpha1.ApplicationManagerSpec{
//					AppName:      "test-app",
//					AppNamespace: "default",
//					AppOwner:     "test-owner",
//					Type:         appv1alpha1.App,
//					Config:       `{"chartsName": "../testdata/windows"}`,
//				},
//				Status: appv1alpha1.ApplicationManagerStatus{
//					State: appv1alpha1.Pending,
//				},
//			}
//			Expect(k8sClient.Create(ctx, am)).To(Succeed())
//
//			node := &corev1.Node{
//				ObjectMeta: metav1.ObjectMeta{
//					Name: "test-node",
//				},
//				Status: corev1.NodeStatus{
//					Conditions: []corev1.NodeCondition{
//						{
//							Type:   corev1.NodeReady,
//							Status: corev1.ConditionTrue,
//						},
//					},
//				},
//			}
//			Expect(k8sClient.Create(ctx, node)).To(Succeed())
//
//			req := reconcile.Request{
//				NamespacedName: types.NamespacedName{
//					Name: "test-app",
//				},
//			}
//
//			result, err := controller.Reconcile(ctx, req)
//			Expect(err).NotTo(HaveOccurred())
//			Expect(result).To(Equal(ctrl.Result{}))
//
//			updatedAm := &appv1alpha1.ApplicationManager{}
//			Expect(k8sClient.Get(ctx, types.NamespacedName{Name: "test-app"}, updatedAm)).To(Succeed())
//			Expect(updatedAm.Status.State).To(Equal(appv1alpha1.Downloading))
//		})
//	})
//
//	Context("handleDownloading", func() {
//		It("should handle downloading state", func() {
//			appConfig := &appinstaller.ApplicationConfig{
//				AppName:    "test-app",
//				Namespace:  "default",
//				OwnerName:  "test-owner",
//				ChartsName: "test-chart",
//			}
//			configBytes, err := json.Marshal(appConfig)
//			Expect(err).NotTo(HaveOccurred())
//
//			am := &appv1alpha1.ApplicationManager{
//				ObjectMeta: metav1.ObjectMeta{
//					Name: "test-app",
//				},
//				Spec: appv1alpha1.ApplicationManagerSpec{
//					AppName:      "test-app",
//					AppNamespace: "default",
//					AppOwner:     "test-owner",
//					Type:         appv1alpha1.App,
//					Config:       string(configBytes),
//				},
//				Status: appv1alpha1.ApplicationManagerStatus{
//					State: appv1alpha1.Downloading,
//				},
//			}
//
//			node := &corev1.Node{
//				ObjectMeta: metav1.ObjectMeta{
//					Name: "test-node",
//				},
//				Status: corev1.NodeStatus{
//					Conditions: []corev1.NodeCondition{
//						{
//							Type:   corev1.NodeReady,
//							Status: corev1.ConditionTrue,
//						},
//					},
//				},
//			}
//			Expect(k8sClient.Create(ctx, node)).To(Succeed())
//
//			im := &appv1alpha1.ImageManager{
//				ObjectMeta: metav1.ObjectMeta{
//					Name: "test-app",
//				},
//				Spec: appv1alpha1.ImageManagerSpec{
//					AppName:      "test-app",
//					AppNamespace: "default",
//					AppOwner:     "test-owner",
//					Nodes:        []string{"test-node"},
//				},
//				Status: appv1alpha1.ImageManagerStatus{
//					State:      "completed",
//					Message:    "success",
//					Conditions: map[string]map[string]map[string]string{},
//				},
//			}
//			Expect(k8sClient.Create(ctx, im)).To(Succeed())
//
//			err = controller.HandleDownloading(ctx, am)
//			if err != nil {
//				Expect(err.Error()).To(ContainSubstring("failed to get image refs"))
//			}
//
//			Expect(am.Status.State).To(Equal(appv1alpha1.Installing))
//		})
//	})
//
//	Context("handleInstalling", func() {
//		It("should handle installing state", func() {
//			appConfig := &appinstaller.ApplicationConfig{
//				AppName:    "test-app",
//				Namespace:  "default",
//				OwnerName:  "test-owner",
//				ChartsName: "test-chart",
//			}
//			configBytes, err := json.Marshal(appConfig)
//			Expect(err).NotTo(HaveOccurred())
//
//			am := &appv1alpha1.ApplicationManager{
//				ObjectMeta: metav1.ObjectMeta{
//					Name: "test-app",
//				},
//				Spec: appv1alpha1.ApplicationManagerSpec{
//					AppName:      "test-app",
//					AppNamespace: "default",
//					AppOwner:     "test-owner",
//					Type:         appv1alpha1.App,
//					Config:       string(configBytes),
//				},
//				Status: appv1alpha1.ApplicationManagerStatus{
//					State: appv1alpha1.Installing,
//					Payload: map[string]string{
//						"token": "test-token",
//					},
//				},
//			}
//
//			err = controller.HandleInstalling(ctx, am)
//			if err != nil {
//				Expect(err.Error()).To(ContainSubstring("failed to create helm ops"))
//			}
//
//
//			Expect(am.Status.State).To(Equal(appv1alpha1.Initializing))
//		})
//	})
//
//	Context("handleInitializing", func() {
//		It("should handle initializing state", func() {
//			appConfig := &appinstaller.ApplicationConfig{
//				AppName:    "test-app",
//				Namespace:  "default",
//				OwnerName:  "test-owner",
//				ChartsName: "test-chart",
//			}
//			configBytes, err := json.Marshal(appConfig)
//			Expect(err).NotTo(HaveOccurred())
//
//			am := &appv1alpha1.ApplicationManager{
//				ObjectMeta: metav1.ObjectMeta{
//					Name: "test-app",
//				},
//				Spec: appv1alpha1.ApplicationManagerSpec{
//					AppName:      "test-app",
//					AppNamespace: "default",
//					AppOwner:     "test-owner",
//					Type:         appv1alpha1.App,
//					Config:       string(configBytes),
//				},
//				Status: appv1alpha1.ApplicationManagerStatus{
//					State: appv1alpha1.Initializing,
//					Payload: map[string]string{
//						"token": "test-token",
//					},
//				},
//			}
//
//			err = controller.HandleInitializing(ctx, am)
//			if err != nil {
//				Expect(err.Error()).To(ContainSubstring("failed to create helm ops"))
//			}
//
//			Expect(am.Status.State).To(Equal(appv1alpha1.Running))
//		})
//	})
//
//	Context("handleUninstalling", func() {
//		It("should handle uninstalling state", func() {
//			am := &appv1alpha1.ApplicationManager{
//				ObjectMeta: metav1.ObjectMeta{
//					Name: "test-app",
//				},
//				Spec: appv1alpha1.ApplicationManagerSpec{
//					AppName:      "test-app",
//					AppNamespace: "default",
//					AppOwner:     "test-owner",
//					Type:         appv1alpha1.App,
//				},
//				Status: appv1alpha1.ApplicationManagerStatus{
//					State: appv1alpha1.Uninstalling,
//					Payload: map[string]string{
//						"token": "test-token",
//					},
//				},
//			}
//
//			err := controller.HandleUninstalling(ctx, am)
//			if err != nil {
//				Expect(err.Error()).To(ContainSubstring("failed to create helm ops"))
//			}
//
//			Expect(am.Status.State).To(Equal(appv1alpha1.Uninstalled))
//		})
//	})
//})

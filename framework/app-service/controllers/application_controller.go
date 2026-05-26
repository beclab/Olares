package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/cache"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/cluster"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/gateway"
	"github.com/beclab/Olares/framework/app-service/pkg/gateway/routecontrol"
	"github.com/beclab/Olares/framework/app-service/pkg/helm"
	"github.com/beclab/Olares/framework/app-service/pkg/kubesphere"
	"github.com/beclab/Olares/framework/app-service/pkg/users/userspace"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	"github.com/beclab/api/pkg/generated/clientset/versioned"

	"github.com/thoas/go-funk"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/storage/driver"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	applicationFinalizer = "finalizers.bytetrade.io/application"
)

var protectedRelease = []string{"headscale"}

// ApplicationReconciler reconciles a Application object
type ApplicationReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	AppClientset *versioned.Clientset
	Kubeconfig   *rest.Config
}

//+kubebuilder:rbac:groups=app.bytetrade.io,resources=applications,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=app.bytetrade.io,resources=applications/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=app.bytetrade.io,resources=applications/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Application object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *ApplicationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	ctrl.Log.Info("reconcile request", "name", req.Name, "namespace", req.Namespace)

	if req.Namespace == "" {
		// Cluster-scoped Application CR updates (e.g. kubectl annotate route-mode)
		// enqueue with an empty namespace. Reconcile gateway routes directly.
		if req.Name != "" {
			app, err := r.AppClientset.AppV1alpha1().Applications().Get(ctx, req.Name, metav1.GetOptions{})
			if err != nil {
				if !apierrors.IsNotFound(err) {
					return ctrl.Result{}, err
				}
			} else if err := r.ensureAppGatewayRouteMode(ctx, app); err != nil {
				klog.Warningf("ensure gateway route-mode for Application %s err=%v", req.Name, err)
			} else if err := r.ensureCallerInClusterAnnotation(ctx, app); err != nil {
				klog.Warningf("ensure in-cluster annotation for Application %s err=%v", req.Name, err)
			} else if err := r.reconcileSharedRouteRegistry(ctx, app); err != nil {
				klog.Warningf("reconcile SharedRouteRegistry for Application %s err=%v", req.Name, err)
			} else {
				r.reconcileCallerNamespace(ctx, app)
			}
		}
		return ctrl.Result{}, nil
	}

	validAppObjects := make(map[string]client.Object)
	deletingObjects := make(map[string]client.Object)

	reqAppNames := strings.Split(req.Name, ",")
	for _, name := range reqAppNames {
		// init requested app object
		validAppObjects[name] = nil
	}

	// get deployments installed by app installer
	findAppObject := func(list client.ObjectList) error {
		if err := r.List(ctx, list, client.InNamespace(req.Namespace)); err == nil {
			listObjects, err := apimeta.ExtractList(list)
			if err != nil {
				ctrl.Log.Error(err, "extract list error", "name label", req.Name, "namespace", req.Namespace)
				return err
			}

			for _, o := range listObjects {
				d := o.(client.Object)
				if owner, ok := d.GetLabels()[constants.ApplicationOwnerLabel]; !ok || owner == "" {
					// ignore ownerless deployments
					continue
				}
				if middleware, ok := d.GetLabels()[constants.ApplicationMiddlewareLabel]; ok && middleware == "true" {
					continue
				}
				// for multi-app in one deployment/statefulset, we can not find only one object via
				// namespace and label filter, so have to filter in object list
				apps := getAppName(d)
				if len(apps) == 0 {
					continue
				}
				klog.Infof("apps: %v", apps)
				for _, name := range apps {
					// found a valid app object
					if d.GetDeletionTimestamp() == nil {
						validAppObjects[name] = d
						klog.Errorf("valid app name: %s", name)
					} else {
						deletingObjects[name] = d
						klog.Errorf("deleting app name: %s", name)
					} // end if deployment is deleted
				}

			} // end loop deployment.Items
		} else {
			ctrl.Log.Error(err, "list deployments or statefulset error", "name label", req.Name, "namespace", req.Namespace)
			return err
		} // end if get deployments list

		return nil
	}

	var deployemnts appsv1.DeploymentList
	err := findAppObject(&deployemnts)
	if err != nil {
		return ctrl.Result{}, err
	}

	// try to get statefulset
	var statefulsets appsv1.StatefulSetList
	err = findAppObject(&statefulsets)
	if err != nil {
		return ctrl.Result{}, err
	}

	for name := range deletingObjects {
		if _, ok := validAppObjects[name]; !ok {
			validAppObjects[name] = nil
		}
	}

	for name, validAppObject := range validAppObjects {
		app, err := r.AppClientset.AppV1alpha1().Applications().Get(ctx, fmtAppName(name, req.Namespace), metav1.GetOptions{})
		klog.Infof("get app err=%v, validateAPpis nil %v,app=%v", err, validAppObject == nil, fmtAppName(name, req.Namespace))
		if validAppObject != nil {
			// create or update application
			if err != nil {
				if apierrors.IsNotFound(err) {
					// check if a new deployment created or not
					ctrl.Log.Info("create app from deployment watching", "name", validAppObject.GetName(), "namespace", validAppObject.GetNamespace(), "appname", name)
					err = r.createApplication(ctx, req, validAppObject, name)
					if err != nil {
						ctrl.Log.Info("create app failed", "app", name, "err", err)
						return ctrl.Result{}, err
					}
					continue
				}
				return ctrl.Result{}, err
			} // end if error

			ctrl.Log.Info("Application update", "name", app.Name, "spec.name", app.Spec.Name, "spec.owner", app.Spec.Owner)
			err = r.updateApplication(ctx, req, validAppObject, app, name)
			if err != nil {
				return ctrl.Result{Requeue: true}, err
			}
			//}
		} else {
			// deployment or statefulset is nil, delete application
			if err == nil && app != nil {
				ctrl.Log.Info("Application delete", "name", app.Name, "spec.name", app.Spec.Name, "spec.owner", app.Spec.Owner)
				err = r.Delete(ctx, app.DeepCopy())
				if err != nil && !apierrors.IsNotFound(err) {
					return ctrl.Result{}, err
				}
				if funk.Contains(protectedRelease, app.Spec.Name) {
					return ctrl.Result{}, nil
				}
				err = r.clearHelmHistory(app.Spec.Name, app.Spec.Namespace)
				if err != nil && !errors.Is(err, driver.ErrReleaseNotFound) {
					return ctrl.Result{RequeueAfter: 2 * time.Second}, err
				}

			} else if apierrors.IsNotFound(err) {
				// app not found, just return
				return ctrl.Result{}, nil
			}
		}
	}

	// v2 shared pilots often run the workload in a *-shared namespace while the
	// Application.spec.namespace points at the installer's user namespace. Local
	// Deployments may also lack installer owner/name labels, so the loop above
	// never calls updateApplication. Always reconcile gateway routes from the
	// Application CR(s) tied to this workload namespace.
	if err := r.reconcileGatewayRoutesForWorkloadNS(ctx, req.Namespace); err != nil {
		klog.Warningf("reconcile gateway routes for workload namespace %s err=%v", req.Namespace, err)
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	c, err := ctrl.NewControllerManagedBy(mgr).
		For(&appv1alpha1.Application{}).
		Build(r)

	if err != nil {
		return err
	}

	// watch the application enqueue formarted request
	err = c.Watch(source.Kind(
		mgr.GetCache(),
		&appv1alpha1.Application{},
		handler.TypedEnqueueRequestsFromMapFunc(
			func(ctx context.Context, app *appv1alpha1.Application) []reconcile.Request {
				return []reconcile.Request{{NamespacedName: types.NamespacedName{
					Name:      app.Spec.Name,
					Namespace: app.Spec.Namespace}},
				}
			}),
	))

	if err != nil {
		return err
	}

	err = c.Watch(source.Kind(
		mgr.GetCache(),
		&appv1alpha1.Application{},
		handler.TypedEnqueueRequestsFromMapFunc(
			func(ctx context.Context, app *appv1alpha1.Application) []reconcile.Request {
				if app == nil || app.Spec.Namespace == "" {
					return nil
				}
				return []reconcile.Request{{NamespacedName: types.NamespacedName{
					Name:      app.Spec.Name,
					Namespace: app.Spec.Namespace,
				}}}
			}),
		predicate.TypedFuncs[*appv1alpha1.Application]{
			CreateFunc: func(e event.TypedCreateEvent[*appv1alpha1.Application]) bool {
				app := e.Object
				return app != nil && strings.EqualFold(app.Annotations[gateway.AnnotationInCluster], gateway.InClusterGateway)
			},
			UpdateFunc: func(e event.TypedUpdateEvent[*appv1alpha1.Application]) bool {
				return inClusterAnnotationChanged(e.ObjectOld, e.ObjectNew)
			},
			DeleteFunc: func(e event.TypedDeleteEvent[*appv1alpha1.Application]) bool {
				app := e.Object
				return app != nil && app.Annotations[gateway.AnnotationInCluster] != ""
			},
		},
	))
	if err != nil {
		return err
	}

	watches := []client.Object{
		&appsv1.Deployment{},
		&appsv1.StatefulSet{},
	}

	// watch the object installed by app-installer
	for _, w := range watches {
		if err = r.addWatch(c, mgr.GetCache(), w); err != nil {
			return err
		}
	}
	return nil
}

func (r *ApplicationReconciler) addWatch(c controller.Controller, cache cache.Cache, watchedObject client.Object) error {
	return c.Watch(source.Kind(
		cache,
		watchedObject,
		handler.EnqueueRequestsFromMapFunc(
			func(ctx context.Context, h client.Object) []reconcile.Request {
				appNames := getAppName(h)
				return []reconcile.Request{{NamespacedName: types.NamespacedName{
					Name:      strings.Join(appNames, ","),
					Namespace: h.GetNamespace()}}}
			}),
		predicate.Funcs{
			UpdateFunc: func(e event.UpdateEvent) bool {
				return isApp(e.ObjectNew, e.ObjectOld)
			},
			CreateFunc: func(e event.CreateEvent) bool {
				return isApp(e.Object)
			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				return isApp(e.Object)
			},
		},
	))
}

// TODO: get application other spec info
// TODO: make sure entrance service is applied
func (r *ApplicationReconciler) createApplication(ctx context.Context, req ctrl.Request,
	deployment client.Object, name string) error {
	owner := deployment.GetLabels()[constants.ApplicationOwnerLabel]
	appNames := getAppName(deployment)
	isMultiApp := len(appNames) > 1
	icon := getAppIcon(deployment)
	entrancesMap, err := r.getEntranceServiceAddress(ctx, deployment, isMultiApp)
	if err != nil {
		ctrl.Log.Error(err, "get entrance error")
	}
	servicePortsMap, err := r.getAppPorts(ctx, deployment, isMultiApp)
	if err != nil {
		klog.Warningf("get app ports err=%v", err)
	}
	tailScale, err := r.getAppTailScale(deployment)
	if err != nil {
		klog.Warningf("get app tailscale acls err=%v", err)
	}

	var appid string
	var isSysApp bool
	if userspace.IsSysApp(name) {
		appid = name
		isSysApp = true
	} else {
		appid = appcfg.AppName(name).GetAppID()
	}
	settings, sharedEntrances := r.getAppSettings(ctx, name, appid, owner, deployment, isMultiApp, entrancesMap[name])

	rawAppName := name
	if deployment.GetLabels()[constants.ApplicationRawAppNameLabel] != "" {
		rawAppName = deployment.GetLabels()[constants.ApplicationRawAppNameLabel]
	}
	appLabels := map[string]string{}
	if v, ok := deployment.GetLabels()[constants.AppApiVersionLabel]; ok && v != "" {
		appLabels[constants.AppApiVersionLabel] = v
	}
	// create the application cr
	newapp := &appv1alpha1.Application{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:   fmtAppName(name, req.Namespace),
			Labels: appLabels,
		},
		Spec: appv1alpha1.ApplicationSpec{
			Name:            name,
			RawAppName:      rawAppName,
			Appid:           appid,
			IsSysApp:        isSysApp,
			Namespace:       req.Namespace,
			Owner:           owner, // get from deployment
			DeploymentName:  deployment.GetName(),
			Entrances:       entrancesMap[name],
			SharedEntrances: sharedEntrances,
			Ports:           servicePortsMap[name],
			Icon:            icon[name],
			Settings:        settings,
		},
	}
	if tailScale != nil {
		newapp.Spec.TailScale = *tailScale
	}
	if err := gateway.ApplyRouteModeAnnotation(ctx, r.Client, newapp); err != nil {
		klog.Warningf("apply gateway route-mode for new app %s err=%v", name, err)
	}
	gateway.ApplyCallerInClusterAnnotation(newapp)
	app, err := r.AppClientset.AppV1alpha1().Applications().Create(ctx, newapp, metav1.CreateOptions{})
	if err != nil {
		ctrl.Log.Error(err, "create application error")
	}
	now := metav1.Now()
	appCopy := app.DeepCopy()
	if userspace.IsSysApp(app.Spec.Name) {
		err = apputils.CreateSysAppMgr(app.Spec.Name, app.Spec.Owner)
		if err != nil {
			klog.Errorf("Failed to create applicationmanagers for system app=%s err=%v", app.Spec.Name, err)
		}
	}

	app.Status.StatusTime = &now
	app.Status.UpdateTime = &now
	app.Status.State = appv1alpha1.AppNotReady.String()

	entranceStatues := make([]appv1alpha1.EntranceStatus, 0, len(app.Spec.Entrances))

	for _, e := range app.Spec.Entrances {
		if e.Skip {
			continue
		}
		state := appv1alpha1.EntranceNotReady
		if userspace.IsSysApp(app.Spec.Name) {
			state = appv1alpha1.EntranceRunning
		}
		entranceStatues = append(entranceStatues, appv1alpha1.EntranceStatus{
			Name:       e.Name,
			State:      state,
			StatusTime: &now,
			Reason:     state.String(),
		})
	}
	app.Status.EntranceStatuses = entranceStatues

	err = r.Status().Patch(ctx, app, client.MergeFrom(appCopy))
	if err != nil {
		klog.Infof("Failed to patch err=%v", err)
	}

	if srrErr := r.reconcileSharedRouteRegistry(ctx, app); srrErr != nil {
		klog.Warningf("reconcile SharedRouteRegistry for app=%s err=%v", app.Spec.Name, srrErr)
	}

	return err
}

func (r *ApplicationReconciler) updateApplication(ctx context.Context, req ctrl.Request,
	deployment client.Object, app *appv1alpha1.Application, name string) error {
	// Skip update if triggered by app modification (not deployment change)
	if app.Annotations != nil {
		if lastVersion := app.Annotations[deploymentResourceVersionAnnotation]; lastVersion == deployment.GetResourceVersion() {
			klog.Infof("skip updateApplication: deployment %s not changed, triggered by app modification", deployment.GetName())
			// Annotation changes (e.g. gateway.olares.io/route-mode opt-in/out)
			// don't bump the deployment resource version. Run the SRR
			// reconciler so toggling routeMode is declarative.
			if err := r.ensureAppGatewayRouteMode(ctx, app); err != nil {
				klog.Warningf("ensure gateway route-mode on app-only update for %s err=%v", app.Spec.Name, err)
			}
			if err := r.ensureCallerInClusterAnnotation(ctx, app); err != nil {
				klog.Warningf("ensure in-cluster annotation on app-only update for %s err=%v", app.Spec.Name, err)
			}
			if srrErr := r.reconcileSharedRouteRegistry(ctx, app); srrErr != nil {
				klog.Warningf("reconcile SharedRouteRegistry on app-only update for %s err=%v", app.Spec.Name, srrErr)
			}
			return nil
		}
	}

	appCopy := app.DeepCopy()
	appNames := getAppName(deployment)
	isMultiApp := len(appNames) > 1

	tailScale, err := r.getAppTailScale(deployment)
	if err != nil {
		klog.Errorf("failed to get tailscale err=%v", err)
	}

	owner := deployment.GetLabels()[constants.ApplicationOwnerLabel]
	klog.Infof("in updateApplication ....appname: %v", app.Spec.Name)
	icons := getAppIcon(deployment)
	var icon string

	icon = icons[name]

	entrancesMap, err := r.getEntranceServiceAddress(ctx, deployment, isMultiApp)
	if err != nil {
		ctrl.Log.Error(err, "get entrance error")
	}
	servicePortsMap, err := r.getAppPorts(ctx, deployment, isMultiApp)
	if err != nil {
		klog.Warningf("get app ports err=%v", err)
	}
	var appid string
	if userspace.IsSysApp(name) {
		appid = name
	} else {
		appid = appcfg.AppName(name).GetAppID()
	}
	settings, sharedEntrances := r.getAppSettings(ctx, name, appid, owner, deployment, isMultiApp, entrancesMap[name])

	appCopy.Spec.Name = name
	appCopy.Spec.Namespace = deployment.GetNamespace()
	appCopy.Spec.Owner = owner
	appCopy.Spec.DeploymentName = deployment.GetName()
	appCopy.Spec.Icon = icon
	appCopy.Spec.SharedEntrances = sharedEntrances
	appCopy.Spec.Ports = servicePortsMap[name]

	// Merge entrances: preserve authLevel from existing, update other fields
	appCopy.Spec.Entrances = mergeEntrances(app.Spec.Entrances, entrancesMap[name])

	if appCopy.Spec.Settings == nil {
		appCopy.Spec.Settings = make(map[string]string)
	}
	if settings["defaultThirdLevelDomainConfig"] != "" {
		appCopy.Spec.Settings["defaultThirdLevelDomainConfig"] = settings["defaultThirdLevelDomainConfig"]
	}

	if incomingPolicy := settings[applicationSettingsPolicyKey]; incomingPolicy != "" {
		existingPolicy := appCopy.Spec.Settings[applicationSettingsPolicyKey]
		appCopy.Spec.Settings[applicationSettingsPolicyKey] = mergePolicySettings(existingPolicy, incomingPolicy)
	}
	if settings["clusterScoped"] == "true" {
		appCopy.Spec.Settings["clusterScoped"] = "true"
		if settings["clusterAppRef"] != "" {
			appCopy.Spec.Settings["clusterAppRef"] = settings["clusterAppRef"]
		}
	}
	if settings[gateway.SettingGatewayRouteMode] != "" {
		appCopy.Spec.Settings[gateway.SettingGatewayRouteMode] = settings[gateway.SettingGatewayRouteMode]
	}
	if settings[gateway.SettingInClusterMode] != "" {
		appCopy.Spec.Settings[gateway.SettingInClusterMode] = settings[gateway.SettingInClusterMode]
	}

	if tailScale != nil {
		appCopy.Spec.TailScale = *tailScale
	}

	actionConfig, _, err := helm.InitConfig(r.Kubeconfig, appCopy.Spec.Namespace)
	if err != nil {
		ctrl.Log.Error(err, "init helm config error")
	}

	if !userspace.IsSysApp(app.Spec.Name) {
		version, _, err := apputils.GetDeployedReleaseVersion(actionConfig, name)
		if err != nil && !errors.Is(err, driver.ErrReleaseNotFound) {
			ctrl.Log.Error(err, "get deployed release version error")
		}
		if err == nil {
			appCopy.Spec.Settings["version"] = version
		}
	}

	// Record deployment resourceVersion to detect app-only modifications
	if appCopy.Annotations == nil {
		appCopy.Annotations = make(map[string]string)
	}
	klog.Infof("deploymentname: %s, version: %v", deployment.GetName(), deployment.GetResourceVersion())
	appCopy.Annotations[deploymentResourceVersionAnnotation] = deployment.GetResourceVersion()

	if err := gateway.ApplyRouteModeAnnotation(ctx, r.Client, appCopy); err != nil {
		klog.Warningf("apply gateway route-mode for app %s err=%v", appCopy.Spec.Name, err)
	}
	gateway.ApplyCallerInClusterAnnotation(appCopy)

	// Propagate the v3 marker from the deployment so the
	// Application CR carries it for downstream visibility / proxy fan-out.
	if v, ok := deployment.GetLabels()[constants.AppApiVersionLabel]; ok && v != "" {
		if appCopy.Labels == nil {
			appCopy.Labels = make(map[string]string)
		}
		appCopy.Labels[constants.AppApiVersionLabel] = v
	}

	err = r.Patch(ctx, appCopy, client.MergeFrom(app))
	if err != nil {
		klog.Infof("update spec failed %v", err)
		return err
	}

	klog.Infof("appCopy.Status: %v", appCopy.Status)
	newAppState := r.calAppState(&appCopy.Status)
	klog.Infof("application controller newAppState: %v", newAppState)
	klog.Infof("application controller oldAppState: %v", appCopy.Status.State)

	if appCopy.Status.State != newAppState {
		klog.Infof("set appCopy.State:.......new: %v", newAppState)
		appCopy.Status.State = newAppState
		now := metav1.Now()
		appCopy.Status.LastTransitionTime = &now

		err = r.Status().Patch(ctx, appCopy, client.MergeFrom(app))
		if err != nil {
			klog.Infof("update xxx error: %v", err)
			return err
		}
	}

	// merge settings
	//for k, v := range settings {
	//	if setting, ok := appCopy.Spec.Settings[k]; !ok || setting != v {
	//		appCopy.Spec.Settings[k] = v
	//	}
	//}

	//var a appv1alpha1.Application
	//err = r.Get(ctx, types.NamespacedName{Name: app.Name}, &a)
	//if err != nil {
	//	klog.Infof("get app failed %v", err)
	//	return err
	//}
	//klog.Infof("appState: ..%v", a.Status.State)
	if srrErr := r.reconcileSharedRouteRegistry(ctx, appCopy); srrErr != nil {
		klog.Warningf("reconcile SharedRouteRegistry for app=%s err=%v", appCopy.Spec.Name, srrErr)
	}
	r.reconcileCallerNamespace(ctx, appCopy)

	return err
}

func (r *ApplicationReconciler) reconcileCallerNamespace(ctx context.Context, app *appv1alpha1.Application) {
	if app == nil || app.Spec.Namespace == "" {
		return
	}
	cr := &routecontrol.CallerReconciler{Client: r.Client}
	if err := cr.Reconcile(ctx, app.Spec.Namespace); err != nil {
		klog.Warningf("caller reconciler for ns=%s app=%s err=%v", app.Spec.Namespace, app.Spec.Name, err)
	}
}

// reconcileGatewayRoutesForWorkloadNS runs reconcileSharedRouteRegistry for every
// cluster Application whose spec.namespace matches the reconciled workload namespace.
func (r *ApplicationReconciler) reconcileGatewayRoutesForWorkloadNS(ctx context.Context, workloadNS string) error {
	if workloadNS == "" {
		return nil
	}
	list, err := r.AppClientset.AppV1alpha1().Applications().List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	for i := range list.Items {
		app := &list.Items[i]
		if app.Spec.Namespace != workloadNS {
			continue
		}
		if err := r.ensureAppGatewayRouteMode(ctx, app); err != nil {
			return fmt.Errorf("app %s route-mode: %w", app.Name, err)
		}
		if err := r.ensureCallerInClusterAnnotation(ctx, app); err != nil {
			return fmt.Errorf("app %s in-cluster: %w", app.Name, err)
		}
		if err := r.reconcileSharedRouteRegistry(ctx, app); err != nil {
			return fmt.Errorf("app %s: %w", app.Name, err)
		}
		r.reconcileCallerNamespace(ctx, app)
	}
	return nil
}

func inClusterAnnotationChanged(oldApp, newApp *appv1alpha1.Application) bool {
	if oldApp == nil || newApp == nil {
		return false
	}
	oldV := ""
	if oldApp.Annotations != nil {
		oldV = oldApp.Annotations[gateway.AnnotationInCluster]
	}
	newV := ""
	if newApp.Annotations != nil {
		newV = newApp.Annotations[gateway.AnnotationInCluster]
	}
	return oldV != newV
}

// ensureAppGatewayRouteMode persists gateway.olares.io/route-mode when the
// automation policy (ClusterConfig + manifest settings) requires it. Explicit
// operator annotations (gateway or direct) are never overwritten.
func (r *ApplicationReconciler) ensureAppGatewayRouteMode(ctx context.Context, app *appv1alpha1.Application) error {
	if app == nil {
		return nil
	}
	need, mode, err := gateway.ComputeRouteModePatch(ctx, r.Client, app)
	if err != nil || !need {
		return err
	}
	appCopy := app.DeepCopy()
	if appCopy.Annotations == nil {
		appCopy.Annotations = map[string]string{}
	}
	appCopy.Annotations[gateway.AnnotationRouteMode] = mode
	updated, err := r.AppClientset.AppV1alpha1().Applications().Update(ctx, appCopy, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	if app.Annotations == nil {
		app.Annotations = map[string]string{}
	}
	app.Annotations[gateway.AnnotationRouteMode] = updated.Annotations[gateway.AnnotationRouteMode]
	return nil
}

// ensureCallerInClusterAnnotation persists gateway.olares.io/in-cluster=gateway
// for callers with clusterAppRef when the manifest or defaults require it.
func (r *ApplicationReconciler) ensureCallerInClusterAnnotation(ctx context.Context, app *appv1alpha1.Application) error {
	_ = ctx
	if app == nil {
		return nil
	}
	need, value := gateway.ComputeCallerInClusterPatch(app)
	if !need {
		return nil
	}
	appCopy := app.DeepCopy()
	if appCopy.Annotations == nil {
		appCopy.Annotations = map[string]string{}
	}
	appCopy.Annotations[gateway.AnnotationInCluster] = value
	updated, err := r.AppClientset.AppV1alpha1().Applications().Update(ctx, appCopy, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	if app.Annotations == nil {
		app.Annotations = map[string]string{}
	}
	app.Annotations[gateway.AnnotationInCluster] = updated.Annotations[gateway.AnnotationInCluster]
	return nil
}

// reconcileSharedRouteRegistry writes / deletes the SharedRouteRegistry that
// declares this shared app's exposure through the shared Envoy Gateway data plane.
// Only acts when the Application carries gateway.olares.io/route-mode=gateway.
// Qualifying apps are v3 installs or v2 cluster-scoped apps with
// spec.sharedEntrances (see appcfg.IsGatewaySharedApp). For any other case
// (no shared entrances, no annotation, direct mode) the SRR is removed so
// toggling routeMode is truly declarative.
func (r *ApplicationReconciler) reconcileSharedRouteRegistry(ctx context.Context, app *appv1alpha1.Application) error {
	if app == nil || app.Spec.Namespace == "" || app.Spec.Name == "" {
		return nil
	}
	if !appcfg.IsGatewaySharedApp(app) {
		klog.V(2).Infof("SRR skip app=%s: not a gateway shared app (need v3 label or clusterScoped+sharedEntrances)", app.Spec.Name)
		return gateway.DeleteAllForApp(ctx, r.Client, app)
	}
	if !gateway.IsOptedIn(app) {
		klog.V(2).Infof("SRR skip app=%s: route-mode is not gateway", app.Spec.Name)
		return gateway.DeleteAllForApp(ctx, r.Client, app)
	}
	klog.Infof("SRR reconcile start app=%s ns=%s entrances=%d", app.Spec.Name, app.Spec.Namespace, len(app.Spec.SharedEntrances))

	// Remove legacy "shared-<appName>" SRR if present, then write one SRR per
	// sharedEntrance with logical hostPattern <hash8>.*.<platformDomain>.
	if err := gateway.Delete(ctx, r.Client, app); err != nil {
		return fmt.Errorf("remove legacy SRR: %w", err)
	}

	platformDomain := cluster.GetPlatformDomain(ctx)
	if platformDomain == "" {
		return fmt.Errorf("platformDomain is empty (ClusterConfig missing and env unset)")
	}

	appid := strings.TrimSpace(app.Spec.Appid)
	if appid == "" {
		appid = appcfg.AppName(app.Spec.Name).GetAppID()
	}

	// Track which entrance SRRs should exist after this pass; everything else
	// owned by the Application gets cleaned up to handle entrance removals.
	desired := make(map[string]struct{}, len(app.Spec.SharedEntrances))

	for i := range app.Spec.SharedEntrances {
		entrance := app.Spec.SharedEntrances[i]
		if entrance.Name == "" {
			klog.Warningf("SRR skip: app=%s entrance#%d has empty name", app.Spec.Name, i)
			continue
		}
		if entrance.Host == "" {
			return fmt.Errorf("shared entrance %q on app %s has empty host", entrance.Name, app.Spec.Name)
		}
		svc, err := gateway.ResolveSharedEntranceService(ctx, r.Client, app, entrance.Host)
		if err != nil {
			return fmt.Errorf("resolve backing service for entrance %q: %w", entrance.Name, err)
		}
		spec, err := gateway.BuildSpecForEntrance(app, entrance, svc, platformDomain)
		if err != nil {
			return fmt.Errorf("build SRR spec for entrance %q: %w", entrance.Name, err)
		}
		name := gateway.ResourceNameForEntrance(appid, entrance.Name)
		if err := gateway.CheckLogicalPatternUniqueness(ctx, r.Client, spec.HostPatterns[0], app.Spec.Namespace, name); err != nil {
			return fmt.Errorf("uniqueness check for entrance %q: %w", entrance.Name, err)
		}
		srrObj, err := gateway.ReconcileForEntrance(ctx, r.Client, app, entrance, spec)
		if err != nil {
			return err
		}
		desired[name] = struct{}{}
		klog.V(1).Infof("SRR reconciled app=%s/%s entrance=%s name=%s hostPatterns=%v upstream=%s/%s:%d",
			app.Spec.Namespace, app.Spec.Name, entrance.Name, name, spec.HostPatterns,
			spec.Upstream.ServiceNamespace, spec.Upstream.ServiceName, spec.Upstream.Port)

		// Shared ingress route control: after the SRR is written, app-service
		// ensures the HTTPRoute and NetworkPolicy for this entrance exist in
		// the Application namespace. Route apply errors are recorded on
		// SRR.status and do not fail the Application reconcile loop.
		routeRes, routeErr := routecontrol.ReconcileSharedRoute(ctx, r.Client, routecontrol.GatewayRef{}, srrObj)
		if routeErr != nil {
			klog.Warningf("reconcile shared route %s/%s failed: %v", srrObj.Namespace, srrObj.Name, routeErr)
			routeRes = routecontrol.ReconcileResult{
				Status:  metav1.ConditionFalse,
				Reason:  routecontrol.ReasonRouteApplyFailed,
				Message: routeErr.Error(),
			}
		}
		if statusErr := routecontrol.UpdateSRRStatus(ctx, r.Client, srrObj, routeRes); statusErr != nil {
			klog.Warningf("update SRR route status %s/%s failed: %v", srrObj.Namespace, srrObj.Name, statusErr)
		}
	}

	// Garbage-collect stale per-entrance SRRs (e.g. when sharedEntrances was
	// trimmed in a Helm upgrade). OwnerReferences ultimately delete on app
	// removal, but this keeps an opted-in Application's SRR set tight while
	// the Application still exists.
	if err := gateway.PruneEntranceSRRs(ctx, r.Client, app, desired); err != nil {
		return fmt.Errorf("prune stale SRRs: %w", err)
	}
	return nil
}

func (r *ApplicationReconciler) getEntranceServiceAddress(ctx context.Context, deployment client.Object, isMultiApp bool) (map[string][]appv1alpha1.Entrance, error) {
	entrancesLabel := deployment.GetAnnotations()[constants.ApplicationEntrancesKey]
	entrancesMap := make(map[string][]appv1alpha1.Entrance)

	if len(entrancesLabel) == 0 {
		return entrancesMap, errors.New("invalid service address label")
	}
	klog.Infof("isMultiApp: %v", isMultiApp)
	var err error
	if isMultiApp {
		err = json.Unmarshal([]byte(entrancesLabel), &entrancesMap)
		if err != nil {
			klog.Infof("unmarshalMAp error=%v", err)
			return nil, err
		}
	} else {
		appName := deployment.GetLabels()[constants.ApplicationNameLabel]
		entrances := make([]appv1alpha1.Entrance, 0)
		err = json.Unmarshal([]byte(entrancesLabel), &entrances)
		if err != nil {
			klog.Infof("unmarshal error=%v", err)
			return nil, err
		}
		entrancesMap[appName] = entrances
	}

	// set default value and check if service exists
	for _, entrances := range entrancesMap {
		for i, e := range entrances {
			if e.AuthLevel == "" {
				entrances[i].AuthLevel = constants.AuthorizationLevelOfPrivate
			}
			if e.OpenMethod == "" {
				entrances[i].OpenMethod = "default"
			}
			objectKey := types.NamespacedName{Namespace: deployment.GetNamespace(), Name: e.Host}
			var svc corev1.Service
			if err = r.Get(ctx, objectKey, &svc); err == nil {
				if !checkPortOfService(&svc, e.Port) {
					return nil, fmt.Errorf("entrance: %s not found", e.Host)
				}
			} else {
				return nil, err
			}
		}
	}
	return entrancesMap, nil
}

func (r *ApplicationReconciler) getAppSettings(ctx context.Context, appName, appId, owner string, deployment client.Object,
	isMulti bool, entrances []appv1alpha1.Entrance) (settings map[string]string, sharedEntrances []appv1alpha1.Entrance) {
	settings = make(map[string]string)
	settings["source"] = api.Unknown.String()
	rawAppName := appName
	if deployment.GetLabels()[constants.ApplicationRawAppNameLabel] != "" {
		rawAppName = deployment.GetLabels()[constants.ApplicationRawAppNameLabel]
	}

	if chartSource, ok := deployment.GetAnnotations()[constants.ApplicationSourceLabel]; ok {
		settings["source"] = chartSource
	}

	if marketSource, ok := deployment.GetAnnotations()[constants.AppMarketSourceKey]; ok {
		settings["market_source"] = marketSource
	}

	if systemService, ok := deployment.GetLabels()[constants.ApplicationSystemServiceLabel]; ok {
		settings["system_service"] = systemService
	}

	titles := getAppTitle(deployment)
	settings["title"] = titles[appName]

	if target, ok := deployment.GetLabels()[constants.ApplicationTargetLabel]; ok {
		settings["target"] = target
	}

	versions := getAppVersion(deployment)
	settings["version"] = versions[appName]

	settings["clusterScoped"] = "false"
	settings["requiredGPU"] = deployment.GetAnnotations()[constants.ApplicationRequiredGPU]
	//clusterScoped, ok := deployment.GetAnnotations()[constants.ApplicationClusterScoped]
	//if ok && clusterScoped == "true" {
	//	settings["clusterScoped"] = "true"
	//}

	if defaultDomainAnnotation, ok := deployment.GetAnnotations()[constants.ApplicationDefaultThirdLevelDomain]; ok {
		var allDomainConfigs []appv1alpha1.DefaultThirdLevelDomainConfig
		err := json.Unmarshal([]byte(defaultDomainAnnotation), &allDomainConfigs)
		if err != nil {
			klog.Errorf("Failed to unmarshal default domain annotation err=%v", err)
		} else {
			var appDomainConfigs []appv1alpha1.DefaultThirdLevelDomainConfig
			for _, config := range allDomainConfigs {
				if config.AppName == appName {
					appDomainConfigs = append(appDomainConfigs, config)
				}
			}

			if len(appDomainConfigs) > 0 {
				domainConfigBytes, err := json.Marshal(appDomainConfigs)
				if err != nil {
					klog.Errorf("Failed to marshal domain configs err=%v", err)
				} else {
					settings["defaultThirdLevelDomainConfig"] = string(domainConfigBytes)
				}
			}
		}
	}

	// not sys applications.
	if !userspace.IsSysApp(rawAppName) {
		if appCfg, err := appcfg.GetAppInstallationConfig(appName, owner); err != nil {
			klog.Infof("Failed to get app configuration appName=%s owner=%s err=%v", appName, owner, err)
		} else {
			policyStr, err := getApplicationPolicy(appCfg.Policies, appCfg.Entrances)
			if err != nil {
				klog.Errorf("Failed to encode json err=%v", err)
			} else if len(policyStr) > 0 {
				settings[applicationSettingsPolicyKey] = policyStr
			}

			// set cluster-scoped info to settings
			if appCfg.AppScope.ClusterScoped {
				settings["clusterScoped"] = "true"
				if len(appCfg.AppScope.AppRef) > 0 {
					settings["clusterAppRef"] = strings.Join(appCfg.AppScope.AppRef, ",")
				}

				sharedEntrances = appCfg.SharedEntrances
			}
			if mode := strings.TrimSpace(appCfg.GatewayRouteMode); mode != "" {
				settings[gateway.SettingGatewayRouteMode] = strings.ToLower(mode)
			}
			if mode := strings.TrimSpace(appCfg.InClusterMode); mode != "" {
				settings[gateway.SettingInClusterMode] = strings.ToLower(mode)
			}
			if appCfg.MobileSupported {
				settings["mobileSupported"] = "true"
			} else {
				settings["mobileSupported"] = "false"
			}

			if appCfg.OIDC.Enabled {
				// get oidc client id and secret created at installing
				var secret corev1.Secret
				err = r.Get(ctx,
					types.NamespacedName{Namespace: deployment.GetNamespace(), Name: constants.OIDCSecret},
					&secret)
				if err != nil {
					klog.Errorf("Failed to get app's oidc secret err=%v, app=%s, namespace=%s", err, appName, deployment.GetNamespace())
				} else {
					settings["oidc.client.id"] = string(secret.Data["id"])

					encryptSecret, err := utils.Pbkdf2Crypto(string(secret.Data["secret"]))
					if err != nil {
						klog.Error("encrypt secret error, ", err)
					}
					settings["oidc.client.secret"] = encryptSecret

					zone, err := kubesphere.GetUserZone(ctx, owner)
					if err != nil {
						klog.Error("get user zone error, ", err)
					} else {

						multiEntrance := len(appCfg.Entrances) > 1
						for i, e := range appCfg.Entrances {
							if e.Name == appCfg.OIDC.EntranceName {
								var appUrl string
								if multiEntrance {
									appUrl = fmt.Sprintf("https://%s%d.%s%s", appId, i, zone, appCfg.OIDC.RedirectUri)
								} else {
									appUrl = fmt.Sprintf("https://%s.%s%s", appId, zone, appCfg.OIDC.RedirectUri)
								}
								settings["oidc.client.redirect_uri"] = appUrl
							}
						}

					} // end of if get zone
				} // end of if get secret
			}
		}
	} else {
		// sys applications.
		type Policies struct {
			Policies []appcfg.Policy `json:"policies"`
		}
		applicationPoliciesFromAnnotation, ok := deployment.GetAnnotations()[constants.ApplicationPolicies]

		var policy Policies
		if ok {
			if isMulti {
				m := make(map[string]Policies)
				err := json.Unmarshal([]byte(applicationPoliciesFromAnnotation), &m)
				if err != nil {
					klog.Errorf("Failed to unmarshal applicationPoliciesFromAnnotation err=%v", err)
				}
				policy = m[appName]
			} else {
				err := json.Unmarshal([]byte(applicationPoliciesFromAnnotation), &policy)
				if err != nil {
					klog.Errorf("Failed to unmarshal applicationPoliciesFromAnnotation err=%v", err)
				}
			}
		}
		klog.Infof("applicationPoliciesFromAnnotation: %s", applicationPoliciesFromAnnotation)
		klog.Infof("policy: %#v", policy)

		// transform from Policy to AppPolicy
		var appPolicies []appcfg.AppPolicy
		for _, p := range policy.Policies {
			d, _ := time.ParseDuration(p.Duration)
			appPolicies = append(appPolicies, appcfg.AppPolicy{
				EntranceName: p.EntranceName,
				URIRegex:     p.URIRegex,
				Level:        p.Level,
				OneTime:      p.OneTime,
				Duration:     d,
			})
		}
		policyStr, err := getApplicationPolicy(appPolicies, entrances)
		if err != nil {
			klog.Errorf("Failed to encode json err=%v", err)
		} else if len(policyStr) > 0 {
			settings[applicationSettingsPolicyKey] = policyStr
		}
		settings["source"] = api.System.String()
		mobileSupported, ok := deployment.GetAnnotations()[constants.ApplicationMobileSupported]
		settings["mobileSupported"] = "false"
		if ok {
			settings["mobileSupported"] = mobileSupported
		}
	}

	return
}

func (r *ApplicationReconciler) clearHelmHistory(appname, namespace string) error {
	actionConfig, _, err := helm.InitConfig(r.Kubeconfig, namespace)
	if err != nil {
		return err
	}
	klog.Infof("clearHelmHistory: appname:%s, namespace:%s", appname, namespace)

	histClient := action.NewHistory(actionConfig)
	histClient.Max = 1
	_, err = histClient.Run(appname)
	klog.Infof("appname in clearHelmHistory: %v", appname)
	klog.Infof("err in clearHelmHistory: err=%v", err)

	if err != nil {
		if errors.Is(err, driver.ErrReleaseNotFound) {
			return nil
		}
		return err
	}

	return helm.UninstallCharts(actionConfig, appname)
}

func (r *ApplicationReconciler) getAppPorts(ctx context.Context, deployment client.Object, isMultiApp bool) (map[string][]appv1alpha1.ServicePort, error) {
	portsLabel := deployment.GetAnnotations()[constants.ApplicationPortsKey]
	portsMap := make(map[string][]appv1alpha1.ServicePort)
	if len(portsLabel) == 0 {
		return portsMap, errors.New("invalid service port")
	}
	var err error
	if isMultiApp {
		err = json.Unmarshal([]byte(portsLabel), &portsMap)
		if err != nil {
			klog.Errorf("unmarshal portMap err=%v", err)
			return nil, err
		}
	} else {
		appName := deployment.GetLabels()[constants.ApplicationNameLabel]
		ports := make([]appv1alpha1.ServicePort, 0)
		err = json.Unmarshal([]byte(portsLabel), &ports)
		if err != nil {
			klog.Errorf("unmarshal service port error=%v", err)
			return nil, err
		}
		portsMap[appName] = ports
	}
	return portsMap, nil
}

func (r *ApplicationReconciler) getAppTailScale(deployment client.Object) (*appv1alpha1.TailScale, error) {
	tailScale := appv1alpha1.TailScale{}
	tailScaleString := deployment.GetAnnotations()[constants.ApplicationTailScaleKey]
	if len(tailScaleString) == 0 {
		return nil, nil
	}
	err := json.Unmarshal([]byte(tailScaleString), &tailScale)
	if err != nil {
		return nil, err
	}
	return &tailScale, nil
}

func (r *ApplicationReconciler) calAppState(status *appv1alpha1.ApplicationStatus) string {
	entranceLen := len(status.EntranceStatuses)
	klog.Infof("entranceLen: %v", entranceLen)
	if entranceLen == 0 {
		return "running"
	}
	for _, es := range status.EntranceStatuses {
		if es.State == appv1alpha1.EntranceStopped {
			return "stopped"
		}
		if es.State == appv1alpha1.EntranceNotReady {
			return "notReady"
		}
	}
	return "running"
}

func checkPortOfService(s *corev1.Service, port int32) bool {
	for _, p := range s.Spec.Ports {
		if p.Port == port {
			return true
		}
	}

	return false
}

func fmtAppName(name, namespace string) string {
	return appv1alpha1.AppResourceName(name, namespace)
}

func isApp(obs ...metav1.Object) bool {
	for _, o := range obs {

		if o.GetLabels() == nil {
			return false
		}

		if _, ok := o.GetLabels()[constants.ApplicationNameLabel]; !ok {
			return false
		}
	}
	return true
}

func isWorkflow(obs ...metav1.Object) bool {
	for _, o := range obs {

		if o.GetLabels() == nil {
			return false
		}

		if _, ok := o.GetLabels()[constants.WorkflowNameLabel]; !ok {
			return false
		}
	}
	return true
}

func getApplicationPolicy(policies []appcfg.AppPolicy, entrances []appv1alpha1.Entrance) (string, error) {
	subPolicy := make(map[string][]*applicationSettingsSubPolicy)

	for _, p := range policies {
		subPolicy[p.EntranceName] = append(subPolicy[p.EntranceName],
			&applicationSettingsSubPolicy{
				URI:      p.URIRegex,
				Policy:   p.Level,
				OneTime:  p.OneTime,
				Duration: int32(p.Duration / time.Second),
			})
	}

	policy := make(map[string]applicationSettingsPolicy)
	for _, e := range entrances {
		defaultPolicy := "system"
		sp := subPolicy[e.Name]
		if e.AuthLevel == constants.AuthorizationLevelOfPublic {
			defaultPolicy = constants.AuthorizationLevelOfPublic
		}
		policy[e.Name] = applicationSettingsPolicy{
			DefaultPolicy: defaultPolicy,
			OneTime:       false,
			Duration:      0,
			SubPolicies:   sp,
		}
	}

	policyStr, err := json.Marshal(policy)
	if err != nil {
		return "", err
	}
	return string(policyStr), nil
}

func getEntranceFromAnnotations(deployment client.Object) ([]appv1alpha1.Entrance, error) {
	entrancesLabel := deployment.GetAnnotations()[constants.ApplicationEntrancesKey]
	entrances := make([]appv1alpha1.Entrance, 0)

	if len(entrancesLabel) == 0 {
		return entrances, errors.New("invalid service address label")
	}

	if err := json.Unmarshal([]byte(entrancesLabel), &entrances); err != nil {
		return entrances, err
	}
	for i, e := range entrances {
		if e.OpenMethod == "" {
			entrances[i].OpenMethod = "default"
		}
	}

	return entrances, nil
}

func getAppName(deployment client.Object) []string {
	names := make([]string, 0)
	isMultiApp := deployment.GetLabels()[constants.ApplicationAppGroupLabel] == "true"
	if isMultiApp {
		apps := make(map[string]interface{})
		keys := deployment.GetAnnotations()[constants.ApplicationEntrancesKey]
		if keys == "" {
			klog.Infof("Application entrances label is empty")
			return nil
		}
		// multi-app in one deployment/statefulset, get all app names
		err := json.Unmarshal([]byte(keys), &apps)
		if err != nil {
			klog.Infof("Failed to unmarshal application entrances label err=%v", err)
			return nil
		}
		for k := range apps {
			names = append(names, k)
		}
		return names
	}
	name := deployment.GetLabels()[constants.ApplicationNameLabel]
	if name == "" {
		return nil
	}
	return []string{name}
}

func getAppIcon(deployment client.Object) map[string]string {
	ret := make(map[string]string)
	if deployment.GetLabels()[constants.ApplicationAppGroupLabel] == "true" {
		err := json.Unmarshal([]byte(deployment.GetAnnotations()[constants.ApplicationIconLabel]), &ret)
		if err != nil {
			klog.Infof("Failed to unmarshal application icon label err=%v", err)
		}
	} else {
		ret[deployment.GetLabels()[constants.ApplicationNameLabel]] = deployment.GetAnnotations()[constants.ApplicationIconLabel]
	}
	return ret
}

func getAppVersion(deployment client.Object) map[string]string {
	ret := make(map[string]string)
	if deployment.GetLabels()[constants.ApplicationAppGroupLabel] == "true" {
		err := json.Unmarshal([]byte(deployment.GetAnnotations()[constants.ApplicationVersionLabel]), &ret)
		if err != nil {
			klog.Infof("Failed to unmarshal application icon label err=%v", err)
		}
	} else {
		ret[deployment.GetLabels()[constants.ApplicationNameLabel]] = deployment.GetAnnotations()[constants.ApplicationVersionLabel]
	}
	return ret
}

func getAppTitle(deployment client.Object) map[string]string {
	ret := make(map[string]string)
	if deployment.GetLabels()[constants.ApplicationAppGroupLabel] == "true" {
		err := json.Unmarshal([]byte(deployment.GetAnnotations()[constants.ApplicationTitleLabel]), &ret)
		if err != nil {
			klog.Infof("Failed to unmarshal application icon label err=%v", err)
		}
	} else {
		ret[deployment.GetLabels()[constants.ApplicationNameLabel]] = deployment.GetAnnotations()[constants.ApplicationTitleLabel]
	}
	return ret
}

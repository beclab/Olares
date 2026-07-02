package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	iamv1alpha2 "github.com/beclab/api/iam/v1alpha2"

	"github.com/beclab/Olares/framework/app-service/controllers"
	"github.com/beclab/Olares/framework/app-service/pkg/apiserver"
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/appinstaller"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/adapters/computeops"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/adapters/computevalidator"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/adapters/ctrlrequeue"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/adapters/helmops"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/adapters/imgpull"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/adapters/k8sstore"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/adapters/mwops"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/adapters/natsrecorder"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/adapters/realclock"
	lifecycleapi "github.com/beclab/Olares/framework/app-service/pkg/applifecycle/apiserver"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/handlers"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/runner"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/runner/ctrladapter"
	appevent "github.com/beclab/Olares/framework/app-service/pkg/event"
	"github.com/beclab/Olares/framework/app-service/pkg/gateway"
	"github.com/beclab/Olares/framework/app-service/pkg/gateway/callerjwt"
	"github.com/beclab/Olares/framework/app-service/pkg/gateway/meshinagent"
	"github.com/beclab/Olares/framework/app-service/pkg/gateway/routecontrol"
	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/images"
	"github.com/beclab/Olares/framework/app-service/pkg/mesh"
	"github.com/beclab/Olares/framework/app-service/pkg/middlewareinstaller"
	"github.com/beclab/Olares/framework/app-service/pkg/webhook"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	sysv1alpha1 "github.com/beclab/api/api/sys.bytetrade.io/v1alpha1"
	"github.com/beclab/api/pkg/generated/clientset/versioned"

	kbappsv1 "github.com/apecloud/kubeblocks/apis/apps/v1"
	kbopv1alphav1 "github.com/apecloud/kubeblocks/apis/operations/v1alpha1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	//"k8s.io/client-go/dynamic"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(appv1alpha1.AddToScheme(scheme))
	utilruntime.Must(sysv1alpha1.AddToScheme(scheme))
	utilruntime.Must(iamv1alpha2.AddToScheme(scheme))
	utilruntime.Must(kbappsv1.AddToScheme(scheme))
	utilruntime.Must(kbopv1alphav1.AddToScheme(scheme))
	utilruntime.Must(srrv1alpha1.AddToScheme(scheme))

	//+kubebuilder:scaffold:scheme
}

var shutdownSignals = []os.Signal{os.Interrupt, syscall.SIGTERM}

const (
	kubeSphereHostAddr = "KS_APISERVER_SERVICE_HOST" // env name in cluster
	kubeSphereHostPort = "KS_APISERVER_SERVICE_PORT"
)

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":6080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":6081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		Development: true,
		TimeEncoder: zapcore.RFC3339TimeEncoder,
	}
	opts.BindFlags(flag.CommandLine)
	// Most of this codebase logs through klog, and a good part of its
	// diagnostics sit behind klog.V(2)/V(4). Without registering klog's flags
	// the verbosity is pinned at 0 and there is no way to reach them at all, so
	// those lines could never be turned on in a running cluster.
	klog.InitFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	config := ctrl.GetConfigOrDie()

	mgr, err := ctrl.NewManager(config, ctrl.Options{
		Scheme:                 scheme,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "5117a667.bytetrade.io",
		// LeaderElectionReleaseOnCancel defines if the leader should step down voluntarily
		// when the Manager ends. This requires the binary to immediately end when the
		// Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
		// speeds up voluntary leader transitions as the new leader don't have to wait
		// LeaseDuration time first.
		//
		// In the default scaffold provided, the program ends immediately after
		// the manager stops, so would be fine to enable this option. However,
		// if you are doing or is intended to do any operation such as perform cleanups
		// after the manager stops then its usage might be unsafe.
		// LeaderElectionReleaseOnCancel: true,
	})
	if err != nil {
		setupLog.Error(err, "Unable to start manager")
		os.Exit(1)
	}

	// initialize process environment variables from existing SystemEnv CRs
	initEnvClient, initEnvErr := client.New(config, client.Options{Scheme: scheme})
	if initEnvErr != nil {
		setupLog.Error(initEnvErr, "Unable to create uncached client for SystemEnv initialization")
		os.Exit(1)
	}
	initEnvCtx, cancelInitEnvCtx := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelInitEnvCtx()
	if initEnvErr := controllers.InitializeSystemEnvProcessEnv(initEnvCtx, initEnvClient); initEnvErr != nil {
		setupLog.Error(initEnvErr, "Failed to initialize process env from SystemEnv")
		os.Exit(1)
	}

	appClient := versioned.NewForConfigOrDie(config)
	kubeClient := kubernetes.NewForConfigOrDie(config)
	ictx, cancelFunc := context.WithCancel(context.Background())

	rolloutWorker := meshinagent.NewRolloutWorker(mgr.GetClient(), meshinagent.DefaultRolloutQueue)
	meshinagent.SetDefaultWorker(rolloutWorker)
	meshinagent.SetMeshControlPlaneReadyCheck(mesh.IsControlPlaneReady)
	rolloutWorker.Start(ictx)

	if err = (&controllers.ApplicationReconciler{
		Client:       mgr.GetClient(),
		Scheme:       mgr.GetScheme(),
		AppClientset: appClient,
		Kubeconfig:   config,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "Unable to create controller", "controller", "Application")
		os.Exit(1)
	}

	if err = (&controllers.OverlayMACLifecycleReconciler{
		Kube:          kubeClient,
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		DynamicClient: dynamic.NewForConfigOrDie(config),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "Unable to create controller", "controller", "OverlayMACLifecycle")
		os.Exit(1)
	}

	if err = (&controllers.MeshInjectRolloutReconciler{
		Client: mgr.GetClient(),
		Kube:   kubeClient,
		Worker: rolloutWorker,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "Unable to create controller", "controller", "MeshInjectRollout")
		os.Exit(1)
	}

	if err = (&controllers.SecurityReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		DynamicClient: dynamic.NewForConfigOrDie(config),
	}).SetupWithManager(ictx, mgr); err != nil {
		setupLog.Error(err, "Unable to create controller", "controller", "Security")
		os.Exit(1)
	}
	appEventQueue := appevent.NewAppEventQueue(ictx, nil)
	appevent.SetAppEventQueue(appEventQueue)
	go appEventQueue.Run()

	defer func() {
		if nc := appEventQueue.GetNatsConn(); nc != nil {
			nc.Drain()
		}
	}()

	// ---------------------------------------------------------------
	// applifecycle Runner wiring — LIVE MODE (cutover complete).
	//
	// After this block the Runner is the AUTHORITATIVE writer of
	// every ApplicationManager.Status field. The legacy
	// ApplicationManagerController is still registered below but its
	// Reconcile method is now a no-op (see controllers/appmgr_controller.go
	// for the retirement rationale). The only piece of legacy logic
	// still active is preEnqueueCheckForUpdate's publishStateChangeEvent
	// side effect — a state-change → NATS bridge that bfl depends on.
	// Runner-driven state changes flow through the same code path
	// because they also produce AM Update events on the informer.
	//
	// Ports wiring (all REAL adapters):
	//
	//   * AMStore   → k8sstore(live)   — Runner reads AND writes AM.
	//                 Shadow mode is off; Runner-emitted PatchStatus
	//                 effects now persist to real k8s.
	//   * Clock     → realclock         — time.Now / time.After.
	//   * Requeuer  → runnerRequeuer    — shared with OpDispatcher so
	//                 HTTP TriggerOp deliveries reach the same
	//                 workqueue as async CallWorkload callbacks.
	//   * WorkloadOps → helmops         — thin passthrough over the
	//                 legacy versioned.NewHelmOps pipeline. All 9
	//                 write methods + WaitForNamespaceDeleted come
	//                 through this adapter now.
	//   * Validator → computevalidator  — wraps pkg/compute/validation
	//                 chain; NotifyInsufficientCompute uses the
	//                 recorder below via a small closure so payload
	//                 shape is byte-identical to legacy.
	//   * ImagePuller → imgpull         — wraps images.ImageManagerClient
	//                 for CreateJob/CancelJob/UpdateJobStatus and
	//                 provides a pure Progress reader. The RefResolver
	//                 closure runs a helm dry-run through
	//                 appinstaller.BuildBaseHelmValues +
	//                 appstate.GetRefsForImageManager (matches
	//                 defaultResolveImageRefs in pkg/appstate/deps.go
	//                 verbatim).
	//   * Recorder  → natsrecorder      — publishes AppLifecycleEvent
	//                 and generic notifications on the SAME
	//                 QueuedEventController the legacy
	//                 publishStateChangeEvent enqueues into. bfl sees
	//                 no schema change; only the origin of the events
	//                 shifts as legacy handlers stop firing.
	//
	// The transition table is populated end-to-end: NewDefaultTable
	// seeds every legacy edge with PlaceholderHandler so the topology
	// is dispatch-complete on day one; handlers.Register overwrites
	// every state family that has a real ported handler (35 of 37
	// states — the sole survivor is the InitialState "" → Pending
	// bootstrap edge, a pure state advance with no side effects).
	// ---------------------------------------------------------------
	runnerRequeuer := ctrlrequeue.New()
	runnerTable := domain.DefaultTable
	handlers.Register(runnerTable)
	runnerEngine := domain.NewEngine(runnerTable)

	// Recorder is constructed FIRST so computevalidator can borrow
	// its RecordNotification method for the "insufficient compute"
	// notification — one publisher, one queue, one NATS subject
	// contract.
	runnerRecorder := natsrecorder.New(appEventQueue)

	// notify is the seam computevalidator's NotifyInsufficientCompute
	// delegates to. Sends the pkg/compute.BuildInsufficientPayload
	// output to NATS via the recorder's generic notification path.
	notify := func(ctx context.Context, subject string, payload any) error {
		return runnerRecorder.RecordNotification(ctx, subject, payload)
	}

	// refResolver is the seam imgpull's ResolveRefs delegates to.
	// Byte-identical to pkg/appstate/deps.go's defaultResolveImageRefs
	// — factored inline here so the imgpull adapter can be wired
	// without a hard dependency on pkg/appstate (which is being
	// deleted in the follow-up cleanup pass).
	refResolver := func(ctx context.Context, am *appv1alpha1.ApplicationManager, cfg *appcfg.ApplicationConfig) ([]appv1alpha1.Ref, error) {
		values, err := appinstaller.BuildBaseHelmValues(ctx, config, cfg, am.Spec.AppOwner, true)
		if err != nil {
			return nil, err
		}
		return images.GetRefsForImageManager(cfg, values)
	}

	mongoUninstall := func(ctx context.Context, cfg *middlewareinstaller.MiddlewareConfig) error {
		return middlewareinstaller.Uninstall(ctx, config, cfg)
	}

	appRunner, err := runner.New(runner.Config{
		Engine: runnerEngine,
		// LIVE k8sstore. Runner is now the sole writer of
		// AM.Status. Legacy controller's Reconcile is gutted;
		// see controllers/appmgr_controller.go.
		AMStore:       k8sstore.New(mgr.GetClient()),
		WorkloadOps:   helmops.New(mgr.GetClient(), config),
		Validator:     computevalidator.New(mgr.GetClient(), notify),
		ImagePuller:   imgpull.New(mgr.GetClient(), refResolver),
		Recorder:      runnerRecorder,
		Clock:         realclock.New(),
		Requeuer:      runnerRequeuer,
		ComputeOps:    computeops.New(mgr.GetClient()),
		MiddlewareOps: mwops.New(mgr.GetClient(), mongoUninstall),
	})
	if err != nil {
		setupLog.Error(err, "Unable to construct app-lifecycle Runner")
		os.Exit(1)
	}
	// Stop drains in-flight CallWorkload goroutines (helm installs,
	// image pulls, namespace waits).
	defer appRunner.Stop()

	if err = ctrladapter.Setup(mgr, appRunner, runnerRequeuer); err != nil {
		setupLog.Error(err, "Unable to register app-lifecycle-runner reconciler")
		os.Exit(1)
	}

	// OpDispatcher wraps the Runner's PendingActionQueue + the same
	// Requeuer registered above; v2 HTTP handlers get this via
	// PrepareRun and use it to deliver TriggerOp actions.
	var opDispatcher lifecycleapi.OpDispatcher = lifecycleapi.NewRunnerDispatcher(appRunner, runnerRequeuer)

	if err = (&controllers.ApplicationManagerGCReconciler{
		Client: mgr.GetClient(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "Unable to create controller", "controller", "Application Manager GC")
		os.Exit(1)
	}

	if err = (&controllers.EntranceStatusManagerController{
		Client: mgr.GetClient(),
	}).SetUpWithManager(mgr); err != nil {
		setupLog.Error(err, "Unable to create controller", "controller", "EntranceStatus Manager")
		os.Exit(1)
	}

	if err = (&controllers.EvictionManagerController{
		Client: mgr.GetClient(),
	}).SetUpWithManager(mgr); err != nil {
		setupLog.Error(err, "Unable to create controller", "controller", "Eviction Manager")
		os.Exit(1)
	}

	if err = (&controllers.PodAbnormalSuspendAppController{
		Client: mgr.GetClient(),
		Ops:    opDispatcher,
	}).SetUpWithManager(mgr); err != nil {
		setupLog.Error(err, "Unable to create controller", "controller", "PodAbnormalSuspendApp")
		os.Exit(1)
	}

	if err = (&controllers.TerminatedPodGCController{
		Client: mgr.GetClient(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "Unable to create controller", "controller", "TerminatedPodGC")
		os.Exit(1)
	}

	if err = (&controllers.TailScaleACLController{
		Client: mgr.GetClient(),
	}).SetUpWithManager(mgr); err != nil {
		setupLog.Error(err, "Unable to create controller", "controller", "tailScaleACL manager")
		os.Exit(1)
	}

	if err = (&controllers.UserController{
		Client:     mgr.GetClient(),
		KubeConfig: config,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "Unable to create controller", "controller", "User")
		os.Exit(1)
	}

	if err = (&controllers.NamespaceReconciler{
		Client: mgr.GetClient(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "Unable to create controller", "controller", "namespace")
		os.Exit(1)
	}

	if err = (&controllers.NodeAlertController{
		Client:     mgr.GetClient(),
		KubeConfig: config,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "Unable to create controller", "controller", "NodeAlert")
		os.Exit(1)
	}

	if err = (&controllers.SystemEnvController{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "Unable to create controller", "controller", "SystemEnv")
		os.Exit(1)
	}

	if err = (&controllers.SystemEnvProcessEnvController{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "Unable to create controller", "controller", "SystemEnvProcessEnv")
		os.Exit(1)
	}

	if err = (&controllers.UserEnvController{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "Unable to create controller", "controller", "UserEnv")
		os.Exit(1)
	}

	if err = (&controllers.UserEnvSyncController{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "Unable to create controller", "controller", "UserEnvSync")
		os.Exit(1)
	}

	if err = (&controllers.AppEnvController{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Ops:    opDispatcher,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "Unable to create controller", "controller", "AppEnv")
		os.Exit(1)
	}

	if err = (&gateway.SharedRouteProducerReconciler{Client: mgr.GetClient()}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "Unable to create controller", "controller", "SharedRouteProducer")
		os.Exit(1)
	}

	if err = (&routecontrol.SharedRouteReconciler{Client: mgr.GetClient()}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "Unable to create controller", "controller", "SharedRoute")
		os.Exit(1)
	}

	if err = (&routecontrol.EntranceTLSReconciler{Client: mgr.GetClient()}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "Unable to create controller", "controller", "EntranceTLS")
		os.Exit(1)
	}

	if err = (&routecontrol.MeshInSharedHostsReconciler{Client: mgr.GetClient()}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "Unable to create controller", "controller", "MeshInSharedHosts")
		os.Exit(1)
	}

	if err = (&routecontrol.EntranceTLSListenerReconciler{Client: mgr.GetClient()}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "Unable to create controller", "controller", "EntranceTLSListener")
		os.Exit(1)
	}

	if err = (&routecontrol.CustomDomainTLSReconciler{Client: mgr.GetClient()}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "Unable to create controller", "controller", "CustomDomainTLS")
		os.Exit(1)
	}

	if err = (&routecontrol.GatewayInClusterIngressNPReconciler{Client: mgr.GetClient()}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "Unable to create controller", "controller", "GatewayInClusterIngressNP")
		os.Exit(1)
	}

	if err = (&callerjwt.IssuerReconciler{Client: mgr.GetClient(), Scheme: mgr.GetScheme()}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "Unable to create controller", "controller", "CallerJWTIssuer")
		os.Exit(1)
	}

	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "Unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "Unable to set up ready check")
		os.Exit(1)
	}

	// sync the api server and the manager with context
	errCh := make(chan error) // api server error
	defer close(errCh)

	c := make(chan os.Signal, 2)
	signal.Notify(c, shutdownSignals...)
	go func() {
		select {
		case <-c:
			cancelFunc()
			<-c
			os.Exit(1) // second signal. Exit directly.
		case err := <-errCh:
			cancelFunc()
			setupLog.Error(err, "Unable to running api server")
			os.Exit(1)
		}
	}()

	// api server run with request's token
	// get kubesphere host from env or config file
	ksHost := os.Getenv(kubeSphereHostAddr)
	ksPort := os.Getenv(kubeSphereHostPort)
	if ksHost == "" || ksPort == "" {
		cancelFunc()
		setupLog.Error(err, "Failed to get the kubesphere api server host from env")
		os.Exit(1)
	}

	// start api server and wait for webhook listener before controllers reconcile
	webhookReady := make(chan struct{})
	func(ctx context.Context, errCh chan error, ksHost string, kubeConfig *rest.Config) {
		go func() {
			if err := runAPIServer(ctx, webhookReady, ksHost, kubeConfig, mgr.GetClient(), opDispatcher); err != nil {
				errCh <- err
			}
		}()
	}(ictx, errCh, fmt.Sprintf("%s:%s", ksHost, ksPort), config)

	select {
	case <-webhookReady:
		setupLog.Info("Webhook server is listening, starting controllers")
	case <-ictx.Done():
		os.Exit(1)
	}

	// Wait until Service Endpoints/EndpointSlice point at this pod before
	// reconciling stop/scale paths that must pass admission webhooks.
	epTimeout, epInterval := webhook.ParseServiceEndpointWaitConfig()
	if err := webhook.WaitForServiceEndpointReady(
		ictx,
		kubeClient,
		"os-framework",
		"app-service",
		webhook.ResolvePodIP(),
		8433,
		epTimeout,
		epInterval,
	); err != nil {
		setupLog.Error(err, "Waiting for webhook service endpoint interrupted")
		os.Exit(1)
	}

	setupLog.Info("Starting manager")
	if err := mgr.Start(ictx); err != nil {
		cancelFunc()
		setupLog.Error(err, "Unable to running manager")
		os.Exit(1)
	}

	cancelFunc()
}

func runAPIServer(ctx context.Context, webhookReady chan struct{}, ksHost string, kubeConfig *rest.Config, client client.Client, opDispatcher lifecycleapi.OpDispatcher) error {
	server, err := apiserver.New(ctx)
	if err != nil {
		return err
	}

	stopCh := make(chan struct{})
	defer close(stopCh)

	err = server.PrepareRun(ksHost, kubeConfig, client, opDispatcher, stopCh)
	if err != nil {
		return err
	}

	err = server.Run(webhookReady)
	return err
}

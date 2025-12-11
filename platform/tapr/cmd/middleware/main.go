package main

import (
	"context"
	"flag"
	"log"

	"net/http"
	_ "net/http/pprof"

	"bytetrade.io/web3os/tapr/cmd/middleware/app"
	kvrocksbakcup "bytetrade.io/web3os/tapr/cmd/middleware/operator/kvrocks-bakcup"
	kvrocksrestore "bytetrade.io/web3os/tapr/cmd/middleware/operator/kvrocks-restore"
	middlewarerequest "bytetrade.io/web3os/tapr/cmd/middleware/operator/middleware-request"
	"bytetrade.io/web3os/tapr/cmd/middleware/operator/pgcluster"
	pgclusterbackup "bytetrade.io/web3os/tapr/cmd/middleware/operator/pgcluster-backup"
	pgclusterrestore "bytetrade.io/web3os/tapr/cmd/middleware/operator/pgcluster-restore"
	"bytetrade.io/web3os/tapr/cmd/middleware/operator/redixcluster"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/klog/v2"

	aprv1 "bytetrade.io/web3os/tapr/pkg/apis/apr/v1alpha1"
	"bytetrade.io/web3os/tapr/pkg/signals"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

func main() {
	if flag.CommandLine.Lookup("add_dir_header") == nil {
		klog.InitFlags(nil)
	}
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	config := ctrl.GetConfigOrDie()
	apiCtx, cancel := context.WithCancel(context.Background())

	// catch signal
	stopCh := signals.SetupSignalHandler(apiCtx, cancel)

	requestController, requestLister := middlewarerequest.NewController(config, apiCtx)
	configMapController, _ := middlewarerequest.NewConfigmapController(config, apiCtx)
	pgclusterController, pgclusterLister := pgcluster.NewController(config, apiCtx, requestLister, func(cluster *aprv1.PGCluster) {
		requestController.PGClusterRecreated(cluster)
	})
	pgBackupController := pgclusterbackup.NewController(config, apiCtx, pgclusterLister)
	pgRestoreController := pgclusterrestore.NewController(config, apiCtx, pgclusterLister)

	redixClusterController, redixLister := redixcluster.NewController(config, apiCtx, func(cluster *aprv1.RedixCluster) {})
	kvrocksBackupController := kvrocksbakcup.NewController(config, apiCtx)
	kvrocksRestoreController := kvrocksrestore.NewController(config, apiCtx)

	runControllers := func() {
		go func() { utilruntime.Must(pgclusterController.Run(1)) }()
		go func() { utilruntime.Must(requestController.Run(1)) }()
		go func() { utilruntime.Must(pgBackupController.Run(1)) }()
		go func() { utilruntime.Must(pgRestoreController.Run(1)) }()
		go func() { utilruntime.Must(redixClusterController.Run(1)) }()
		go func() { utilruntime.Must(kvrocksBackupController.Run(1)) }()
		go func() { utilruntime.Must(kvrocksRestoreController.Run(1)) }()
		go func() { utilruntime.Must(configMapController.Run(1)) }()
		// go func() { backupWatcher.Start() }()
	}

	cmd := &cobra.Command{
		Use:   "middleware-operator",
		Short: "middleware operator server",
		Long:  `The middleware operator server provides the os middleware services`,
		Run: func(cmd *cobra.Command, args []string) {
			runControllers()

			s := &app.Server{
				Ctx:         apiCtx,
				KubeConfig:  config,
				MrLister:    requestLister,
				PgLister:    pgclusterLister,
				RedixLister: redixLister,
			}

			go func() {
				<-stopCh
				s.Shutdown()
			}()

			go func() {
				log.Println(http.ListenAndServe("localhost:6060", nil))
			}()

			s.ServerRun()
			cancel()

			klog.Info("middleware operator shutdown")
		},
	}

	klog.Info("middleware operator starting ... ")

	if err := cmd.Execute(); err != nil {
		klog.Fatalln(err)
	}

}

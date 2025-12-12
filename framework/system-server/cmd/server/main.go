package main

import (
	"context"
	"errors"
	"flag"
	"net/http"

	apiserver "bytetrade.io/web3os/system-server/pkg/apiserver/v1alpha1"
	sysclientset "bytetrade.io/web3os/system-server/pkg/generated/clientset/versioned"
	informers "bytetrade.io/web3os/system-server/pkg/generated/informers/externalversions"
	"bytetrade.io/web3os/system-server/pkg/generated/listers/sys/v1alpha1"
	prodiverregistry "bytetrade.io/web3os/system-server/pkg/providerregistry/v1alpha1"
	"bytetrade.io/web3os/system-server/pkg/signals"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
)

func main() {
	klog.InitFlags(nil)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	config := ctrl.GetConfigOrDie()

	ctx := context.Background()
	apiCtx, cancel := context.WithCancel(ctx)

	stopCh := signals.SetupSignalHandler(apiCtx, cancel)

	sysClient := sysclientset.NewForConfigOrDie(config)

	informerFactory := informers.NewSharedInformerFactory(sysClient, 0)
	providerInformer := informerFactory.Sys().V1alpha1().ProviderRegistries()
	informerFactory.Sys().V1alpha1().ApplicationPermissions().Informer()
	permissionInformer := informerFactory.Sys().V1alpha1().ApplicationPermissions()

	controller := prodiverregistry.NewController(sysClient, providerInformer)

	cmd := &cobra.Command{
		Use:   "system-server",
		Short: "system server",
		Long:  `The system server provides underlayer IPC and event messages flow`,
		Run: func(cmd *cobra.Command, args []string) {
			go func() {
				defer cancel()
				if err := APIRun(apiCtx, config, sysClient,
					permissionInformer.Lister(), providerInformer.Lister()); err != nil {
					panic(err)
				}
			}()

			defer func() {
				informerFactory.Shutdown()
				cancel()
			}()
			informerFactory.Start(stopCh)

			if err := controller.Run(1, stopCh); err != nil {
				panic(err)
			}
		},
	}

	klog.Info("system-server starting ... ")

	if err := cmd.Execute(); err != nil {
		klog.Fatalln(err)
	}
}

// APIRun is responsible for running the API server.
func APIRun(ctx context.Context, kubeconfig *rest.Config, sysclientset *sysclientset.Clientset,
	permissionLister v1alpha1.ApplicationPermissionLister, providerLister v1alpha1.ProviderRegistryLister,
) error {
	server, err := apiserver.New(ctx)
	if err != nil {
		return err
	}

	err = server.PrepareRun(kubeconfig, sysclientset, permissionLister, providerLister)
	if err != nil {
		return err
	}

	err = server.Run()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

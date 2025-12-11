package main

import (
	"context"

	"bytetrade.io/web3os/tapr/cmd/sys-event/apiserver"
	"bytetrade.io/web3os/tapr/cmd/sys-event/watchers"
	"bytetrade.io/web3os/tapr/cmd/sys-event/watchers/apps"
	"bytetrade.io/web3os/tapr/cmd/sys-event/watchers/dnspod"
	"bytetrade.io/web3os/tapr/cmd/sys-event/watchers/metrics"
	"bytetrade.io/web3os/tapr/cmd/sys-event/watchers/users"
	"bytetrade.io/web3os/tapr/cmd/sys-event/watchers/workflows"
	"bytetrade.io/web3os/tapr/pkg/app/application"
	"bytetrade.io/web3os/tapr/pkg/signals"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	_ = signals.SetupSignalHandler(ctx, cancel)

	config := ctrl.GetConfigOrDie()

	notification := watchers.Notification{
		DynamicClient: dynamic.NewForConfigOrDie(config),
	}
	w := watchers.NewWatchers(ctx, config, 0)

	// add event subscriber to watchers
	watchers.AddToWatchers[application.Application](w, application.GVR,
		(&apps.Subscriber{Subscriber: watchers.NewSubscriber(w).WithNotification(&notification)}).HandleEvent())
	watchers.AddToWatchers[corev1.Namespace](w, corev1.SchemeGroupVersion.WithResource("namespaces"),
		(&workflows.Subscriber{Subscriber: watchers.NewSubscriber(w).WithNotification(&notification)}).WithKubeConfig(config).HandleEvent())
	watchers.AddToWatchers[corev1.Pod](w, corev1.SchemeGroupVersion.WithResource("pods"),
		(&dnspod.PodSubscriber{Subscriber: watchers.NewSubscriber(w).WithNotification(&notification)}).WithKubeConfig(config).HandleEvent())
	watchers.AddToWatchers[corev1.Node](w, corev1.SchemeGroupVersion.WithResource("nodes"),
		(&dnspod.NodeSubscriber{Subscriber: watchers.NewSubscriber(w).WithNotification(&notification)}).WithKubeConfig(config).HandleEvent())

	go w.Run(1)

	// metrics monitoring
	go metrics.NewMetricsWatcherOrDie(ctx, w, &notification, config).Run()

	// api server
	api := apiserver.NewServer(w, &notification)
	go api.Run()

	userWatcher := users.NewWatcher(ctx, config, w, &notification)
	userWatcher.Start()

	<-ctx.Done()
	api.ShutDown()
	klog.Info("shutdown sys-event manager")
}
